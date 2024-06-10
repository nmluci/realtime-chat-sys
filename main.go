package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nmluci/realtime-chat-sys/cmd/server"
	"github.com/nmluci/realtime-chat-sys/internal/component"
)

func main() {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}

	time.Local = loc

	logger := component.NewLogger(component.NewLoggerParams{
		ServiceName: "realtime-chat",
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
		server.StartServer(logger)
	case "client":
		fmt.Println("unimplemented lol")
	}
}
