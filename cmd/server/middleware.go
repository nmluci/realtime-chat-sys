package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/nmluci/realtime-chat-sys/internal/inconst"
	"github.com/nmluci/realtime-chat-sys/internal/indto"
	"github.com/nmluci/realtime-chat-sys/internal/model"
	inrepo "github.com/nmluci/realtime-chat-sys/internal/repository"
	"github.com/nmluci/realtime-chat-sys/pkg/dto"
	"github.com/nmluci/realtime-chat-sys/pkg/errs"
	"github.com/nmluci/realtime-chat-sys/pkg/structutil"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 1024
)

type LiveChatSocketParams struct {
	Repo   inrepo.Repository
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
	UserID int64
	hub    *LiveChatHub
	conn   *websocket.Conn
	logger zerolog.Logger
	in     chan dto.LiveChatSocketEvent
}

func HandleLiveChatSocket(params *LiveChatSocketParams) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		ctx := c.Request().Context()

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			params.Logger.Error().Err(err).Msg("failed to upgrade connection to WS")
			return err
		}

		client := &LiveChatSocketMiddleware{
			hub:    params.Hub,
			conn:   ws,
			logger: params.Logger,
			in:     make(chan dto.LiveChatSocketEvent, 256),
		}

		authenticated := false
		msg := &dto.LiveChatSocketEvent{}

		sendMessage := func(msg any) (err error) {
			if v, ok := msg.(error); ok {
				msg = v.Error()
			}

			bJson, err := json.Marshal(dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatErrorMsgEvent,
				Data:      msg,
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

			return
		}

		for !authenticated {
			if err := ws.ReadJSON(msg); err != nil {
				params.Logger.Error().Err(err).Msg("failed to parse initial msg")
				ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()), time.Now().Add(time.Second))
				ws.Close()
				return nil
			}

			switch msg.EventName {
			case inconst.LiveChatAuthLoginEvent:
				cred := structutil.MapToStruct[*dto.AuthLoginPayload](msg.Data.(map[string]any))
				if cred == nil {
					client.logger.Error().Err(err).Msg("failed to parse msg body")
					continue
				}

				userMeta, err := params.Repo.FindUser(ctx, &indto.UserParams{Username: cred.Username})
				if err != nil || userMeta == nil {
					params.Logger.Error().Err(err).Msg("failed to validate user")

					sendMessage(errs.ErrInvalidCred)
					continue
				}

				if err = bcrypt.CompareHashAndPassword([]byte(userMeta.Password), []byte(cred.Password)); err != nil {
					params.Logger.Error().Err(err).Msg("failed to validate credentials")

					sendMessage(errs.ErrInvalidCred)
					continue
				}

				client.UserID = userMeta.ID
				authenticated = true

				params.Logger.Info().Str("username", userMeta.Username).Msg("user logged in")
			case inconst.LiveChatAuthSignupEvent:
				cred := structutil.MapToStruct[*dto.AuthLoginPayload](msg.Data.(map[string]any))
				if cred == nil {
					client.logger.Error().Err(err).Msg("failed to parse msg body")
					continue
				}

				userMeta, err := params.Repo.FindUser(ctx, &indto.UserParams{Username: cred.Username})
				if err != nil {
					params.Logger.Error().Err(err).Msg("failed to validate user")

					sendMessage(errs.ErrInvalidCred)
					continue
				}

				if userMeta != nil {
					sendMessage(errs.ErrUserExisted)
					continue
				}

				hashed, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
				if err != nil {
					params.Logger.Error().Err(err).Msg("failed to hash password")

					sendMessage(errs.ErrUnknown)
					continue
				}

				newUser := &model.User{
					Username: cred.Username,
					Password: string(hashed),
				}

				err = params.Repo.InsertUser(ctx, newUser)
				if err != nil {
					params.Logger.Error().Err(err).Msg("failed to save usermeta")

					sendMessage(errs.ErrUnknown)
					continue
				}

				sendMessage("ok")
			default:
				sendMessage("not yet authenticated")
			}
		}

		client.hub.register <- client

		bJson, err := json.Marshal(&dto.LiveChatSocketEvent{
			EventName: inconst.LiveChatAuthAckEvent,
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
