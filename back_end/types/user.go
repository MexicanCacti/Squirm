package types

// User metadata

type User struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	TimeCreated Datetime `json:"time_created"`
}
