package websocket

// Note: Users can have multiple active connections (think two browser tabs)

type Client struct {
	UserID        int64
	ChatroomID    int64
	OutboundQueue MessageQueue
	// Websocket
}

func (client *Client) ReadMessages() {
	/*
		Continuously read incoming websocket messages
		Validate message format
		Forward valid messages to appropriate message handler
		Detect connection failures
	*/
}

func (client *Client) WriteMessages() {
	/*
		Conitnuously process messages from the outbound queue
		Write messages to the Websocket
		Handle write failures
	*/
}

func (client *Client) CloseConnection() {
	/*
		Close Websocket connection
		Realease connection-related resources
		Notify appriopriate component of client disconnect
	*/
}

func (client *Client) QueueMessage() {
	/*
		Add outogoing message to the client's outbound queue
		Handle queue full
	*/
}
