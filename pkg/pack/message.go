package pack

import (
	"errors"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils/json"
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
	PlayerImprovStart                 = "player_improv_start"
	CardData                          = "card_data"
	InterceptionCardData              = "intercept_card_data"
	TimerFinished                     = "timer_finished"
	ScoreSubmission                   = "score_submission"
	GameFinished                      = "game_finished"
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
	Name *string `json:"name"`
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

// Message sent from web clients indicating that a player has selected their card.
// Web -> Server
type CardDataMessage struct {
	Message
	Card *Card
}

// Message sent to the game client indicating that a player has started improv.
// Server -> Game
type PlayerImprovStartMessage struct {
	PlayerIDMessage
	SelectedCard  *Card `json:"selected_card"`
	JobCard       *Card `json:"job_card"`
	TimeInSeconds int   `json:"time_in_seconds"`
}

// Message sent to and from clients to represent the submission of salary scores in cents
// Web -> Server / Server -> Game
type ScoreSubmissionMessage struct {
	Message
	ScoreInCents int `json:"score_in_cents"`
}

// Message sent from the server to game clients to represent a card interception being played.
// Server -> Game
type InterceptionCardMessage struct {
	PlayerIDMessage
	InterceptedCard *Card `json:"intercepted_card"`
	TimeInSeconds   int   `json:"time_in_seconds"`
}

// Creates a Message.
func CreateBasicMessage(mt MessageType) *Message {
	return &Message{
		MessageType: mt,
	}
}

// Creates and marshals a Message.
func MarshalBasicMessage(mt MessageType) []byte {
	return json.MarshalJSONBytes[Message](CreateBasicMessage(mt))
}

// Creates a LobbyCodeMessage.
func CreateLobbyCodeMessage(lc *string) *LobbyCodeMessage {
	return &LobbyCodeMessage{
		Message:   *CreateBasicMessage(LobbyCode),
		LobbyCode: lc,
	}
}

// Verifies the integrity of the `LobbyJoinAttemptMessage`, reports errors as required
func (l *LobbyJoinAttemptMessage) Verify(lc *string) error {
	if l.LobbyCode == nil {
		return errors.New("Lobby join request was received, but no lobby code was specified.")
	}

	if l.Name == nil {
		return errors.New("Lobby join request was received, but no player name was specified.")
	}

	if *l.LobbyCode != *lc {
		return errors.New("Lobby join request was received, but the lobby code was incorrect.")
	}

	return nil
}

// Creates a PlayerIDMessage.
func CreatePlayerIDMessage(mt MessageType, uuid *uuid.UUID) *PlayerIDMessage {
	return &PlayerIDMessage{
		Message:  *CreateBasicMessage(mt),
		PlayerID: *uuid,
	}
}

// Creates and marshals a PlayerIDMessage.
func MarshalPlayerIDMessage(mt MessageType, uuid *uuid.UUID) []byte {
	return json.MarshalJSONBytes[PlayerIDMessage](CreatePlayerIDMessage(mt, uuid))
}

// Creates a Player.
func CreatePlayer(uuid *uuid.UUID, name *string) *Player {
	return &Player{
		PlayerID: *uuid,
		Name:     name,
	}
}

// Creates a PlayerJoinedMessage.
func CreatePlayerJoinedMessage(uuid *uuid.UUID, name *string) *PlayerJoinedMessage {
	return &PlayerJoinedMessage{
		Message: *CreateBasicMessage(PlayerJoined),
		Player:  *CreatePlayer(uuid, name),
	}
}

// Creates a GameStartMessage.
func CreateGameStartMessage(n int) *GameStartMessage {
	return &GameStartMessage{
		Message:      *CreateBasicMessage(GameStart),
		NumberOfJobs: n,
	}
}

// Verifies the integrity of the `JobSubmittedMessage`, reports errors as required
func (j *JobSubmittedMessage) Verify() error {
	if j.JobInput == nil {
		return errors.New("Job submission request was received, but no job was specified.")
	}

	return nil
}

// Creates a ReceivedCardsMessage.
func CreateReceivedCardsMessage(dc []*Card, jc *Card) *ReceivedCardsMessage {
	return &ReceivedCardsMessage{
		Message:    *CreateBasicMessage(ReceivedCards),
		DrawnCards: dc,
		JobCard:    jc,
	}
}

// Creates and marshals a ReceivedCardsMessage.
func MarshalReceivedCardsMessage(dc []*Card, jc *Card) []byte {
	return json.MarshalJSONBytes[ReceivedCardsMessage](CreateReceivedCardsMessage(dc, jc))
}

// Verifies the integrity of the `CardDataMessage`, reports errors as required.
func (c *CardDataMessage) Verify() error {
	if c.Card == nil {
		return errors.New("Card submission request was received, but no card was specfied.")
	}

	if err := uuid.Validate(c.Card.CardID.String()); err != nil {
		return errors.New("Card submission request was receieved, but the card had a malformed UUID.")
	}

	if c.Card.JobText == nil {
		return errors.New("Card submission request was received, but the card had malformed text.")
	}

	return nil
}

// Creates a PlayerImprovStartMessage.
func CreatePlayerImprovStartMessage(uuid *uuid.UUID, sc *Card, jc *Card, t int) *PlayerImprovStartMessage {
	return &PlayerImprovStartMessage{
		PlayerIDMessage: *CreatePlayerIDMessage(PlayerImprovStart, uuid),
		SelectedCard:    sc,
		JobCard:         jc,
		TimeInSeconds:   t,
	}
}

// Creates and marshals a PlayerImprovStartMessage.
func MarshalPlayerImprovStartMessage(uuid *uuid.UUID, sc *Card, jc *Card, t int) []byte {
	return json.MarshalJSONBytes[PlayerImprovStartMessage](CreatePlayerImprovStartMessage(uuid, sc, jc, t))
}

// Creates a ScoreSubmissionMessage.
func CreateScoreSubmissionMessage(sc int) *ScoreSubmissionMessage {
	return &ScoreSubmissionMessage{
		Message:      *CreateBasicMessage(ScoreSubmission),
		ScoreInCents: sc,
	}
}

// Creates and marshals a ScoreSubmissionMessage.
func MarshalScoreSubmissionMessage(sc int) []byte {
	return json.MarshalJSONBytes[ScoreSubmissionMessage](CreateScoreSubmissionMessage(sc))
}

// Creates an InterceptionCardMessage.
func CreateInterceptionCardMessage(uuid *uuid.UUID, c *Card, t int) *InterceptionCardMessage {
	return &InterceptionCardMessage{
		PlayerIDMessage: *CreatePlayerIDMessage(InterceptionCardData, uuid),
		InterceptedCard: c,
		TimeInSeconds:   t,
	}
}

// Creates and marshals an InterceptionCardMessage,
func MarshalInterceptionCardMessage(uuid *uuid.UUID, c *Card, t int) []byte {
	return json.MarshalJSONBytes[InterceptionCardMessage](CreateInterceptionCardMessage(uuid, c, t))
}
