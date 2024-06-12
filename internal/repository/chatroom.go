package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/nmluci/realtime-chat-sys/internal/indto"
	"github.com/nmluci/realtime-chat-sys/internal/model"
	"github.com/rs/zerolog"
)

func (r *repository) FindRooms(ctx context.Context, params *indto.ChatRoomParams) (res []*model.ChatRoom, err error) {
	logger := zerolog.Ctx(ctx)

	stmt, args, err := squirrel.Select("r.id", "r.room_name", "r.is_dm", "(count(rp.id) == 1) is_joined").From("rooms r").
		LeftJoin("room_participants rp on r.id = rp.room_id and rp.user_id = ?", params.UserID).ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate sql")
		return
	}

	res = []*model.ChatRoom{}
	err = r.sqliteDB.QueryRowxContext(ctx, stmt, args...).StructScan(res)
	if err != nil {
		logger.Error().Err(err).Msg("faild to fetch room meta")
		return
	}

	return
}

func (r *repository) FindRoom(ctx context.Context, params *indto.ChatRoomParams) (res *model.ChatRoom, err error) {
	logger := zerolog.Ctx(ctx)

	cond := squirrel.And{}
	if params.ID != 0 {
		cond = append(cond, squirrel.Eq{"id": params.ID})
	} else if params.RoomName != "" {
		cond = append(cond, squirrel.Eq{"room_name": params.RoomName})
	}

	stmt, args, err := squirrel.Select("r.id", "r.room_name").From("rooms r").
		LeftJoin("room_participants rp on r.id = rp.room_id and rp.user_id = ?", params.UserID).
		Where(cond).ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate sql")
		return
	}

	res = &model.ChatRoom{}
	err = r.sqliteDB.QueryRowxContext(ctx, stmt, args...).StructScan(res)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("faild to fetch room meta")
		return
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return
}

func (r *repository) CreateRoom(ctx context.Context, params *model.ChatRoom) (err error) {
	logger := zerolog.Ctx(ctx)

	stmt, args, err := squirrel.Insert("rooms").Columns("room_name").
		Values(params.RoomName).ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate sql")
		return
	}

	_, err = r.sqliteDB.ExecContext(ctx, stmt, args...)
	if err != nil {
		logger.Error().Err(err).Msg("faild to fetch room meta")
		return
	}

	return
}
