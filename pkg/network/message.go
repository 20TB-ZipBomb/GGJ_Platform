package network

type MessageType string

const (
	ConnectionRejected MessageType = "connection_rejected"
	CreateLobby                    = "create_lobby"
	LobbyCode                      = "lobby_code"
	LobbyJoinAttempt               = "lobby_join_attempt"
	PlayerJoined                   = "player_joined"
)

// Generic communication message containing a message type
type Message struct {
	MessageType MessageType `json:"message_type"`
}

// Message containing a lobby code to send to game clients
// Server -> Game Client
type LobbyCodeMessage struct {
	Message
	LobbyCode string `json:"lobby_code"`
}

// Message containing information for web clients attempting to join a lobby
// Web Client -> Server
type LobbyJoinAttemptMessage struct {
	LobbyCodeMessage
	PlayerName string `json:"player_name"`
}

// todo: Move this somewhere else
type Player struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
}

// Message containing information sent back to a player once they connect to the server
// Server -> Web Client
type PlayerJoinedMessage struct {
	Message
	Player Player `json:"player"`
}
