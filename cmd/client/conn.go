package client

import (
	"github.com/gorilla/websocket"
	"github.com/nmluci/realtime-chat-sys/internal/inconst"
	"github.com/nmluci/realtime-chat-sys/internal/indto"
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/nmluci/realtime-chat-sys/pkg/structutil"
	"github.com/rs/zerolog"
)

type LiveClient struct {
	username string
	logger   *zerolog.Logger
	conn     *websocket.Conn
}

func (lc *LiveClient) reader() {
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
			lc.logger.Info().Str("Sender", meta.SenderName).Msg(meta.Content)
			continue
		}

	}
}
