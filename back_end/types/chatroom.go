package types

// Chatroom metadata

type Chatroom struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	OwnerID     int64    `json:"owner_id"`
	CreatedAt   Datetime `json:"created_at"`
	Description string   `json:"description"`
}
