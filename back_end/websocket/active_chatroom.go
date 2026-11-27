package websocket

import "sync"

// chatroom currently being managed in memory

type ActiveChatroom struct {
	ChatroomID       int64
	ConnectedClients map[*Client]struct{}
	mu               sync.RWMutex
}

func NewActiveChatroom(chatroomID int64) *ActiveChatroom {
	return &ActiveChatroom{ChatroomID: chatroomID, ConnectedClients: make(map[*Client]struct{})}
}

func (activeChatroom *ActiveChatroom) AddClient(client *Client) {
	activeChatroom.mu.Lock()
	defer activeChatroom.mu.Unlock()

	activeChatroom.ConnectedClients[client] = struct{}{}
}

func (activeChatroom *ActiveChatroom) RemoveClient(client *Client) {
	activeChatroom.mu.Lock()
	defer activeChatroom.mu.Unlock()

	delete(activeChatroom.ConnectedClients, client)
}

func (activeChatroom *ActiveChatroom) BroadcastMessage(message []byte) {
	activeChatroom.mu.RLock()

	var slowClients []*Client

	for client := range activeChatroom.ConnectedClients {
		if !client.QueueMessage(message) {
			slowClients = append(slowClients, client)
		}
	}

	activeChatroom.mu.RUnlock()

	for _, client := range slowClients {
		client.CloseConnection()
	}
}

func (activeChatroom *ActiveChatroom) GetOnlineUsers() []int64 {
	activeChatroom.mu.RLock()
	defer activeChatroom.mu.RUnlock()

	uniqueUsers := make(map[int64]struct{})

	for client := range activeChatroom.ConnectedClients {
		uniqueUsers[client.UserID] = struct{}{}
	}

	users := make([]int64, 0, len(uniqueUsers))

	for userID := range uniqueUsers {
		users = append(users, userID)
	}

	return users
}

func (activeChatroom *ActiveChatroom) GetConnectionCount() int {
	activeChatroom.mu.RLock()
	defer activeChatroom.mu.RUnlock()

	return len(activeChatroom.ConnectedClients)
}

func (activeChatroom *ActiveChatroom) IsEmpty() bool {
	activeChatroom.mu.RLock()
	defer activeChatroom.mu.RUnlock()

	return len(activeChatroom.ConnectedClients) == 0
}
