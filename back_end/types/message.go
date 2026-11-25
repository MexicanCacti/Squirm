package types

import "sync"

// Message metadata, also defines Thread-Safe MessageQueue

type Message struct {
	MessageID   int64    `json:"message_id"`
	ChatroomID  int64    `json:"chatroom_id"`
	SenderID    int64    `json:"sender_id"`
	Contents    string   `json:"contents"`
	TimeCreated Datetime `json:"time_created"`
}

type MessageQueue struct {
	Lock   *sync.Mutex
	Values []Message
}

func Init() MessageQueue {
	return MessageQueue{&sync.Mutex{}, make([]Message, 0)}
}

func (msg_q *MessageQueue) Enqueue(msg Message) {
	for {
		msg_q.Lock.Lock()
		msg_q.Values = append(msg_q.Values, msg)
		msg_q.Lock.Unlock()
		return
	}
}

func (msg_q *MessageQueue) Dequeue() *Message {
	for {
		if len(msg_q.Values) > 0 {
			msg_q.Lock.Lock()
			msg := msg_q.Values[0]
			msg_q.Values = msg_q.Values[1:]
			msg_q.Lock.Unlock()
			return &msg
		}
		return nil
	}
	return nil
}
