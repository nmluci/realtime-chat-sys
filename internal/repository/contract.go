package repository

import "github.com/jmoiron/sqlx"

type Repository interface {
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
