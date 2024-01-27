package ws

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	lobby *Lobby
	conn  *websocket.Conn
	send  chan []byte
}

type Lobby struct {
	players map[*Player]bool
	// todo: Check if these need to change
	broadcast  chan []byte
	register   chan *Player
	unregister chan *Player
}

func newLobby() *Lobby {
	return &Lobby{
		players:    make(map[*Player]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Player),
		unregister: make(chan *Player),
	}
}

func (l *Lobby) run() {
	for {
		select {
		case player := <-l.register:
			l.players[player] = true
		case player := <-l.unregister:
			if _, ok := l.players[player]; ok {
				delete(l.players, player)
				close(player.send)
			}
		case message := <-l.broadcast:
			for player := range l.players {
				select {
				case player.send <- message:
				default:
					close(player.send)
					delete(l.players, player)
				}
			}
		}
	}
}
