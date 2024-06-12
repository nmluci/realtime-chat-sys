package server

import (
	"context"
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
	UserID       int64
	username     string
	ctx          context.Context
	hub          *LiveChatHub
	conn         *websocket.Conn
	logger       zerolog.Logger
	repo         inrepo.Repository
	in           chan dto.LiveChatSocketEvent
	activeRoomID int64
	isDM         bool
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
			ctx:    ctx,
			hub:    params.Hub,
			conn:   ws,
			logger: params.Logger,
			repo:   params.Repo,
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
				client.username = userMeta.Username
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
		client.ctx = context.Background()

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
				EventName: inconst.LiveChatErrorMsgEvent,
				Data:      "failed to parse msg",
			}
			break
		}

		switch event.EventName {
		case inconst.LiveChatCreateRoomEvent:
			if exists, err := lc.repo.FindRoom(lc.ctx, &indto.ChatRoomParams{RoomName: event.Data.(string)}); err != nil {
				lc.logger.Error().Err(err).Msg("failed to fetch room data")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "failed to fetch room data",
				}
				continue
			} else if exists != nil {
				lc.logger.Error().Err(err).Msg("room already exists")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "room already exists",
				}
				continue
			}

			if err = lc.repo.CreateRoom(lc.ctx, &model.ChatRoom{RoomName: event.Data.(string)}); err != nil {
				lc.logger.Error().Err(err).Msg("failed to create room data")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "failed to create room data",
				}
				continue
			}

			lc.in <- dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatCreatedEvent,
			}
			continue
		case inconst.LiveChatJoinRoomEvent:
			roomMeta, err := lc.repo.FindRoom(lc.ctx, &indto.ChatRoomParams{RoomName: event.Data.(string)})
			if err != nil {
				lc.logger.Error().Err(err).Msg("failed to fetch room data")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "failed to fetch room data",
				}
				continue
			} else if roomMeta == nil {
				lc.logger.Error().Err(err).Msg("room doesnt exists")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "room doesnt exists",
				}
				continue
			}

			lc.hub.JoinRoom(roomMeta.ID, lc)
			lc.in <- dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatJoinedEvent,
			}
			lc.activeRoomID = roomMeta.ID

			continue
		case inconst.LiveChatLeaveRoomEvent:
			roomID := lc.activeRoomID

			roomMeta, err := lc.repo.FindRoom(lc.ctx, &indto.ChatRoomParams{ID: roomID})
			if err != nil {
				lc.logger.Error().Err(err).Msg("failed to fetch room data")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "failed to fetch room data",
				}
				continue
			} else if roomMeta == nil {
				lc.logger.Error().Err(err).Msg("room doesnt exists")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "room doesnt exists",
				}
				continue
			}

			lc.hub.LeaveRoom(roomMeta.ID, lc)
			lc.in <- dto.LiveChatSocketEvent{
				EventName: inconst.LiveChatLeftEvent,
			}
			lc.activeRoomID = 0

			continue
		case inconst.LiveChatSendRoomMsgEvent:
			incomingMessage := &indto.IncomingMessage{
				SenderID:   lc.UserID,
				SenderName: lc.username,
				Content:    event.Data.(string),
				IsDM:       false,
			}

			lc.hub.broadcast <- dto.LiveChatBroadcastEvent{
				Room: lc.activeRoomID,
				Event: dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatIncomingMsgEvent,
					Data:      incomingMessage,
				},
			}
		case inconst.LiveChatSendDirectMsgEvent:
			payload := structutil.MapToStruct[*dto.ChatDMPayload](event.Data.(map[string]any))

			incomingMessage := &indto.IncomingMessage{
				SenderID:   lc.UserID,
				SenderName: lc.username,
				Content:    payload.Content,
				IsDM:       true,
			}

			recipientMeta, err := lc.repo.FindUser(lc.ctx, &indto.UserParams{Username: payload.RecipientUsername})
			if err != nil {
				lc.logger.Error().Err(err).Msg("failed to fetch recipient meta")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "failed to fetch recipient meta",
				}
				continue
			}
			incomingMessage.RecipientID = recipientMeta.ID

			if recipientMeta == nil {
				lc.logger.Error().Err(err).Msg("recipient doesnt existed")
				lc.in <- dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatErrorMsgEvent,
					Data:      "recipient doesnt exists",
				}
				continue
			}

			lc.hub.msgChan <- &dto.LiveChatSocketRequest{
				SenderID:    lc.UserID,
				RecipientID: recipientMeta.ID,
				Event: dto.LiveChatSocketEvent{
					EventName: inconst.LiveChatIncomingMsgEvent,
					Data:      incomingMessage,
				},
			}
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
