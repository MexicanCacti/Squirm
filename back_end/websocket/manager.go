package websocket

type WebSocketManager struct {
}

func (webSocketManager *WebSocketManager) HandleConnection() {
	/*
		Authenticate conn request
		Verify access to chatroom
		Upgrade HTTP conn to websocket
		Create client instance
		Register client w/ ActiveChatroom
		Start client r/w
	*/
}

func (webSocketManager *WebSocketManager) HandleDisconnect() {
	/*
		Remove client from chatroom
		Release connection resources
		Notify remaining users about disconnect
	*/
}

func (webSocketManager *WebSocketManager) HandleIncomingEvent() {
	/*
		Determine event type
		Route chat messages to MessageService
		Reject unsupported event types
	*/
}

func (webSocketManager *WebSocketManager) Shutdown() {
	/*
		Stop accepting new connections
		Close existing websocket connections
		RElease resoureces during shutdown
	*/
}
