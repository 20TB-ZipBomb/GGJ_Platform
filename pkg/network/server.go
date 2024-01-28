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

		logger.Infof("New websocket connection made on %s.", r.URL)
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

// The main service and entrypoint for serving new clients via websocket connections.
func serveWebSocket(s *WebSocketServer, w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Upgrade error: %v", err)
		return
	}
	defer c.Close()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("Error reading message: %v", err)
			}
			break
		}

		logger.Debugf("Received: %s", msg)

		switch msgJSON := json.UnmarshalJSON[Message](msg); msgJSON.MessageType {
		case CreateLobby:
			s.tryCreateLobby(c)
		case LobbyJoinAttempt:
			ljam := json.UnmarshalJSON[LobbyJoinAttemptMessage](msg)
			s.tryAddClientToLobby(c, &ljam)
		default:
			rejectConnection(c)
		}
	}
}

// Attempts to create a new lobby on the server and initialize the "hosting" game client.
// Note that only one lobby can exist on the server at a given time, so redundant requests to create lobbies are ignored.
func (s *WebSocketServer) tryCreateLobby(c *websocket.Conn) {
	if s.lobby != nil {
		logger.Warn("Attempting to create another lobby on this server when one already exists or is in-progress. Ignoring.")
		return
	}

	s.lobby = createLobby()
	go s.lobby.run()

	client := createClient(s.lobby, c, Game)
	client.lobby.register <- client
}

// Attempts to add a web client to the server's lobby.
// This operation requires that messages sent by the client adhere to the `LobbyJoinAttemptMessage` specification.
func (s *WebSocketServer) tryAddClientToLobby(c *websocket.Conn, ljam *LobbyJoinAttemptMessage) {
	if s.lobby == nil {
		logger.Warn("Lobby join request was recevied, but no game has been created yet.")
		rejectConnection(c)
		return
	}

	if err := ljam.Verify(&s.lobby.lobbyCode); err != nil {
		logger.Warnf("Lobby join failure: %v", err)
		rejectConnection(c)
		return
	}

	client := createClient(s.lobby, c, Web)
	client.Name = *ljam.PlayerName
	client.lobby.register <- client

	// Send a message to the game client indicating that a web client has connected.
	pjam := &PlayerJoinedMessage{
		Message: Message{
			MessageType: PlayerJoined,
		},
		Player: Player{
			PlayerID: client.ID,
			Name:     client.Name,
		},
	}
	client.lobby.unicastGame <- []byte(json.MarshalJSON[PlayerJoinedMessage](pjam))
}

// Rejects an incoming connection, responding with a connection rejected message.
func rejectConnection(c *websocket.Conn) {
	connRejMsg := &Message{
		MessageType: ConnectionRejected,
	}

	c.WriteJSON(connRejMsg)
}
