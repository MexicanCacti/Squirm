package tests

import (
	"Squirm/back_end/handlers"
	"Squirm/back_end/repository"
	"Squirm/back_end/router"
	"Squirm/back_end/services"
	wsapp "Squirm/back_end/websocket"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const wsTestTimeout = 4 * time.Second

type wsFixture struct {
	router   *gin.Engine
	db       *sql.DB
	registry *wsapp.ChatroomRegistry
	manager  *wsapp.WebSocketManager
	server   *httptest.Server
}

func newWSFixture(t *testing.T) *wsFixture {
	t.Helper()
	db := setupTestDB(t)
	repo := repository.NewChatroomRepository(db)
	service := services.NewChatroomService(repo)
	chatHandler := handlers.NewChatroomHandler(service)
	registry := wsapp.NewChatroomRegistry()
	manager := wsapp.NewWebSocketManager(registry, service)
	wsHandler := handlers.NewWebSocketHandler(manager)
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	auth := func(c *gin.Context) { c.Set("userID", int64(1)); c.Next() }
	router.RegisterChatroomRoutes(engine, chatHandler, wsHandler, auth)
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)
	return &wsFixture{engine, db, registry, manager, server}
}

func (f *wsFixture) url(roomID int64) string {
	return "ws" + strings.TrimPrefix(f.server.URL, "http") + fmt.Sprintf("/chatrooms/%d/ws", roomID)
}

