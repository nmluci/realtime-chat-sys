package component

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/nmluci/realtime-chat-sys/internal/config"
	_ "modernc.org/sqlite"

	"github.com/rs/zerolog"
)

type NewSQliteDBParams struct {
	Logger zerolog.Logger
}

func NewSQliteDB(params *NewSQliteDBParams) (db *sqlx.DB, err error) {
	conf := config.Get()

	connString := filepath.Join(conf.FilePath, "db", conf.SqliteDBConfig.DBName)

	// check if DB already existed
	_, err = os.Stat(connString)
	if os.IsNotExist(err) {
		os.Create(connString)
	}

	db, err = sqlx.Connect("sqlite", connString)
	if err != nil {
		params.Logger.Error().Err(err).Msg("failed to connect to db")
		return
	}

	params.Logger.Info().Msg("db init successfully")

	dbMigrate, err := migrate.New("file://migrations", fmt.Sprintf("sqlite://%s", filepath.ToSlash(connString)))
	if err != nil {
		params.Logger.Error().Err(err).Msg("failed to connect to migration engine")
		return
	}

	if err = dbMigrate.Up(); err != nil && err != migrate.ErrNoChange {
		params.Logger.Error().Err(err).Msg("failed to perform migrations")
		return
	}

	rev, isDirty, err := dbMigrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		params.Logger.Error().Err(err).Msg("failed to fetch migration version")
		return
	}

	if isDirty {
		params.Logger.Warn().Msg("SQlite migration is dirty")
	}

	if err == migrate.ErrNilVersion {
		params.Logger.Info().Msg("SQlite Migration Version: None")
	} else {
		params.Logger.Info().Msgf("SQlite Migration Version: %d", rev)
	}

	return
}
