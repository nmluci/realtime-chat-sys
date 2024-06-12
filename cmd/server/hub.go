package server

import (
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/rs/zerolog"
)

type LiveChatHub struct {
	connectionPool map[int64]*LiveChatSocketMiddleware
	rooms          *rooms
	broadcast      chan dto.LiveChatBroadcastEvent
	register       chan *LiveChatSocketMiddleware
	unregister     chan *LiveChatSocketMiddleware
	logger         zerolog.Logger
	msgChan        chan *dto.LiveChatSocketRequest
	doneChan       chan int
}

type LiveChatHubParms struct {
	Logger   zerolog.Logger
	MsgChan  chan *dto.LiveChatSocketRequest
	DoneChan chan int
}

func NewLiveChatHub(params *LiveChatHubParms) *LiveChatHub {
	return &LiveChatHub{
		connectionPool: make(map[int64]*LiveChatSocketMiddleware),
		rooms:          newRooms(),
		broadcast:      make(chan dto.LiveChatBroadcastEvent),
		register:       make(chan *LiveChatSocketMiddleware),
		unregister:     make(chan *LiveChatSocketMiddleware),
		logger:         params.Logger,
		msgChan:        params.MsgChan,
		doneChan:       params.DoneChan,
	}
}

func (lc *LiveChatHub) Run() {
	for {
		select {
		case conn := <-lc.register:
			lc.connectionPool[conn.UserID] = conn
		case conn := <-lc.unregister:
			if _, ok := lc.connectionPool[conn.UserID]; ok {
				delete(lc.connectionPool, conn.UserID)
				close(conn.in)
			}
		case msg := <-lc.broadcast:
			connMap := lc.rooms.getRoom(msg.Room)

			event := msg.Event
			for k, conn := range connMap {
				select {
				case conn.in <- event:
				default:
					close(conn.in)
					delete(lc.connectionPool, k)
				}
			}
		case msg := <-lc.msgChan:
			recipient, ok := lc.connectionPool[msg.RecipientID]
			if !ok {
				lc.logger.Warn().Msg("message dropped due to recipient unreachable")
				continue
			}

			recipient.in <- msg.Event
			lc.connectionPool[msg.SenderID].in <- msg.Event
		}

	}
}

func (lc *LiveChatHub) JoinRoom(roomID int64, conn *LiveChatSocketMiddleware) {
	lc.rooms.joinRoom(roomID, conn)
}

func (lc *LiveChatHub) LeaveRoom(roomID int64, conn *LiveChatSocketMiddleware) {
	lc.rooms.leaveRoom(roomID, conn)
}
