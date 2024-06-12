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
