package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nmluci/realtime-chat-sys/cmd/client"
	"github.com/nmluci/realtime-chat-sys/cmd/server"
	"github.com/nmluci/realtime-chat-sys/internal/component"
	"github.com/nmluci/realtime-chat-sys/internal/config"
	"github.com/nmluci/realtime-chat-sys/pkg/initutil"
)

var (
	environment string = "local"
)

func main() {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}

	time.Local = loc

	config.Init(environment)
	conf := config.Get()

	initutil.InitDirectory()

	logger := component.NewLogger(component.NewLoggerParams{
		ServiceName: conf.ServiceName,
		PrettyPrint: true,
	})

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("\t<app name> server \t run in server mode")
		fmt.Println("\t<app name> client \t run in client mode")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "server":
		server.StartServer(conf, logger)
	case "client":
		client.StartClient(logger)
		fmt.Println("unimplemented lol")
	}
}
