package ws

import (
	"net"
	"net/http"
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	Addr           string
	HTTPTimeout    time.Duration
	MaxHeaderBytes int
	listener       net.Listener
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

func Echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("upgrade: %s", err)
		return
	}
	defer conn.Close()

	for {
		logger.Debug("waiting...")
		mt, message, err := conn.ReadMessage()
		if err != nil {
			logger.Errorf("read: %s", err)
			break
		}

		logger.Debugf("recv: %s", message)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			logger.Errorf("write: %s", err)
			break
		}
	}
}

func (server *WebSocketServer) Start() {
	http.HandleFunc("/", Echo)

	httpServer := &http.Server{
		Addr:           server.Addr,
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: server.MaxHeaderBytes,
	}

	err := httpServer.ListenAndServe()
	if err != nil {
		logger.Fatalf("Failed to listen and serve: %v", err)
	}
}
