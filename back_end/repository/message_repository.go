package repository

type MessageRepository struct {
}

func (messageRepository *MessageRepository) SaveMessage() {
	/*
		persist message
	*/
}

func (messageRepository *MessageRepository) GetMessagesFromChatroom() {
	/*
		retirive paginated message history
	*/
}

func (messageRepository *MessageRepository) DeleteMessagesFromChatroom() {
	/*
		delete message history w/ pagination
	*/
}
