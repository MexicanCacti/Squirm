package websocket

import (
	"Squirm/back_end/services"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type IncomingEvent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type OutgoingEvent struct {
	Type       string `json:"type"`
	UserID     int64  `json:"user_id,omitempty"`
	ChatroomID int64  `json:"chatroom_id,omitempty"`
	Content    string `json:"content,omitempty"`
}

type WebSocketManager struct {
	chatroomRegistry *ChatroomRegistry
	chatroomService  *services.ChatroomService

	websocketUpgrader websocket.Upgrader
	mu                sync.RWMutex
	shuttingDown      bool
}

func NewWebSocketManager(chatroomRegistry *ChatroomRegistry, chatroomService *services.ChatroomService) *WebSocketManager {
	return &WebSocketManager{chatroomRegistry: chatroomRegistry, chatroomService: chatroomService,
		websocketUpgrader: websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return r.Header.Get("Origin") == "" || r.Header.Get("Origin") == "http://localhost:8080"
			}},
	}
}

func (webSocketManager *WebSocketManager) HandleConnection(w http.ResponseWriter, r *http.Request, chatroomID int64, userID int64) error {
	if chatroomID <= 0 || userID <= 0 {
		return errors.New("invalid chatroom/user ID")
	}

	_, err := webSocketManager.chatroomService.GetChatroom(r.Context(), chatroomID)
	if err != nil {
		return err
	}

	// ideally check access rules but w/e

	webSocketManager.mu.Lock()
	if webSocketManager.shuttingDown {
		webSocketManager.mu.Unlock()
		return errors.New("websocket manager shutting down")
	}

	conn, err := webSocketManager.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		webSocketManager.mu.Unlock()
		return err
	}

	room := webSocketManager.chatroomRegistry.RegisterChatroom(chatroomID)

	client := &Client{
		UserID:        userID,
		ChatroomID:    chatroomID,
		Connection:    conn,
		Chatroom:      room,
		Manager:       webSocketManager,
		OutboundQueue: make(chan []byte, 256),
		Done:          make(chan struct{}),
	}
	room.AddClient(client)
	webSocketManager.mu.Unlock()

	go client.WriteMessages()
	client.ReadMessages()
	webSocketManager.HandleDisconnect(client)
	return nil
}

func (webSocketManager *WebSocketManager) HandleDisconnect(client *Client) {
	if client == nil {
		return
	}

	room := client.Chatroom

	client.CloseConnection()

	if room == nil {
		return
	}

	event := OutgoingEvent{Type: "user_left", UserID: client.UserID, ChatroomID: client.ChatroomID}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal disconnect :%v", err)
		return
	}

	room.BroadcastMessage(payload)
}

func (webSocketManager *WebSocketManager) HandleIncomingEvent(client *Client, message []byte) error {
	var event IncomingEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return errors.New("invalid JSON event")
	}

	switch event.Type {
	case "chat_message":
		if len(event.Content) == 0 {
			return errors.New("message can't be empty")
		}
		if len(event.Content) > 4096 {
			return errors.New("message exceeds max size")
		}

		outgoing := OutgoingEvent{
			Type:       "chat_message",
			UserID:     client.UserID,
			ChatroomID: client.ChatroomID,
			Content:    event.Content,
		}

		payload, err := json.Marshal(outgoing)
		if err != nil {
			return err
		}
		client.Chatroom.BroadcastMessage(payload)

		return nil

	default:
		return errors.New("unsupported event type")
	}
}

func (webSocketManager *WebSocketManager) Shutdown(ctx context.Context) error {
	webSocketManager.mu.Lock()

	if webSocketManager.shuttingDown {
		webSocketManager.mu.Unlock()
		return nil
	}

	webSocketManager.shuttingDown = true
	rooms := webSocketManager.chatroomRegistry.ListActiveChatrooms()
	var clients []*Client

	for _, room := range rooms {
		room.mu.RLock()
		for client := range room.ConnectedClients {
			clients = append(clients, client)
		}

		room.mu.RUnlock()
	}

	webSocketManager.mu.Unlock()

	for _, client := range clients {
		client.CloseConnection()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
