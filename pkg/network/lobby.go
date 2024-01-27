package network

import (
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils/json"
)

type Lobby struct {
	clients    map[*Client]bool
	lobbyCode  string
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

func newLobby() *Lobby {
	return &Lobby{
		clients:    make(map[*Client]bool),
		lobbyCode:  "1234", // todo
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

func (l *Lobby) run() {
	for {
		select {
		// Triggered when a client is registered on the server
		case client := <-l.register:
			l.registerClient(client)
		// Triggered when a client is unregistered from the server
		case client := <-l.unregister:
			if _, ok := l.clients[client]; ok {
				delete(l.clients, client)
				close(client.send)
			}
		// Triggered when a message is broadcasted to server clients
		case message := <-l.broadcast:
			for client := range l.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(l.clients, client)
				}
			}
		}
	}
}

func (l *Lobby) registerClient(c *Client) {
	l.clients[c] = true

	if c.clientType == Game {
		l.registerGameClient(c)
	} else if c.clientType == Web {
		l.registerWebClient(c)
	} else {
		panic("Unknown client type")
	}
}

func (l *Lobby) registerGameClient(c *Client) {
	logger.Info("Registered a new Game client.")

	lobbyCodeMessage := &LobbyCodeMessage{
		Message: Message{
			MessageType: LobbyCode,
		},
		LobbyCode: l.lobbyCode,
	}

	// When the game client connects send data back with the lobby code
	c.conn.WriteJSON(json.MarshalJSON[LobbyCodeMessage](lobbyCodeMessage))
}

func (l *Lobby) registerWebClient(c *Client) {
	logger.Info("Registered a new Web client.")

	playerJoinedMessage := &PlayerJoinedMessage{
		Message: Message{
			MessageType: PlayerJoined,
		},
		Player: Player{
			PlayerID: "9999", // todo
			Name:     "Jake",
		},
	}

	// When a new web client is registered send data back to the client
	c.conn.WriteJSON(json.MarshalJSON[PlayerJoinedMessage](playerJoinedMessage))
}
