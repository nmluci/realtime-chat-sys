package model

type ChatHistory struct {
	ID            int64  `db:"id"`
	RoomID        int64  `db:"room_id"`
	SenderID      int64  `db:"sender_id"`
	SenderName    string `db:"sender_name"`
	RecipientID   int64  `db:"recipient_id"`
	RecipientName string `db:"recipient_name"`
	Message       string `db:"message"`
}
