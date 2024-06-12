package client

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/nmluci/realtime-chat-sys/internal/inconst"
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/rs/zerolog"
)

func StartClient(logger zerolog.Logger) {
	logger.Info().Msg("starting client")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/api/v1/chat", nil)
	if err != nil {
		logger.Error().Err(err).Msg("failed to connect to server")
		return
	}
	defer c.Close()

	client := &LiveClient{
		logger: &logger,
		conn:   c,
	}

	msg := &dto.LiveChatSocketEvent{}
	authenticated := false

	for !authenticated {
		fmt.Printf("1. Login\n2. Sign Up\n9. Exit\n")
		fmt.Printf("Option: ")

		var opt int64
		fmt.Scanf("%d\n", &opt)

		switch opt {
		case 1:
			var username, password string
			fmt.Printf("Username: ")
			fmt.Scanf("%s\n", &username)

			fmt.Printf("Password: ")
			fmt.Scanf("%s\n", &password)

			err := c.WriteJSON(dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatAuthLoginEvent,
				Data: dto.AuthLoginPayload{
					Username: username,
					Password: password,
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("failed to send message")
				continue
			}

			err = c.ReadJSON(msg)
			if err != nil {
				logger.Error().Err(err).Msg("failed to parse message")
				continue
			}

			switch msg.EventName {
			case inconst.LiveChatErrorMsgEvent:
				logger.Error().Err(err).Msg(msg.Data.(string))
				continue
			case inconst.LiveChatAuthAckEvent:
				client.username = username
				authenticated = true
			}

		case 2:
			var username, password string
			fmt.Printf("Username: ")
			fmt.Scanf("%s\n", &username)

			fmt.Printf("Password: ")
			fmt.Scanf("%s\n", &password)

			err := c.WriteJSON(dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatAuthSignupEvent,
				Data: dto.AuthLoginPayload{
					Username: username,
					Password: password,
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("failed to send message")
				continue
			}

			err = c.ReadJSON(msg)
			if err != nil {
				logger.Error().Err(err).Msg("failed to parse message")
				continue
			}

			switch msg.EventName {
			case inconst.LiveChatErrorMsgEvent:
				logger.Error().Err(err).Msg(msg.Data.(string))
				continue
			}
		case 9:
			os.Exit(0)
		}
	}

}
