package network

import (
	// "encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils"
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils/json"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/game"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/pack"
	"github.com/gorilla/websocket"
)

const (
	// Minimum number of players
	minimumNumberOfPlayers = 3
)

type WebSocketServer struct {
	Addr           string
	HTTPTimeout    time.Duration
	MaxHeaderBytes int
	listener       net.Listener
	lobby          *Lobby
	upgrader       websocket.Upgrader
	gameState      *game.State
}

func (server *WebSocketServer) Start() {
	server.lobby = nil
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		server.upgrader = websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		}

		logger.Infof("[server] New websocket connection made on %s.", r.URL)
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
		logger.Fatalf("[server] Failed to listen and serve: %v", err)
	}
}

// The main service and entrypoint for serving new clients via websocket connections.
func serveWebSocket(s *WebSocketServer, w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("[server] Upgrade error: %v", err)
		return
	}

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			// Filter generic close errors
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("[server] Error reading message: %v", err)
			}

			// If the lobby has been created, treat this like a disconnect
			if s.lobby != nil && c != nil {
				if disconnectedClient := s.lobby.socketsToClients[c]; disconnectedClient.clientType == Game {
					s.lobby.disconnect <- c
					s.lobby.closeLobby()
					s.lobby = nil
					if s.gameState != nil {
						s.gameState.Reset()
					}
				} else {
					disconnectedClient.closeClient()
				}
			}
			break
		}

		strMsg := string(msg)
		logger.Verbosef("[payload] %s", strMsg[:len(strMsg)-1])

		switch msgJSON := json.UnmarshalJSON[pack.Message](msg); msgJSON.MessageType {
		case pack.CreateLobby:
			s.tryCreateLobby(c)
		case pack.LobbyJoinAttempt:
			ljam := json.UnmarshalJSON[pack.LobbyJoinAttemptMessage](msg)
			s.tryAddClientToLobby(c, &ljam)
		case pack.GameStart:
			s.startGame(c)
		case pack.JobSubmitted:
			jsm := json.UnmarshalJSON[pack.JobSubmittedMessage](msg)
			s.addJobToGameState(c, &jsm)
		default:
			rejectConnection(c)
		}
	}
}

// Attempts to create a new lobby on the server and initialize the "hosting" game client.
// Note that only one lobby can exist on the server at a given time, so redundant requests to create lobbies are ignored.
func (s *WebSocketServer) tryCreateLobby(c *websocket.Conn) {
	if s.lobby != nil {
		logger.Warn("[server] Attempting to create another lobby on this server when one already exists or is in-progress. Ignoring.")
		return
	}

	s.lobby = createLobby()
	go s.lobby.run()

	client := createClient(s.lobby, c, Game)
	client.lobby.register <- client
}

// Attempts to add a web client to the server's lobby.
// This operation requires that messages sent by the client adhere to the `LobbyJoinAttemptMessage` specification.
func (s *WebSocketServer) tryAddClientToLobby(c *websocket.Conn, ljam *pack.LobbyJoinAttemptMessage) {
	if s.lobby == nil {
		logger.Warn("[server] Lobby join request was recevied, but no lobby has been created yet.")
		rejectConnection(c)
		return
	}

	if err := ljam.Verify(&s.lobby.lobbyCode); err != nil {
		logger.Warnf("[server] Lobby join failure: %v", err)
		rejectConnection(c)
		return
	}

	client := createClient(s.lobby, c, Web)
	client.Name = *ljam.PlayerName
	client.lobby.register <- client

	// Send a message to the game client indicating that a web client has connected.
	pjam := &pack.PlayerJoinedMessage{
		Message: pack.Message{
			MessageType: pack.PlayerJoined,
		},
		Player: pack.Player{
			PlayerID: client.UUID,
			Name:     &client.Name,
		},
	}
	client.lobby.unicastGame <- []byte(json.MarshalJSON[pack.PlayerJoinedMessage](pjam))
}

// Echoes a start game request to all clients on the server.
func (s *WebSocketServer) startGame(c *websocket.Conn) {
	if s.lobby == nil {
		logger.Warn("[server] Start game request was recevied, but no lobby has been created yet.")
		rejectConnection(c)
		return
	}

	numPlayers := len(s.lobby.webClients)
	// todo: Remove production environment constraint for minimum number of players?
	if utils.IsProductionEnv() && numPlayers < minimumNumberOfPlayers {
		logger.Warn("Start game request was received, but the lobby has less than the minimum amount of clients connected that are required to play.")
		rejectConnection(c)
		return
	}

	s.gameState = game.CreateGameState(numPlayers)

	sgm := &pack.Message{
		MessageType: pack.GameStart,
	}
	s.lobby.broadcast <- []byte(json.MarshalJSON[pack.Message](sgm))
}

// Adds a job requested by the game state.
func (s *WebSocketServer) addJobToGameState(c *websocket.Conn, jsm *pack.JobSubmittedMessage) {
	if s.lobby == nil {
		logger.Warn("[server] Request to add a job was recevied, but no lobby has been created yet.")
		rejectConnection(c)
		return
	}

	if err := jsm.Verify(); err != nil {
		logger.Warnf("[server] Job submission failure: %v", err)
		return
	}

	client, ok := s.lobby.socketsToClients[c]
	if !ok {
		logger.Warnf("[server] Request to add a job was received, but the connecting socket hasn't registered as a player yet!")
		return
	}

	s.gameState.AddJob(client.UUID, jsm.JobInput)
}

// Rejects an incoming connection, responding with a connection rejected message.
func rejectConnection(c *websocket.Conn) {
	connRejMsg := &pack.Message{
		MessageType: pack.ConnectionRejected,
	}

	c.WriteJSON(connRejMsg)
}
