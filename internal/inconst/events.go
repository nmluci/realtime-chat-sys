package inconst

const (
	LiveChatBaseEvent          = "livechat:"
	LiveChatAuthSignupEvent    = LiveChatBaseEvent + "auth:signup"
	LiveChatAuthLoginEvent     = LiveChatBaseEvent + "auth:login"
	LiveChatAuthAckEvent       = LiveChatBaseEvent + "auth:ack"
	LiveChatCreateRoomEvent    = LiveChatBaseEvent + "chat:create_room"
	LiveChatCreatedEvent       = LiveChatBaseEvent + "chat:created"
	LiveChatJoinRoomEvent      = LiveChatBaseEvent + "chat:join_room"
	LiveChatJoinedEvent        = LiveChatBaseEvent + "chat:joined"
	LiveChatLeaveRoomEvent     = LiveChatBaseEvent + "chat:leave_room"
	LiveChatLeftEvent          = LiveChatBaseEvent + "chat:left"
	LiveChatIncomingMsgEvent   = LiveChatBaseEvent + "msg:incoming"
	LiveChatSendRoomMsgEvent   = LiveChatBaseEvent + "msg:room:send"
	LiveChatRoomLogEvent       = LiveChatBaseEvent + "msg:room:log"
	LiveChatSendDirectMsgEvent = LiveChatBaseEvent + "msg:dm:send"
	LiveChatDirectLogEvent     = LiveChatBaseEvent + "msg:dm:log"
	LiveChatErrorMsgEvent      = LiveChatBaseEvent + "error"
	LiveChatMsgLogEvent        = LiveChatBaseEvent + "msg:log"
)
