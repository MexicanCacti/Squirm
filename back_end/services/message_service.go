package services

// Note: Save msg BEFORE send

// Process messages indep of websocket conn lifecycle

type MessageService struct {
}

func (messageService *MessageService) ProcessMessage() {
	/*
		Validate message content
		Verify sender belongs to chatroom
		associate message w/ auth'd user's Identity
		Persist message in RDS
		Request message be broadcast to chatroom
	*/
}

func (messageService *MessageService) GetMessageHistory() {
	/*
		Get previously sent messages for chatroom
		Pagination to avoid returning entire chat history
	*/
}

func (messageService *MessageService) ValidateMessage() {
	/*
		Check msg not empty
		Enforce max msg size
		Reject invalid msg content
	*/
}

func (messageService *MessageService) BroadcastMessage() {
	/*
		Locate ActiveChatroom
		REquest message distrib to connected clients
	*/
}
