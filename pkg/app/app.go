package app

import (
	"flag"
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/ws"
)

func Run() {
	logger.Init()
	defer logger.Sync()

	addr := flag.Lookup("addr").Value.String()

	server := ws.WebSocketServer{
		Addr:           addr,
		HTTPTimeout:    10 * time.Second,
		MaxHeaderBytes: 1024,
	}
	server.Start()
}
