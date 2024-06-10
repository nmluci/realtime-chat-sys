package server

import (
	"github.com/labstack/echo/v4"
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/rs/zerolog"
)

func StartServer(logger zerolog.Logger) {
	// TODO: setup db

	ec := echo.New()
	ec.HideBanner = true
	ec.HidePort = true

	msgChan := make(chan *dto.LiveChatSocketRequest, 20)
	doneChan := make(chan int)
	defer func() {
		doneChan <- 1
	}()

	chatHub := NewLiveChatHub(&LiveChatHubParms{
		Logger:   logger,
		MsgChan:  msgChan,
		DoneChan: doneChan,
	})
	go chatHub.Run()

	ec.Any("/api/v1/chat", HandleLiveChatSocket(
		&LiveChatSocketParams{
			Logger: logger,
			Hub:    chatHub,
		}),
	)

	logger.Info().Msg("starting server")
	if err := ec.Start("127.0.0.1:8080"); err != nil {
		logger.Error().Err(err).Msg("failed to start server")
	}
}
