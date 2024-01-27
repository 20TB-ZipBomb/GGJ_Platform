package network

import (
	// "encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils/json"
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

// Echo endpoint used for testing
func Echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			logger.Errorf("ReadMessage error: %s", err)
			break
		}

		logger.Debugf("Received: %s", message)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			logger.Errorf("WriteMessage error: %s", err)
			break
		}
	}
}

func serveWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Errorf("ReadMessage error: %s", err)
			break
		}

		logger.Debugf("Received: %s", msg)
		recvMsg, _ := json.UnmarshalJSON[Message](msg)
		logger.Debugf("Unmarshal: %s", recvMsg.Type)

		if recvMsg.Type == CreateLobby {
			lobby := newLobby()
			go lobby.run()

			client := &Client{
				clientType: Game,
				lobby:      lobby,
				conn:       conn,
				send:       make(chan []byte, 256),
			}

			client.lobby.register <- client
		} else {
			// If this message is anything else it's not the game client and we can refuse it
			connRejJSON := CreateMessageJSON(ConnectionRejected)
			err = conn.WriteJSON(connRejJSON)
		}
	}
}

func (server *WebSocketServer) Start() {
	http.HandleFunc("/echo", Echo)
	http.HandleFunc("/connect", serveWebSocket)

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
