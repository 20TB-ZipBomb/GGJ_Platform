package pack

import (
	"errors"

	"github.com/google/uuid"
)

type MessageType string

const (
	ConnectionRejected    MessageType = "connection_rejected"
	CreateLobby                       = "create_lobby"
	LobbyCode                         = "lobby_code"
	LobbyJoinAttempt                  = "lobby_join_attempt"
	PlayerID                          = "player_id"
	PlayerJoined                      = "player_joined"
	GameStart                         = "game_start"
	JobSubmitted                      = "job_submitted"
	JobSubmittingFinished             = "player_job_submitting_finished"
	ReceivedCards                     = "received_cards"
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

// Represents a player with a UUID and a name.
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

// Message that acknowledges the start of the game for both web and game clients
// Server -> Web
// Server -> Game
type GameStartMessage struct {
	Message
	NumberOfJobs int `json:"number_of_jobs"`
}

// Message containing the information sent by web clients for submitted jobs.
// Web -> Server
type JobSubmittedMessage struct {
	Message
	JobInput *string `json:"job_input"`
}

// Represents a job provided to players, contains a UUID and text.
type Card struct {
	CardID  uuid.UUID `json:"card_id"`
	JobText *string   `json:"job_text"`
}

// Message sent to web clients containing information about the hand they drew and their target job.
// Server -> Web
type ReceivedCardsMessage struct {
	Message
	DrawnCards []*Card `json:"drawn_cards"`
	JobCard    *Card   `json:"job_card"`
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
