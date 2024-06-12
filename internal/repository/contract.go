package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nmluci/realtime-chat-sys/internal/indto"
	"github.com/nmluci/realtime-chat-sys/internal/model"
)

type Repository interface {
	// ----- Users
	FindUser(context.Context, *indto.UserParams) (*model.User, error)
	InsertUser(context.Context, *model.User) error

	// ----- Users
	FindRoom(context.Context, *indto.ChatRoomParams) (*model.ChatRoom, error)
	CreateRoom(context.Context, *model.ChatRoom) error

	// ----- Message
	FindChatHistory(context.Context, *indto.ChatHistoryParams) ([]*model.ChatHistory, error)
	InsertChatHistory(context.Context, *model.ChatHistory) error
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
