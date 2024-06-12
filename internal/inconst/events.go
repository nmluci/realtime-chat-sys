package inconst

const (
	LiveChatBaseEvent        = "livechat:"
	LiveChatAuthSignupEvent  = LiveChatBaseEvent + "auth:signup"
	LiveChatAuthLoginEvent   = LiveChatBaseEvent + "auth:login"
	LiveChatAuthAckEvent     = LiveChatBaseEvent + "auth:ack"
	LiveChatCreateRoomEvent  = LiveChatBaseEvent + "chat:created_room"
	LiveChatCreatedEvent     = LiveChatBaseEvent + "chat:created"
	LiveChatJoinRoomEvent    = LiveChatBaseEvent + "chat:join_room"
	LiveChatJoinedEvent      = LiveChatBaseEvent + "chat:joined"
	LiveChatLeaveRoomEvent   = LiveChatBaseEvent + "chat:leave_room"
	LiveChatLeftEvent        = LiveChatBaseEvent + "chat:left"
	LiveChatIncomingMsgEvent = LiveChatBaseEvent + "msg:incoming"
	LiveChatSendMsgEvent     = LiveChatBaseEvent + "msg:send"
	LiveChatErrorMsgEvent    = LiveChatBaseEvent + "error"
)
