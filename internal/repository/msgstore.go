package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/nmluci/realtime-chat-sys/internal/indto"
	"github.com/nmluci/realtime-chat-sys/internal/model"
	"github.com/rs/zerolog"
)

func (r *repository) InsertChatHistory(ctx context.Context, params *model.ChatHistory) (err error) {
	logger := zerolog.Ctx(ctx)

	stmt, args, err := squirrel.Insert("chat_histories").Columns("room_id", "sender_id", "recipient_id", "message").
		Values(params.RoomID, params.SenderID, params.RecipientID, params.Message).ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate sql")
		return
	}

	_, err = r.sqliteDB.ExecContext(ctx, stmt, args...)
	if err != nil {
		logger.Error().Err(err).Msg("faild to insert chat")
		return
	}

	return
}

func (r *repository) FindChatHistory(ctx context.Context, params *indto.ChatHistoryParams) (res []*model.ChatHistory, err error) {
	logger := zerolog.Ctx(ctx)

	cond := squirrel.And{}
	if params.RoomName != "" {
		cond = append(cond, squirrel.Eq{"r.name": params.RoomName})
	}

	if params.RoomID != 0 {
		cond = append(cond, squirrel.Eq{"ch.room_id": params.RoomID})
	}

	if params.UserID != 0 {
		cond = append(cond, squirrel.Eq{"ch.sender_id": params.UserID})

		if params.RoomName != "" {
			cond = append(cond, squirrel.Eq{"ch.recipient_id": params.UserID}, squirrel.Eq{"ch.room_id": 0})
		}
	}

	stmt, args, err := squirrel.Select("ch.id", "ch.room_id", "ch.sender_id", "su.username sender_name", "ch.recipient_id", "coalesce(ru.username, '') recipient_name", "ch.message").From("chat_histories ch").
		LeftJoin("users su on ch.sender_id = su.id").
		LeftJoin("users ru on ch.recipient = ru.id and ch.recipient_id <> 0").
		LeftJoin("rooms r on r.id = ch.room_id and ch.room_id <> 0").
		Where(cond).
		ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate sql")
		return
	}

	res = []*model.ChatHistory{}

	rows, err := r.sqliteDB.QueryxContext(ctx, stmt, args...)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch room meta")
		return
	}

	for rows.Next() {
		temp := &model.ChatHistory{}

		if err = rows.StructScan(&temp); err != nil {
			logger.Error().Err(err).Msg("failed to map row result")
			return
		}

		res = append(res, temp)
	}

	return

}
