package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nmluci/realtime-chat-sys/internal/indto"
	"github.com/nmluci/realtime-chat-sys/internal/model"
)

type Repository interface {
	FindUser(context.Context, *indto.UserParams) (*model.User, error)
	InsertUser(context.Context, *model.User) error
}

type repository struct {
	sqliteDB *sqlx.DB
}

type NewRepositoryParams struct {
	SQLiteDB *sqlx.DB
}

func NewRepository(params *NewRepositoryParams) Repository {
	return &repository{
		sqliteDB: params.SQLiteDB,
	}
}
