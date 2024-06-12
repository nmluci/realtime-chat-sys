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
		username:   "",
		logger:     &logger,
		conn:       c,
		writerChan: make(chan dto.LiveChatSocketEvent),
	}

	msg := &dto.LiveChatSocketEvent{}
	authenticated := false

	connectedRoomName := ""

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

	go client.reader()
	go client.writer()

	exit := false
	for !exit {
		fmt.Printf("1. Create Room\n2. Join Room\n3. Leave Room\n4. Send DM\n5. Send Message to Room\n9. Exit\n")
		fmt.Printf("Option: ")

		var opt int64
		fmt.Scanf("%d\n", &opt)

		switch opt {
		case 1:
			var roomName string
			fmt.Printf("Room name: ")
			fmt.Scanf("%s\n", &roomName)

			client.writerChan <- dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatCreateRoomEvent,
				Data:      roomName,
			}
		case 2:
			var roomName string
			fmt.Printf("Room name: ")
			_, err := fmt.Scanf("%s\n", &roomName)
			if err != nil {
				logger.Error().Err(err).Send()
			}

			fmt.Printf("room")
			connectedRoomName = roomName
			client.writerChan <- dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatJoinRoomEvent,
				Data:      roomName,
			}
		case 3:
			var roomName string
			fmt.Printf("Room name: ")
			fmt.Scanf("%s\n", &roomName)

			connectedRoomName = ""
			client.writerChan <- dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatLeaveRoomEvent,
				Data:      roomName,
			}
		case 4:
			var recipient, content string
			fmt.Printf("Recipient username: ")
			fmt.Scanf("%s\n", &recipient)

			fmt.Printf("Content: ")
			fmt.Scanf("%s\n", &content)

			client.writerChan <- dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatSendDirectMsgEvent,
				Data: dto.ChatDMPayload{
					RecipientUsername: recipient,
					Content:           content,
				},
			}
		case 5:
			if connectedRoomName == "" {
				continue
			}

			fmt.Printf("Current Connected Room: %s\n", connectedRoomName)

			var content string
			fmt.Printf("Content: ")
			fmt.Scanf("%s\n", &content)

			client.writerChan <- dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatSendRoomMsgEvent,
				Data:      content,
			}
		case 9:
			exit = true
			os.Exit(0)
		}
	}
}
