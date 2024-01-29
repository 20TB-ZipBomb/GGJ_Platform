package app

import (
	"flag"
	"os"
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/network"
)

var env = flag.String("env", "dev", "server environment")
var verbose = flag.Bool("verbose", false, "enables verbose logging")

func Run() {
	flag.Parse()

	logger.Init()
	defer logger.Sync()

	addr := ":" + os.Getenv("PORT")
	logger.Infof("%s server running on %s", utils.SanitizeEnvFlag(*env), addr)

	server := network.WebSocketServer{
		Addr:           addr,
		HTTPTimeout:    10 * time.Second,
		MaxHeaderBytes: 1024,
	}
	server.Start()
}
