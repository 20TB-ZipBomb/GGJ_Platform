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
	lobby          *Lobby
	upgrader       websocket.Upgrader
}

func (server *WebSocketServer) Start() {
	server.lobby = nil
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		server.upgrader = websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		}

		logger.Infof("New websocket connection made on %s", r.URL)
		serveWebSocket(server, w, r)
	})

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

// The main service and entrypoint for serving new clients via websocket connections
func serveWebSocket(s *WebSocketServer, w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
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

		switch msgJSON := json.UnmarshalJSON[Message](msg); msgJSON.MessageType {
		case CreateLobby:
			if s.lobby == nil {
				s.lobby = newLobby()
				go s.lobby.run()
			}

			client := CreateGameClient(s.lobby, conn)
			client.lobby.register <- client
		case LobbyJoinAttempt:
			if s.lobby == nil {
				logger.Debug("lobby is nil")
				rejectConnection(conn)
				continue
			}

			lobbyJoinJSON := json.UnmarshalJSON[LobbyJoinAttemptMessage](msg)
			logger.Debugf("%s requested lobby %s", lobbyJoinJSON.PlayerName, lobbyJoinJSON.LobbyCode)

			// client := CreateWebClient(lobby, conn)
			// client.lobby.register <- client
		default:
			rejectConnection(conn)
		}
	}
}

// Rejects an incoming connection, responding with a connection rejected message
func rejectConnection(conn *websocket.Conn) {
	connRejMsg := &Message{
		MessageType: ConnectionRejected,
	}

	conn.WriteJSON(json.MarshalJSON[Message](connRejMsg))
}
