package dto

type LiveChatSocketRequest struct {
}

type LiveChatSocketEvent struct {
	EventName string      `json:"event"`
	Data      interface{} `json:"data"`
}