func dialWS(t *testing.T, f *wsFixture, roomID int64) *websocket.Conn {
	t.Helper()
	dialer := websocket.Dialer{HandshakeTimeout: wsTestTimeout}
	conn, resp, err := dialer.Dial(f.url(roomID), nil)
	if err != nil {
		if resp != nil {
			t.Fatalf("dial room %d: HTTP %d: %v", roomID, resp.StatusCode, err)
		}
		t.Fatalf("dial room %d: %v", roomID, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

type wsEvent struct {
	Type       string `json:"type"`
	Content    string `json:"content"`
	UserID     int64  `json:"user_id"`
	ChatroomID int64  `json:"chatroom_id"`
}

func sendChat(t *testing.T, conn *websocket.Conn, content string) {
	t.Helper()
	if err := conn.SetWriteDeadline(time.Now().Add(wsTestTimeout)); err != nil {
		t.Fatal(err)
	}
	if err := conn.WriteJSON(wsEvent{Type: "chat_message", Content: content}); err != nil {
		t.Fatal(err)
	}
}

func readChat(t *testing.T, conn *websocket.Conn, content string, roomID int64) {
	t.Helper()
	deadline := time.Now().Add(wsTestTimeout)
	for time.Now().Before(deadline) {
		if err := conn.SetReadDeadline(deadline); err != nil {
			t.Fatal(err)
		}
		var event wsEvent
		if err := conn.ReadJSON(&event); err != nil {
			t.Fatalf("read chat %q: %v", content, err)
		}
		if event.Type != "chat_message" {
			continue
		} // Ignore presence events.
		if event.Content != content || event.ChatroomID != roomID || event.UserID != 1 {
			t.Fatalf("unexpected chat event: %+v; wanted content=%q room=%d user=1", event, content, roomID)
		}
		return
	}
	t.Fatalf("timed out waiting for chat %q", content)
}

func waitForConnections(t *testing.T, registry *wsapp.ChatroomRegistry, roomID int64, count int) {
	t.Helper()
	deadline := time.Now().Add(wsTestTimeout)
	for time.Now().Before(deadline) {
		room, exists := registry.GetActiveChatroom(roomID)
		if count == 0 && (!exists || room.GetConnectionCount() == 0) {
			return
		}
		if exists && room.GetConnectionCount() == count {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	room, exists := registry.GetActiveChatroom(roomID)
	if !exists {
		t.Fatalf("room %d not active; expected %d connections", roomID, count)
	}
	t.Fatalf("room %d: expected %d connections, got %d", roomID, count, room.GetConnectionCount())
}

func TestWebSocketSendReceive(t *testing.T) {
	f := newWSFixture(t)
	id := createTestChatroom(t, f.router, f.db)
	conn := dialWS(t, f, id)
	waitForConnections(t, f.registry, id, 1)
	sendChat(t, conn, "hello")
	readChat(t, conn, "hello", id)
}

func TestWebSocketBroadcast(t *testing.T) {
	f := newWSFixture(t)
	id := createTestChatroom(t, f.router, f.db)
	a, b := dialWS(t, f, id), dialWS(t, f, id)
	waitForConnections(t, f.registry, id, 2)
	sendChat(t, a, "broadcast")
	readChat(t, a, "broadcast", id)
	readChat(t, b, "broadcast", id)
}

func TestWebSocketChatroomIsolation(t *testing.T) {
	f := newWSFixture(t)
	roomA := createTestChatroom(t, f.router, f.db)
	roomB := createTestChatroom(t, f.router, f.db)
	a, b := dialWS(t, f, roomA), dialWS(t, f, roomB)
	waitForConnections(t, f.registry, roomA, 1)
	waitForConnections(t, f.registry, roomB, 1)

	// A dedicated reader keeps B's connection usable; a read timeout on the
	// WebSocket itself would corrupt subsequent reads.
	receivedB := make(chan wsEvent, 16)
	readErrB := make(chan error, 1)
	go func() {
		for {
			var event wsEvent
			if err := b.ReadJSON(&event); err != nil {
				readErrB <- err
				return
			}
			if event.Type == "chat_message" {
				receivedB <- event
			}
		}
	}()
	sendChat(t, a, "only-A")
	readChat(t, a, "only-A", roomA)
	select {
	case ev := <-receivedB:
		t.Fatalf("cross-room message leaked: %+v", ev)
	case err := <-readErrB:
		t.Fatalf("room B reader failed: %v", err)
	case <-time.After(200 * time.Millisecond):
	}
	sendChat(t, b, "only-B")
	select {
	case ev := <-receivedB:
		if ev.Content != "only-B" || ev.ChatroomID != roomB {
			t.Fatalf("unexpected room B event: %+v", ev)
		}
	case err := <-readErrB:
		t.Fatalf("room B reader failed: %v", err)
	case <-time.After(wsTestTimeout):
		t.Fatal("room B stopped receiving messages")
	}
}

func TestWebSocketNonexistentChatroom(t *testing.T) {
	f := newWSFixture(t)
	dialer := websocket.Dialer{HandshakeTimeout: wsTestTimeout}
	conn, resp, err := dialer.Dial(f.url(999999999), nil)
	if conn != nil {
		_ = conn.Close()
		t.Fatal("unexpected successful connection")
	}
	if err == nil {
		t.Fatal("expected handshake failure")
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected HTTP 404, got response=%v error=%v", resp, err)
	}
	_ = resp.Body.Close()
}

func TestWebSocketInvalidEvent(t *testing.T) {
	f := newWSFixture(t)
	id := createTestChatroom(t, f.router, f.db)
	conn := dialWS(t, f, id)
	if err := conn.WriteJSON(wsEvent{Type: "unknown_event", Content: "ignored"}); err != nil {
		t.Fatal(err)
	}
	sendChat(t, conn, "still-working")
	readChat(t, conn, "still-working", id)
}

func TestWebSocketMalformedJSON(t *testing.T) {
	f := newWSFixture(t)
	id := createTestChatroom(t, f.router, f.db)
	conn := dialWS(t, f, id)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type":`)); err != nil {
		t.Fatal(err)
	}
	sendChat(t, conn, "still-working")
	readChat(t, conn, "still-working", id)
}

func TestWebSocketDisconnect(t *testing.T) {
	f := newWSFixture(t)
	id := createTestChatroom(t, f.router, f.db)
	conn := dialWS(t, f, id)
	waitForConnections(t, f.registry, id, 1)
	_ = conn.Close()
	waitForConnections(t, f.registry, id, 0)
}

func TestWebSocketMultipleConnections(t *testing.T) {
	f := newWSFixture(t)
	id := createTestChatroom(t, f.router, f.db)
	a, b := dialWS(t, f, id), dialWS(t, f, id)
	waitForConnections(t, f.registry, id, 2)
	room, ok := f.registry.GetActiveChatroom(id)
	if !ok {
		t.Fatal("room not registered")
	}
	users := room.GetOnlineUsers()
	if len(users) != 1 || users[0] != 1 {
		t.Fatalf("expected one distinct user 1, got %v", users)
	}
	_ = a.Close()
	waitForConnections(t, f.registry, id, 1)
	sendChat(t, b, "second-tab")
	readChat(t, b, "second-tab", id)
}

func TestWebSocketConcurrentBroadcast(t *testing.T) {
	f := newWSFixture(t)
	id := createTestChatroom(t, f.router, f.db)
	const clients = 5
	const messagesPerClient = 10
	connections := make([]*websocket.Conn, clients)
	for i := range connections {
		connections[i] = dialWS(t, f, id)
	}
	waitForConnections(t, f.registry, id, clients)

	// Each connection has one writer and one reader, as required by Gorilla.
	errorsCh := make(chan error, clients*2)
	var wg sync.WaitGroup
	for i, conn := range connections {
		wg.Add(1)
		go func(index int, c *websocket.Conn) {
			defer wg.Done()
			for j := 0; j < messagesPerClient; j++ {
				if err := c.WriteJSON(wsEvent{Type: "chat_message", Content: fmt.Sprintf("%d-%d", index, j)}); err != nil {
					errorsCh <- err
					return
				}
			}
		}(i, conn)
		wg.Add(1)
		go func(c *websocket.Conn) {
			defer wg.Done()
			_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
			seen := make(map[string]bool)
			for len(seen) < clients*messagesPerClient {
				var ev wsEvent
				if err := c.ReadJSON(&ev); err != nil {
					errorsCh <- err
					return
				}
				if ev.Type != "chat_message" {
					continue
				}
				if ev.ChatroomID != id || ev.UserID != 1 {
					errorsCh <- fmt.Errorf("unexpected event %+v", ev)
					return
				}
				if seen[ev.Content] {
					errorsCh <- fmt.Errorf("duplicate message %q", ev.Content)
					return
				}
				seen[ev.Content] = true
			}
		}(conn)
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		t.Errorf("concurrent broadcast: %v", err)
	}
}

func TestWebSocketShutdown(t *testing.T) {
	f := newWSFixture(t)
	id := createTestChatroom(t, f.router, f.db)
	conn := dialWS(t, f, id)
	waitForConnections(t, f.registry, id, 1)
	ctx, cancel := context.WithTimeout(context.Background(), wsTestTimeout)
	defer cancel()
	if err := f.manager.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(wsTestTimeout)); err != nil {
		t.Fatal(err)
	}
	var event json.RawMessage
	if err := conn.ReadJSON(&event); err == nil {
		t.Fatal("expected connection to close on shutdown")
	}
	waitForConnections(t, f.registry, id, 0)
	dialer := websocket.Dialer{HandshakeTimeout: wsTestTimeout}
	newConn, resp, err := dialer.Dial(f.url(id), nil)
	if newConn != nil {
		_ = newConn.Close()
		t.Fatal("connection accepted after shutdown")
	}
	if err == nil {
		t.Fatal("expected post-shutdown dial to fail")
	}
	if resp != nil {
		_ = resp.Body.Close()
	}
}
