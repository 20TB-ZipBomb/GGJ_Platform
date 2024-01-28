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
	"github.com/google/uuid"
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
		if end := strMsg[len(strMsg):]; end == "\n" {
			logger.Verbosef("[payload] %s", strMsg[:len(strMsg)-1])
		} else {
			logger.Verbosef("[payload] %s", strMsg)
		}

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
		case pack.CardData:
			cd := json.UnmarshalJSON[pack.CardDataMessage](msg)
			s.submitCardToGameState(c, cd)
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
	client.Name = *ljam.Name
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

	clients := s.lobby.webClients
	// todo: Remove production environment constraint for minimum number of players?
	if utils.IsProductionEnv() && len(clients) < minimumNumberOfPlayers {
		logger.Warn("Start game request was received, but the lobby has less than the minimum amount of clients connected that are required to play.")
		rejectConnection(c)
		return
	}

	// This has to be initialized with a list of UUIDs to properly setup the game (for now)
	// Note: If we want to have drop-in/drop-out play later, we'd have to change this
	uuids := make([]uuid.UUID, 0)
	for client := range s.lobby.webClients {
		uuids = append(uuids, client.UUID)
	}
	s.gameState = game.CreateGameState(uuids)

	sgm := &pack.GameStartMessage{
		Message: pack.Message{
			MessageType: pack.GameStart,
		},
		NumberOfJobs: s.gameState.JobInputsPerPlayer,
	}
	s.lobby.broadcast <- []byte(json.MarshalJSON[pack.GameStartMessage](sgm))
}

// Some basic pre-requisites to check before executing game state commands
func (s *WebSocketServer) doesPassPreRequisites(c *websocket.Conn) bool {
	if s.lobby == nil {
		logger.Warn("[server] Request to add a job was recevied, but no lobby has been created yet.")
		rejectConnection(c)
		return false
	}

	_, ok := s.lobby.socketsToClients[c]
	if !ok {
		logger.Warnf("[server] Request to add a job was received, but the connecting socket hasn't registered as a player yet!")
		return false
	}

	if s.gameState == nil {
		logger.Warn("[server] Request to add a job was received, but the game hasn't started yet!")
		return false
	}

	return true
}

// Adds a job requested by the game state.
// This also deals out cards to players once they've all submitted as a side effect.
// todo: Refactor this?
func (s *WebSocketServer) addJobToGameState(c *websocket.Conn, jsm *pack.JobSubmittedMessage) {
	if !s.doesPassPreRequisites(c) {
		return
	}

	if err := jsm.Verify(); err != nil {
		logger.Warnf("[server] Job submission failure: %v", err)
		return
	}

	client, _ := s.lobby.socketsToClients[c]
	s.gameState.AddJob(client.UUID, jsm.JobInput)

	// Once the player has submitted the maximum number of jobs, send infomation to the game client
	if s.gameState.HasUserFinishedSubmittingJobs(client.UUID) {
		pid := &pack.PlayerIDMessage{
			Message: pack.Message{
				MessageType: pack.JobSubmittingFinished,
			},
			PlayerID: client.UUID,
		}

		client.lobby.unicastGame <- []byte(json.MarshalJSON[pack.PlayerIDMessage](pid))
	}

	// Once all players have finished submitting jobs
	if s.gameState.HaveAllUsersFinishedSubmittingJobs() {
		logger.Debug("All users have submitted jobs!")

		// Send a message to the game indicating that players are now receiving their cards
		go func() {
			rcmGame := pack.Message{
				MessageType: pack.ReceivedCards,
			}

			client.lobby.unicastGame <- []byte(json.MarshalJSON[pack.Message](&rcmGame))
		}()

		// Send a message to the web indicating that players are receiving shuffled job cards
		go func() {
			s.gameState.DealJobsToPlayers()

			for cl := range s.lobby.webClients {
				uuidCards := s.gameState.PlayersToSubmittedJobs[cl.UUID]
				drawnCards := uuidCards[0:(len(uuidCards) - 1)]
				jobCard := uuidCards[len(uuidCards)-1]

				// Set the player state inside the game state
				s.gameState.CreatePlayerStateWithUUID(cl.UUID, drawnCards, jobCard)

				rcmWeb := pack.ReceivedCardsMessage{
					Message: pack.Message{
						MessageType: pack.ReceivedCards,
					},
					DrawnCards: drawnCards,
					JobCard:    jobCard,
				}

				rcmData := []byte(json.MarshalJSON[pack.ReceivedCardsMessage](&rcmWeb))
				cl.lobby.dmSocket <- &SocketDMRequest{
					DestSocket: cl.conn,
					Data:       rcmData,
				}
			}
		}()
	}
}

func (s *WebSocketServer) submitCardToGameState(c *websocket.Conn, cd pack.CardDataMessage) {
	if !s.doesPassPreRequisites(c) {
		return
	}

	if err := cd.Verify(); err != nil {
		logger.Warnf("[server] Card submission failure: %v", err)
		return
	}

	// Send data back to the game client that this player has selected a role for improv
	client, _ := s.lobby.socketsToClients[c]
	if ps, ok := s.gameState.PlayersToPlayerState[client.UUID]; ok {
		ps.SelectedCard = cd.Card

		pid := pack.PlayerIDMessage{
			Message: pack.Message{
				MessageType: pack.CardData,
			},
			PlayerID: client.UUID,
		}

		client.lobby.unicastGame <- []byte(json.MarshalJSON[pack.PlayerIDMessage](&pid))
	}

	// Check if all players have selected a job for improv
	if s.gameState.HaveAllUsersSelectedAJobForImprov() {
		ps := s.gameState.GetNextPlayerForImprov()

		// Send an improv start message to the game
		pism := pack.PlayerImprovStartMessage{
			PlayerIDMessage: pack.PlayerIDMessage{
				Message: pack.Message{
					MessageType: pack.PlayerID,
				},
				PlayerID: ps.UUID,
			},
			SelectedCard:  ps.SelectedCard,
			JobCard:       ps.JobCard,
			TimeInSeconds: 30,
		}

		client.lobby.unicastGame <- []byte(json.MarshalJSON[pack.PlayerImprovStartMessage](&pism))

		// Send a generic PlayerID to the web client
		pidm := pack.PlayerIDMessage{
			Message: pack.Message{
				MessageType: pack.PlayerID,
			},
			PlayerID: ps.UUID,
		}

		client.lobby.unicastWeb <- []byte(json.MarshalJSON[pack.PlayerIDMessage](&pidm))
	}
}

// Rejects an incoming connection, responding with a connection rejected message.
func rejectConnection(c *websocket.Conn) {
	connRejMsg := &pack.Message{
		MessageType: pack.ConnectionRejected,
	}

	c.WriteJSON(connRejMsg)
}
