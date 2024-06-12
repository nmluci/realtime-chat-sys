package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/nmluci/realtime-chat-sys/internal/indto"
	"github.com/nmluci/realtime-chat-sys/internal/model"
	"github.com/rs/zerolog"
)

func (r *repository) FindUser(ctx context.Context, params *indto.UserParams) (res *model.User, err error) {
	logger := zerolog.Ctx(ctx)

	stmt, args, err := squirrel.Select("id", "username", "password").From("users").Where(squirrel.And{
		squirrel.Eq{"username": params.Username},
	}).ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate sql stmt")
		return
	}

	res = &model.User{}
	err = r.sqliteDB.QueryRowxContext(ctx, stmt, args...).StructScan(res)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("failed to fetch userdata")
		return
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return
}

func (r *repository) InsertUser(ctx context.Context, params *model.User) (err error) {
	logger := zerolog.Ctx(ctx)

	stmt, args, err := squirrel.Insert("users").Columns("username", "password").
		Values(params.Username, params.Password).ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate sql stmt")
		return
	}

	_, err = r.sqliteDB.ExecContext(ctx, stmt, args...)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch userdata")
		return
	}

	return
}
