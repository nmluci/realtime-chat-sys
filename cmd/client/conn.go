package client

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/nmluci/realtime-chat-sys/internal/inconst"
	"github.com/nmluci/realtime-chat-sys/internal/indto"
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/nmluci/realtime-chat-sys/pkg/structutil"
	"github.com/rs/zerolog"
)

type LiveClient struct {
	username   string
	logger     *zerolog.Logger
	conn       *websocket.Conn
	writerChan chan dto.LiveChatSocketEvent
}

func (lc *LiveClient) reader() {
	defer lc.conn.Close()

	lc.conn.SetReadLimit(256)
	lc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	lc.conn.SetPongHandler(func(string) error { lc.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

	for {
		event := &dto.LiveChatSocketEvent{}
		err := lc.conn.ReadJSON(event)
		if err != nil {
			lc.logger.Error().Err(err).Msg("failed to parse msg")
			break
		}

		switch event.EventName {
		case inconst.LiveChatErrorMsgEvent:
			lc.logger.Error().Err(err).Msg(event.Data.(string))
			continue
		case inconst.LiveChatIncomingMsgEvent:
			meta := structutil.MapToStruct[*indto.IncomingMessage](event.Data.(map[string]any))
			lc.logger.Info().Str("Sender", meta.SenderName).Str("Content", meta.Content).Bool("FromDM", meta.IsDM).Msg(meta.Content)
			continue
		case inconst.LiveChatCreateRoomEvent:
			lc.logger.Info().Msg("room created")
			continue
		case inconst.LiveChatJoinedEvent:
			lc.logger.Info().Msg("room joined")
			continue
		case inconst.LiveChatLeftEvent:
			lc.logger.Info().Msg("room left")
			continue
		}

	}
}

func (lc *LiveClient) writer() {
	ticker := time.NewTicker((6 * 9 / 10 * time.Second))
	defer func() {
		ticker.Stop()
		lc.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-lc.writerChan:
			lc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				lc.logger.Info().Msg("conn closed")
				lc.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := lc.conn.WriteJSON(msg)
			if err != nil {
				lc.logger.Error().Err(err).Msg("failed to write message")
				continue
			}
		case <-ticker.C:
			lc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := lc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				lc.logger.Error().Err(err).Msg("failed to send pong")
				return
			}
		}
	}
}
