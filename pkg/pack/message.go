package pack

import (
	"errors"

	"github.com/google/uuid"
)

type MessageType string

const (
	ConnectionRejected MessageType = "connection_rejected"
	CreateLobby                    = "create_lobby"
	LobbyCode                      = "lobby_code"
	LobbyJoinAttempt               = "lobby_join_attempt"
	PlayerID                       = "player_id"
	PlayerJoined                   = "player_joined"
	GameStart                      = "game_start"
	JobSubmitted                   = "job_submitted"
)

// Generic communication message containing a message type
type Message struct {
	MessageType MessageType `json:"message_type"`
}

// Message containing a lobby code to send to game clients.
// Server -> Game
type LobbyCodeMessage struct {
	Message
	LobbyCode *string `json:"lobby_code"`
}

// Message containing information for web clients attempting to join a lobby.
// Web -> Server
type LobbyJoinAttemptMessage struct {
	LobbyCodeMessage
	PlayerName *string `json:"player_name"`
}

// Message containing the connected player's ID.
// Server -> Web
type PlayerIDMessage struct {
	Message
	PlayerID uuid.UUID `json:"player_id"`
}

// todo: Move this somewhere else
type Player struct {
	PlayerID uuid.UUID `json:"player_id"`
	Name     *string   `json:"name"`
}

// Message containing information sent back to a player once they connect to the server.
// Server -> Game
type PlayerJoinedMessage struct {
	Message
	Player Player `json:"player"`
}

type JobSubmittedMessage struct {
	Message
	JobInput *string `json:"job_input"`
}

// Verifies the integrity of the `LobbyJoinAttemptMessage`, reports errors as required
func (l *LobbyJoinAttemptMessage) Verify(lc *string) error {
	if l.LobbyCode == nil {
		return errors.New("Lobby join request was received, but no lobby code was specified.")
	}

	if l.PlayerName == nil {
		return errors.New("Lobby join request was received, but no player name was specified.")
	}

	if *l.LobbyCode != *lc {
		return errors.New("Lobby join request was received, but the lobby code was incorrect.")
	}

	return nil
}

// Verifies the integrity of the `JobSubmittedMessage`, reports errors as required
func (j *JobSubmittedMessage) Verify() error {
	if j.JobInput == nil {
		return errors.New("Job submission request was received, but no job was specified.")
	}

	return nil
}
