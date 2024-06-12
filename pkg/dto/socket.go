package dto

type LiveChatSocketRequest struct {
	SenderID    int64
	RecipientID int64
	Event       LiveChatSocketEvent
}

type LiveChatSocketEvent struct {
	EventName string      `json:"event"`
	Data      interface{} `json:"data"`
}

type LiveChatBroadcastEvent struct {
	Room  int64
	Event LiveChatSocketEvent
}
