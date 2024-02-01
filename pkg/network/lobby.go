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
	broadcast        chan []byte
	unicastGame      chan []byte
	unicastWeb       chan []byte
	dmSocket         chan *SocketDMRequest
	disconnect       chan *websocket.Conn
}

type SocketDMRequest struct {
	DestSocket *websocket.Conn
	Data       []byte
}

func CreateLobby() *Lobby {
	return &Lobby{
		hostGameClient:   nil,
		webClients:       make(map[*Client]bool),
		socketsToClients: make(map[*websocket.Conn]*Client),
		lobbyCode:        "1234", // todo
		register:         make(chan *Client),
		broadcast:        make(chan []byte),
		unicastGame:      make(chan []byte),
		unicastWeb:       make(chan []byte),
		dmSocket:         make(chan *SocketDMRequest),
		disconnect:       make(chan *websocket.Conn),
	}
}

func CreateSocketDMRequest(c *websocket.Conn, data []byte) *SocketDMRequest {
	return &SocketDMRequest{
		DestSocket: c,
		Data:       data,
	}
}

// Closes the lobby, closing each connected client and the lobby's communication channels.
func (l *Lobby) closeLobby() {
	if l == nil {
		return
	}

	logger.Verbose("[server] Closing lobby.")

	l.hostGameClient.CloseClient()
	for c := range l.webClients {
		c.CloseClient()
	}

	close(l.register)
	close(l.broadcast)
	close(l.unicastGame)
	close(l.unicastWeb)
	close(l.dmSocket)
	close(l.disconnect)
}

// Standard execution of the lobby, goroutine safe.
func (l *Lobby) run() {
	for {
		select {
		// Triggered when a client is registered on the server
		case c := <-l.register:
			l.registerClient(c)
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
		// Triggered when a message needs to be sent to a particular client
		case sdr := <-l.dmSocket:
			l.dmTargetSocket(sdr)
		// Triggered when a client forcefully disconnects from the server, used to end the goroutine
		case <-l.disconnect:
			return
		}
	}
}

// Registers a client on the server.
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

// Registers a game client on the server and responds with the lobby code.
func (l *Lobby) registerGameClient(c *Client) {
	logger.Verbose("[server] Registered a new Game client.")

	lcm := pack.CreateLobbyCodeMessage(&l.lobbyCode)

	// Respond with the lobby code to the game client
	c.conn.WriteJSON(lcm)
}

// Registers a web client and responds with the player's server ID.
func (l *Lobby) registerWebClient(c *Client) {
	logger.Verbose("[server] Registered a new Web client.")

	pidm := pack.CreatePlayerIDMessage(pack.PlayerID, &c.UUID)

	// Respond with the player ID to the web client.
	c.conn.WriteJSON(pidm)
}

// Broadcasts a message to the host game client and all connected clients.
func (l *Lobby) broadcastToClients(msg []byte) {
	l.hostGameClient.conn.WriteMessage(websocket.TextMessage, msg)
	for c := range l.webClients {
		c.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

// Sends a message to the host game client.
func (l *Lobby) unicastToGameClient(msg []byte) {
	l.hostGameClient.conn.WriteMessage(websocket.TextMessage, msg)
}

// Sends a message to all connected web clients.
func (l *Lobby) unicastToWebClients(msg []byte) {
	for c := range l.webClients {
		c.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

// Sends a message directly to a specific socket
func (l *Lobby) dmTargetSocket(sdr *SocketDMRequest) {
	sdr.DestSocket.WriteMessage(websocket.TextMessage, sdr.Data)
}

// Retrieves a client associated with the current socket connection.
func (l *Lobby) GetClientWithSocket(c *websocket.Conn) *Client {
	client, ok := l.socketsToClients[c]

	if !ok {
		logger.Error("[server] Couldn't find a client pertaining to the passed socket.")
		return nil
	}

	return client
}
