package model

type ChatRoom struct {
	ID       int64  `db:"id"`
	RoomName string `db:"room_name"`
}

type RoomParticipant struct {
	ID     int64 `db:"id"`
	RoomID int64 `db:"room_id"`
	UserID int64 `db:"user_id"`
}
