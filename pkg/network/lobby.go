package network

import "github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"

type Lobby struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

func newLobby() *Lobby {
	return &Lobby{
		clients:    make(map[*Client]bool),
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
			logger.Info("Registered a new client.")
			l.clients[client] = true
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
