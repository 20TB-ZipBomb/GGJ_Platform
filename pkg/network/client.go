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
	ID         uuid.UUID
	Name       string
	lobby      *Lobby
	conn       *websocket.Conn
	send       chan []byte
}

// Creates a game client associated with a particular lobby and connection
func createClient(l *Lobby, c *websocket.Conn, clientType ClientType) *Client {
	uuid, err := uuid.NewRandom()
	if err != nil {
		logger.Errorf("Failed to generate new UUID: %v", err)
	}

	return &Client{
		clientType: clientType,
		ID:         uuid,
		lobby:      l,
		conn:       c,
		send:       make(chan []byte, 256),
	}
}

// Closes a client and it's corresponding websocket connection.
func (c *Client) closeClient() {
	close(c.send)
}
