package websocket

import "sync"

// tracks chatrooms during runtime state (when active)

type ChatroomRegistry struct {
	activeChatrooms map[int64]*ActiveChatroom
	mu              sync.RWMutex
}

func NewChatroomRegistry() *ChatroomRegistry {
	return &ChatroomRegistry{activeChatrooms: make(map[int64]*ActiveChatroom)}
}

func (chatroomRegistry *ChatroomRegistry) RegisterChatroom(chatroomID int64) *ActiveChatroom {
	chatroomRegistry.mu.Lock()
	defer chatroomRegistry.mu.Unlock()

	if chatroom, exists := chatroomRegistry.activeChatrooms[chatroomID]; exists {
		return chatroom
	}

	chatroom := NewActiveChatroom(chatroomID)
	chatroomRegistry.activeChatrooms[chatroomID] = chatroom
	return chatroom
}

func (chatroomRegistry *ChatroomRegistry) GetActiveChatroom(chatroomID int64) (*ActiveChatroom, bool) {
	chatroomRegistry.mu.RLock()
	defer chatroomRegistry.mu.RUnlock()

	chatroom, exists := chatroomRegistry.activeChatrooms[chatroomID]
	return chatroom, exists
}

func (chatroomRegistry *ChatroomRegistry) RemoveActiveChatroom(chatroomID int64) bool {
	chatroomRegistry.mu.Lock()
	defer chatroomRegistry.mu.Unlock()

	chatroom, exists := chatroomRegistry.activeChatrooms[chatroomID]

	if !exists {
		return false
	}

	if !chatroom.IsEmpty() {
		return false
	}

	delete(chatroomRegistry.activeChatrooms, chatroomID)

	return true
}

func (chatroomRegistry *ChatroomRegistry) ListActiveChatrooms() []*ActiveChatroom {
	chatroomRegistry.mu.RLock()
	defer chatroomRegistry.mu.RUnlock()

	chatrooms := make([]*ActiveChatroom, 0, len(chatroomRegistry.activeChatrooms))

	for _, chatroom := range chatroomRegistry.activeChatrooms {
		chatrooms = append(chatrooms, chatroom)
	}

	return chatrooms
}

func (chatroomRegistry *ChatroomRegistry) GetActiveChatroomCount() int {
	chatroomRegistry.mu.RLock()
	defer chatroomRegistry.mu.RUnlock()

	return len(chatroomRegistry.activeChatrooms)
}
