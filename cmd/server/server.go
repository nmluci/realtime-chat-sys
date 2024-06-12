package server

import (
	"github.com/labstack/echo/v4"
	"github.com/nmluci/realtime-chat-sys/internal/component"
	"github.com/nmluci/realtime-chat-sys/internal/config"
	"github.com/nmluci/realtime-chat-sys/internal/repository"
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/rs/zerolog"
)

func StartServer(conf *config.Config, logger zerolog.Logger) {
	// initialize SQlite-based DB for chat history
	db, err := component.NewSQliteDB(&component.NewSQliteDBParams{
		Logger: logger,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialized mariaDB")
	}

	ec := echo.New()
	ec.HideBanner = true
	ec.HidePort = true

	msgChan := make(chan *dto.LiveChatSocketRequest, 20)
	doneChan := make(chan int)
	defer func() {
		doneChan <- 1
	}()

	repo := repository.NewRepository(&repository.NewRepositoryParams{
		SQLiteDB: db,
	})

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
			Repo:   repo,
		}),
	)

	logger.Info().Msg("starting server")
	if err := ec.Start(conf.ServiceAddress); err != nil {
		logger.Error().Err(err).Msg("failed to start server")
	}
}
