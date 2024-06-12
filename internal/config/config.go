package config

import (
	"fmt"
	"log"
	"time"
)

type Config struct {
	ServiceName    string
	ServiceAddress string
	Environment    Environment

	FilePath string
	RunSince time.Time

	SqliteDBConfig SqliteDBConfig
}

const logTagConfig = "[Init Config]"

var config *Config

func Init(envString string) {
	conf := Config{
		SqliteDBConfig: SqliteDBConfig{
			DBName: "app.db",
		},
	}

	if envString != "dev" && envString != "prod" && envString != "local" {
		log.Fatalf("%s environment must be either local, dev or prod, found: %s", logTagConfig, envString)
	}

	conf.Environment = Environment(envString)
	conf.ServiceName = fmt.Sprintf("realtime-chat-%s", conf.Environment)

	switch conf.Environment {
	case EnvironmentLocal:
		conf.ServiceAddress = "localhost:8080"
		conf.FilePath = "appdata"
	default:
		conf.ServiceAddress = ":8080"
		conf.FilePath = "/appdata"
	}

	conf.RunSince = time.Now()
	config = &conf
}

func Get() (conf *Config) {
	conf = config
	return
}
