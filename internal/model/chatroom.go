package model

type ChatRoom struct {
	ID       int64  `db:"id"`
	RoomName string `db:"room_name"`
	IsDM     bool   `db:"is_dm"`
	IsJoined bool   `db:"is_joined"`
}

type RoomParticipant struct {
	ID     int64 `db:"id"`
	RoomID int64 `db:"room_id"`
	UserID int64 `db:"user_id"`
}
