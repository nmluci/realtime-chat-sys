package indto

type ChatHistoryParams struct {
	ID       int64
	RoomID   int64
	RoomName string
	UserID   int64
	IsDM     bool
}
