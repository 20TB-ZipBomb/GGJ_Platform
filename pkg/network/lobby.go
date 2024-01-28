package network

import (
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/pack"
	"github.com/gorilla/websocket"
)

type Lobby struct {
	hostGameClient   *Client
	webClients       map[*Client]bool
	socketsToClients map[*websocket.Conn]*Client
	lobbyCode        string
	register         chan *Client
	unregister       chan *Client
	broadcast        chan []byte
	unicastGame      chan []byte
	unicastWeb       chan []byte
	disconnect       chan *websocket.Conn
}

// Creates the lobby and it's communication channels.
func createLobby() *Lobby {
	return &Lobby{
		hostGameClient:   nil,
		webClients:       make(map[*Client]bool),
		socketsToClients: make(map[*websocket.Conn]*Client),
		lobbyCode:        "1234", // todo
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		broadcast:        make(chan []byte),
		unicastGame:      make(chan []byte),
		unicastWeb:       make(chan []byte),
		disconnect:       make(chan *websocket.Conn),
	}
}

// Closes the lobby, closing each connected client and the lobby's communication channels.
func (l *Lobby) closeLobby() {
	if l == nil {
		return
	}

	logger.Verbose("[server] Closing lobby.")

	l.hostGameClient.closeClient()
	for c := range l.webClients {
		c.closeClient()
	}

	close(l.register)
	close(l.unregister)
	close(l.broadcast)
	close(l.unicastGame)
	close(l.unicastWeb)
	close(l.disconnect)
}

// Standard execution of the lobby, goroutine safe.
func (l *Lobby) run() {
	for {
		select {
		// Triggered when a client is registered on the server
		case c := <-l.register:
			l.registerClient(c)
		// Triggered when a client is unregistered from the server
		case c := <-l.unregister:
			l.unregisterClient(c)
		// Triggered when a message is broadcasted to all clients
		case msg := <-l.broadcast:
			l.broadcastToClients(msg)
		// Triggered when a message is unicasted to the game client
		case msg := <-l.unicastGame:
			l.unicastToGameClient(msg)
		// Triggered when a message is unicasted to all web clients
		// (really just a broadcast without the hosting client, but it works)
		case msg := <-l.unicastWeb:
			l.unicastToWebClients(msg)
		// Triggered when a client forcefully disconnects from the server
		case <-l.disconnect:
			return
		}
	}
}

// Registers a client.
func (l *Lobby) registerClient(c *Client) {
	// Map the socket to this client for reverse-lookup later
	l.socketsToClients[c.conn] = c

	if c.clientType == Game {
		l.hostGameClient = c
		l.registerGameClient(c)
	} else if c.clientType == Web {
		l.webClients[c] = true
		l.registerWebClient(c)
	} else {
		panic("Unknown client type")
	}
}

// Registers a game client and responds with the lobby code.
func (l *Lobby) registerGameClient(c *Client) {
	logger.Verbose("[server] Registered a new Game client.")

	lcm := &pack.LobbyCodeMessage{
		Message: pack.Message{
			MessageType: pack.LobbyCode,
		},
		LobbyCode: &l.lobbyCode,
	}

	// Respond with the lobby code to the game client
	c.conn.WriteJSON(lcm)
}

// Registers a web client and responds with the player's server ID.
func (l *Lobby) registerWebClient(c *Client) {
	logger.Verbose("[server] Registered a new Web client.")

	pidm := &pack.PlayerIDMessage{
		Message: pack.Message{
			MessageType: pack.PlayerID,
		},
		PlayerID: c.UUID,
	}

	// Respond with the player ID to the web client.
	c.conn.WriteJSON(pidm)
}

// Unregister a client using their connected socket
func (l *Lobby) unregisterClientWithSocket(c *websocket.Conn) bool {
	if client, ok := l.socketsToClients[c]; ok {
		l.unregisterClient(client)
		return true
	}

	return false
}

// Unregisters a client from the server and closes it's relevant connections.
// If the host game client is unregistered, the entire lobby is closed.
func (l *Lobby) unregisterClient(c *Client) {
	if _, ok := l.webClients[c]; ok {
		logger.Verbosef("[server] Closing web client.")
		delete(l.webClients, c)
		c.closeClient()
	} else if c == l.hostGameClient {
		logger.Verbosef("[server] Closing game client.")
		l.closeLobby()
	}
}

// Broadcasts a message to the host game client and all connected clients.
func (l *Lobby) broadcastToClients(msg []byte) {
	l.hostGameClient.conn.WriteMessage(websocket.BinaryMessage, msg)
	for c := range l.webClients {
		c.conn.WriteMessage(websocket.BinaryMessage, msg)
	}
}

// Sends a message to the host game client.
func (l *Lobby) unicastToGameClient(msg []byte) {
	l.hostGameClient.conn.WriteMessage(websocket.BinaryMessage, msg)
}

// Sends a message to all connected web clients.
func (l *Lobby) unicastToWebClients(msg []byte) {
	for c := range l.webClients {
		c.conn.WriteMessage(websocket.BinaryMessage, msg)
	}
}
