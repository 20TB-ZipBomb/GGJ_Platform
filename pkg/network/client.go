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
