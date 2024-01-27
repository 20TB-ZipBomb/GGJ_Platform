package network

import "github.com/gorilla/websocket"

type ClientType int

const (
	Web ClientType = iota
	Game
)

type Client struct {
	clientType ClientType
	lobby      *Lobby
	conn       *websocket.Conn
	send       chan []byte
}

// Creates a game client associated with a particular lobby and connection
func CreateGameClient(l *Lobby, c *websocket.Conn) *Client {
	return &Client{
		clientType: Game,
		lobby:      l,
		conn:       c,
		send:       make(chan []byte, 256),
	}
}

// Creates a web client associated with a particular lobby and connection
func CreateWebClient(l *Lobby, c *websocket.Conn) *Client {
	return &Client{
		clientType: Web,
		lobby:      l,
		conn:       c,
		send:       make(chan []byte, 256),
	}
}
