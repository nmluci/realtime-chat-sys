package server

import (
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/rs/zerolog"
)

type LiveChatHub struct {
	connectionPool map[int64]*LiveChatSocketMiddleware
	broadcast      chan dto.LiveChatSocketEvent
	register       chan *LiveChatSocketMiddleware
	unregister     chan *LiveChatSocketMiddleware
	logger         zerolog.Logger
	msgChan        chan *dto.LiveChatSocketRequest
	done           chan int
}

type LiveChatHubParms struct {
	Logger   zerolog.Logger
	MsgChan  chan *dto.LiveChatSocketRequest
	DoneChan chan int
}

func NewLiveChatHub(params *LiveChatHubParms) *LiveChatHub {
	return &LiveChatHub{
		connectionPool: make(map[int64]*LiveChatSocketMiddleware),
		broadcast:      make(chan dto.LiveChatSocketEvent),
		register:       make(chan *LiveChatSocketMiddleware),
		unregister:     make(chan *LiveChatSocketMiddleware),
		logger:         params.Logger,
		msgChan:        params.MsgChan,
		done:           params.DoneChan,
	}
}

func (lc *LiveChatHub) Run() {
	for {
		select {
		case conn := <-lc.register:
			lc.connectionPool[0] = conn // TODO: assign distictive ID
		case conn := <-lc.unregister:
			if _, ok := lc.connectionPool[0]; ok {
				delete(lc.connectionPool, 0)
				close(conn.in)
			}
		case msg := <-lc.broadcast:
			for k, conn := range lc.connectionPool {
				select {
				case conn.in <- msg:
				default:
					close(conn.in)
					delete(lc.connectionPool, k)
				}
			}
		case msg := <-lc.msgChan:
			conn, ok := lc.connectionPool[0]
			if !ok {
				lc.logger.Warn().Msg("msg for conn dropped") // TODO: reword warn msg
				continue
			}

			conn.in <- dto.LiveChatSocketEvent{
				EventName: "",
				Data:      msg,
			}
		}

	}
}
