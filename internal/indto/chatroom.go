package indto

type ChatRoomParams struct {
	ID       int64
	RoomName string
	IsDM     bool
	UserID   int64
}

type RoomParticipantParams struct {
	ID     int64
	RoomID int64
	UserID int64
}

type IncomingMessage struct {
	SenderID    int64  `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	RecipientID int64  `json:"recipient_id"`
	Content     string `json:"content"`
	IsDM        bool   `json:"is_dm"`
}
