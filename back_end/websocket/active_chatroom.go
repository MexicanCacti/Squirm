package websocket

// chatroom currently being managed in memory

type ActiveChatroom struct {
	ChatroomID       int64
	ConnectedClients []Client
	// Mutex for sync
}

func (activeChatroom *ActiveChatroom) AddClient() {
	/*
		Register new Websocket connection w/ chatroom
		Update chatroom's connectedClients
	*/
}

func (activeChatroom *ActiveChatroom) RemoveClient() {
	/*
		Remove client connection
		Update connectedClients
	*/
}

func (activeChatroom *ActiveChatroom) BroadcastMessage() {
	/*
		Distribute message to all Currently connected clients
	*/
}

func (activeChatroom *ActiveChatroom) GetOnlineUsers() {
	/*
		Return DISTINCT users connceted to chatroom
	*/
}

func (activeChatroom *ActiveChatroom) GetConnectionCount() {
	/*
		Return # of active websocket connections
	*/
}

func (activeChatroom *ActiveChatroom) IsEmpty() {
	/*
		true if chatroom has no active connections
	*/
}
