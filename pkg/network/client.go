package network

import (
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ClientType int

const (
	Web ClientType = iota
	Game
)

type Client struct {
	clientType ClientType
	UUID       uuid.UUID
	Name       string
	lobby      *Lobby
	conn       *websocket.Conn
}

// Creates a game client associated with a particular lobby and connection
func CreateClient(l *Lobby, c *websocket.Conn, clientType ClientType) *Client {
	uuid, err := uuid.NewRandom()
	if err != nil {
		logger.Errorf("Failed to generate new UUID: %v", err)
	}

	return &Client{
		clientType: clientType,
		UUID:       uuid,
		lobby:      l,
		conn:       c,
	}
}

// Closes a client and it's corresponding websocket connection.
func (c *Client) CloseClient() {
	if c != nil {
		c.conn.Close()
	}
}
