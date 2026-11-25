package types

type Client struct {
	UserID          int64        `json:"user_id"`
	ChatroomID      int64        `json:"chatroom_id"`
	OutMessageQueue MessageQueue `json:"out_message_queue"`
}
