package app

import (
	"flag"
	"strconv"
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/network"
)

var ip = flag.String("ip", "localhost", "target ip address")
var port = flag.Int("port", 4041, "target port")

func Run() {
	flag.Parse()

	logger.Init()
	defer logger.Sync()

	addr := *ip + ":" + strconv.Itoa(*port)
	logger.Infof("Server running on %s", addr)

	server := network.WebSocketServer{
		Addr:           addr,
		HTTPTimeout:    10 * time.Second,
		MaxHeaderBytes: 1024,
	}
	server.Start()
}
