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

// todo: Consolidate use of websocket.Upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

func ServeWebSocket(w http.ResponseWriter, r *http.Request) {
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
		msgJSON := json.UnmarshalJSON[Message](msg)
		logger.Debugf("Unmarshal: %s", msgJSON.MessageType)

		// todo: Add checks for attempting to join the lobby here, deny them if the specified lobby doesn't exist

		if msgJSON.MessageType == CreateLobby {
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
			rejMsg := &Message{
				MessageType: ConnectionRejected,
			}
			rejJSON := json.MarshalJSON[Message](rejMsg)
			conn.WriteJSON(rejJSON)
		}
	}
}

func (server *WebSocketServer) Start() {
	http.HandleFunc("/connect", ServeWebSocket)

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
