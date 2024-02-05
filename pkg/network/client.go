package network

import (
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/pack"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ClientType int

const (
	Web ClientType = iota
	Game
)

const (
	alivePingTimeoutSeconds = 45 * time.Second
)

type Client struct {
	clientType ClientType
	UUID       uuid.UUID
	Name       string
	lobby      *Lobby
	conn       *websocket.Conn
	pingTimer  *time.Timer
}

// Creates a game client associated with a particular lobby and connection
func CreateClient(l *Lobby, c *websocket.Conn, clientType ClientType) *Client {
	uuid, err := uuid.NewRandom()
	if err != nil {
		logger.Errorf("Failed to generate new UUID: %v", err)
	}

	cl := &Client{
		clientType: clientType,
		UUID:       uuid,
		lobby:      l,
		conn:       c,
	}

	// Start a goroutine that pumps ping messages to this client.
	// This ensures the websocket connection stays alive, as, some services
	// such as Heroku will auto disconnect sockets if they remain idle.
	go cl.startAlivePingPump()

	return cl
}

// Closes a client and it's corresponding websocket connection.
func (c *Client) CloseClient() {
	if c != nil {
		c.conn.Close()
	}
}

// Creates a pump that broadcasts a ping to all clients in the lobby in a fixed duration.
func (c *Client) startAlivePingPump() {
	for {
		c.pingTimer = time.NewTimer(alivePingTimeoutSeconds)
		<-c.pingTimer.C

		aliveData := pack.MarshalBasicMessage(pack.Alive)
		c.lobby.dmSocket <- CreateSocketDMRequest(c.conn, aliveData)
	}
}
