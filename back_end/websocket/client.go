package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID        int64
	ChatroomID    int64
	OutboundQueue chan []byte
	Connection    *websocket.Conn
	Chatroom      *ActiveChatroom
	Manager       *WebSocketManager
	closeOnce     sync.Once
	Done          chan struct{}
}

// Doesnt enforce msg limit or content validation rn, need to fix
func (client *Client) ReadMessages() {
	defer client.CloseConnection()
	client.Connection.SetReadLimit(8192)
	for {
		messageType, message, err := client.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				log.Printf("websocket read err for user %d: %v", client.UserID, err)
			}
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		if len(message) == 0 {
			continue
		}

		if err := client.Manager.HandleIncomingEvent(client, message); err != nil {
			log.Printf(
				"invalid event from user %d: %v",
				client.UserID,
				err,
			)
		}
	}
}

func (client *Client) WriteMessages() {
	defer client.CloseConnection()

	for {
		select {
		case <-client.Done:
			return
		case message, ok := <-client.OutboundQueue:
			if !ok {
				return
			}

			err := client.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err != nil {
				log.Printf("failed to set write deadline for user %d : %v", client.UserID, err)
				return
			}

			err = client.Connection.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("websocket write error for user %d : %v", client.UserID, err)
				return
			}
		}
	}
}

func (client *Client) CloseConnection() {
	client.closeOnce.Do(func() {
		close(client.Done)
		if client.Chatroom != nil {
			client.Chatroom.RemoveClient(client)
		}

		if client.Connection != nil {
			if err := client.Connection.Close(); err != nil {
				log.Printf("error closing connection for user %d : %v", client.UserID, err)
			}
		}

		log.Printf("user %d disconnected from chatroom %d", client.UserID, client.ChatroomID)
	})
}

// select makes statement non-blocking for the channel
func (client *Client) QueueMessage(message []byte) bool {
	select {
	case <-client.Done:
		return false
	case client.OutboundQueue <- message:
		return true
	default:
		return false
	}
}
