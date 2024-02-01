package game

import (
	"math/rand"
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/pack"
)

type ImprovSession struct {
	PlayerQueue  []*PlayerState
	SessionTimer *time.Timer
}

type ImprovSessionTimerCallback func()

// Creates an improv session using a list of players, these players are shuffled and placed in a queue.
func CreateImprovSession(players []*PlayerState) *ImprovSession {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})

	return &ImprovSession{
		PlayerQueue: players,
	}
}

// Retrieves the number of players left to participate in the improv round.
func (is *ImprovSession) GetNumberOfPlayersLeftToImprov() int {
	return len(is.PlayerQueue)
}

// Retrieves the player that is currently presenting (the player at the front of the queue)
func (is *ImprovSession) GetCurrentImprovPlayer() *PlayerState {
	if is.PlayerQueue == nil {
		return nil
	}

	return is.PlayerQueue[0]
}

// Retrieves the number of scores submitted for the currently improv'ing player.
func (is *ImprovSession) GetNumberOfScoresSubmittedForCurrentPlayer() int {
	if is.GetCurrentImprovPlayer() == nil {
		logger.Error("[server] Failed to retrieve the number of scores submitted for the current player, no players are on the queue!")
		return -1
	}

	return is.GetCurrentImprovPlayer().NumberOfScoresSubmitted
}

// Pops the top player off the improv queue.
func (is *ImprovSession) PopPlayerOnQueue() *PlayerState {
	if is.PlayerQueue == nil || len(is.PlayerQueue) == 0 {
		return nil
	}

	poppedPlayer := is.PlayerQueue[0]
	is.PlayerQueue = is.PlayerQueue[1:]

	return poppedPlayer
}

// Starts a timer for the current improv session.
func (is *ImprovSession) StartTimerForSession(cb ImprovSessionTimerCallback) {
	if is.SessionTimer != nil {
		logger.Warn("Tried to start a new round of improv but the timer is already going, ignoring.")
		return
	}

	t := Config.GetTypedImprovRoundDurationSeconds()
	logger.Debugf("TIMER: %s", t.String())
	is.SessionTimer = time.NewTimer(t)
	<-is.SessionTimer.C

	cb()

	is.SessionTimer = nil
}

// Resets the improv timer to a time, used during interceptions.
func (is *ImprovSession) ResetSessionTimer(resetTime time.Duration) {
	is.SessionTimer.Reset(resetTime)
}

// Applies a score submission message's data to this player's stats.
func (is *ImprovSession) SubmitScoreForPlayer(ss *pack.ScoreSubmissionMessage) {
	player := is.GetCurrentImprovPlayer()
	player.ScoreInCents += ss.ScoreInCents
	player.NumberOfScoresSubmitted += 1
}
