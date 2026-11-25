package websocket

// tracks chatrooms during runtime state (when active)

type ChatroomRegistry struct {
	// Pointer to DB

}

func (chatroomRegistry *ChatroomRegistry) RegisterChatroom() {
	/*
		Create active chatroom entry
		Prevent duplicate registration
	*/
}

func (chatroomRegistry *ChatroomRegistry) GetActiveChatroom() {
	/*
		Find active chatroom by ID
	*/
}

func (chatroomRegistry *ChatroomRegistry) RemoveActiveChatroom() {
	/*
		Remove chatroom from registry
	*/
}

func (chatroomRegistry *ChatroomRegistry) ListActiveChatrooms() {
	/*
		Return chatrooms currently managed
	*/
}

func (chatroomRegistry *ChatroomRegistry) GetActiveChatroomCount() {
	/*
		Return # of active chatrooms
	*/
}
