package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/rs/zerolog"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 1024
)

type LiveChatSocketParams struct {
	Logger zerolog.Logger
	Hub    *LiveChatHub
}

var (
	newline  = []byte{'\n'}
	space    = []byte{' '}
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type LiveChatSocketMiddleware struct {
	hub      *LiveChatHub
	conn     *websocket.Conn
	logger   zerolog.Logger
	in       chan dto.LiveChatSocketEvent
	retry    chan dto.LiveChatSocketEvent
	retryMap map[string]struct{}
}

func HandleLiveChatSocket(params *LiveChatSocketParams) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			params.Logger.Error().Err(err).Msg("failed to upgrade connection to WS")
			return err
		}

		client := &LiveChatSocketMiddleware{
			hub:      params.Hub,
			conn:     ws,
			logger:   params.Logger,
			in:       make(chan dto.LiveChatSocketEvent, 256),
			retry:    make(chan dto.LiveChatSocketEvent, 256),
			retryMap: make(map[string]struct{}),
		}

		msg := &dto.LiveChatSocketEvent{}
		if err := ws.ReadJSON(msg); err != nil {
			params.Logger.Error().Err(err).Msg("failed to parse initial msg")
			ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()), time.Now().Add(time.Second))
			ws.Close()
			return nil
		} else if msg.EventName != "" { // TODO: set cond
			ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
			ws.Close()
			return nil
		}

		client.hub.register <- client

		bJson, err := json.Marshal(&dto.LiveChatSocketEvent{
			EventName: "",
			Data:      nil,
		})
		if err != nil {
			params.Logger.Error().Err(err).Msg("failed to marshal msg")
			return
		}

		err = ws.WriteMessage(websocket.TextMessage, bJson)
		if err != nil {
			params.Logger.Error().Err(err).Msg("failed to write msg")
			return
		}

		go client.Reader()
		go client.Writer()

		return
	}
}

func (lc *LiveChatSocketMiddleware) Reader() {
	defer func() {
		lc.hub.unregister <- lc
		lc.conn.Close()
	}()

	lc.conn.SetReadLimit(maxMsgSize)
	lc.conn.SetReadDeadline(time.Now().Add(pongWait))
	lc.conn.SetPongHandler(func(string) error { lc.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		event := &dto.LiveChatSocketEvent{}
		err := lc.conn.ReadJSON(event)
		if err != nil {
			lc.logger.Error().Err(err).Msg("failed to parse msg")
			lc.in <- dto.LiveChatSocketEvent{
				EventName: "", // TODO: define const
				Data:      "failed to parse msg",
			}
			break
		}
	}
}

func (lc *LiveChatSocketMiddleware) Writer() {
	ticker := time.NewTicker(pingPeriod)
	retryTicker := time.NewTicker(2 * time.Second)
	defer func() {
		ticker.Stop()
		retryTicker.Stop()
		lc.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-lc.in:
			lc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				lc.logger.Info().Msg("conn closed")
				lc.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			bJson, err := json.Marshal(msg)
			if err != nil {
				lc.logger.Error().Err(err).Msg("failed to marshal msg")
				return
			}

			err = lc.conn.WriteMessage(websocket.TextMessage, bJson)
			if err != nil {
				lc.logger.Error().Err(err).Msg("failed to write msg")
				return
			}
		case <-ticker.C:
			lc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := lc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				lc.logger.Error().Err(err).Msg("failed to send pong")
				return
			}
		}
	}
}
