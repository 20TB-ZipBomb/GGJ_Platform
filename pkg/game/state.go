package game

import (
	"math/rand"
	"strings"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/pack"
	"github.com/google/uuid"
)

// Maintains the state of the game on the server.
type State struct {
	JobPool                []*pack.Card
	JobInputsPerPlayer     int
	PlayersToSubmittedJobs map[uuid.UUID][]*pack.Card
	PlayersToDealtJobs     map[uuid.UUID][]*pack.Card
	PlayersToPlayerState   map[uuid.UUID]*PlayerState
	PlayerImprovOrder      []*PlayerState
}

type PlayerState struct {
	UUID         uuid.UUID
	DrawnCards   []*pack.Card
	JobCard      *pack.Card
	SelectedCard *pack.Card
}

// Initializes the game state with the current number of players extracted from a list of their UUIDs.
func CreateGameState(uuids []uuid.UUID) *State {
	numPlayers := len(uuids)

	// Players are required to come up with N+1 jobs
	numRequiredJobInputs := numPlayers + 1

	s := &State{
		JobPool:                make([]*pack.Card, 0),
		JobInputsPerPlayer:     numRequiredJobInputs,
		PlayersToSubmittedJobs: make(map[uuid.UUID][]*pack.Card),
		PlayersToDealtJobs:     make(map[uuid.UUID][]*pack.Card),
		PlayersToPlayerState:   make(map[uuid.UUID]*PlayerState),
		PlayerImprovOrder:      make([]*PlayerState, 0),
	}

	// Construct the array of jobs for each connected UUID
	// This is why we need that list of UUIDs
	//
	// (╯°□°)╯︵ ┻━┻
	for _, uuid := range uuids {
		if _, ok := s.PlayersToSubmittedJobs[uuid]; !ok {
			s.PlayersToSubmittedJobs[uuid] = make([]*pack.Card, 0)
		}

		if _, ok := s.PlayersToDealtJobs[uuid]; !ok {
			s.PlayersToDealtJobs[uuid] = make([]*pack.Card, 0)
		}
	}

	return s
}

// Creates a player state with UUID and stores it inside the game state.
func (s *State) CreatePlayerStateWithUUID(uuid uuid.UUID, drawnCards []*pack.Card, jobCard *pack.Card) {
	ps := &PlayerState{
		UUID:         uuid,
		DrawnCards:   drawnCards,
		JobCard:      jobCard,
		SelectedCard: nil,
	}

	s.PlayersToPlayerState[uuid] = ps
	s.PlayerImprovOrder = append(s.PlayerImprovOrder, ps)
}

// Resets the current game state.
func (s *State) Reset() {
	if s == nil {
		return
	}

	s.JobPool = make([]*pack.Card, 0)
	s.JobInputsPerPlayer = 0
	s.PlayersToSubmittedJobs = make(map[uuid.UUID][]*pack.Card)
	s.PlayersToDealtJobs = make(map[uuid.UUID][]*pack.Card)
	s.PlayersToPlayerState = make(map[uuid.UUID]*PlayerState)
	s.PlayerImprovOrder = make([]*PlayerState, 0)
}

// Converts the current job pool array to a string.
func (s *State) JobPoolString() string {
	valJobPool := make([]string, 0)

	for _, job := range s.JobPool {
		valJobPool = append(valJobPool, *job.JobText)
	}

	return "[ " + strings.Join(valJobPool, ", ") + " ]"
}

// Prints a pretty format for the jobs submitted by each connected client.
func (s *State) JobUUIDMapToString(jobMap *map[uuid.UUID][]*pack.Card) string {
	out := "\n\n"

	for uuid, jobs := range *jobMap {
		valJobPool := make([]string, 0)
		out = out + uuid.String() + " ~> [ "

		for _, job := range jobs {
			valJobPool = append(valJobPool, *job.JobText)
		}

		out = out + strings.Join(valJobPool, ", ")
		out = out + " ]\n"
	}

	return out
}

// Checks if the user with the provided UUID has finished submitting jobs.
func (s *State) HasUserFinishedSubmittingJobs(uuid uuid.UUID) bool {
	numJobsSubmitted := len(s.PlayersToSubmittedJobs[uuid])
	return numJobsSubmitted == s.JobInputsPerPlayer
}

// Checks if all players have finished submitting jobs.
func (s *State) HaveAllUsersFinishedSubmittingJobs() bool {
	for uuid := range s.PlayersToSubmittedJobs {
		if !s.HasUserFinishedSubmittingJobs(uuid) {
			return false
		}
	}

	return true
}

// Checks if all players have selected a job for improv.
func (s *State) HaveAllUsersSelectedAJobForImprov() bool {
	for _, ps := range s.PlayersToPlayerState {
		if ps.SelectedCard == nil {
			logger.Debugf("the person who had %s has not selected", *ps.JobCard)
			return false
		}
	}

	return true
}

// Adds a job to the list of jobs for a user with the passed UUID.
func (s *State) AddJob(targetUUID uuid.UUID, sj *string) {
	// Early out if this user has already submitted their required jobs
	if s.HasUserFinishedSubmittingJobs(targetUUID) {
		return
	}

	// Create a card using a new random UUID
	newUUID, err := uuid.NewRandom()
	if err != nil {
		logger.Errorf("Failed to generate new UUID: %v", err)
	}

	card := &pack.Card{
		CardID:  newUUID,
		JobText: sj,
	}

	// Append job to the pool of available jobs, as well as the jobs for this user
	s.PlayersToSubmittedJobs[targetUUID] = append(s.PlayersToSubmittedJobs[targetUUID], card)
	s.JobPool = append(s.JobPool, card)

	// Display a helper string for jobs being added outside of production
	logger.Debugf("%s", s.JobUUIDMapToString(&s.PlayersToSubmittedJobs))
}

// Deals jobs to players.
func (s *State) DealJobsToPlayers() {
	rand.Shuffle(len(s.JobPool), func(i, j int) { s.JobPool[i], s.JobPool[j] = s.JobPool[j], s.JobPool[i] })
	logger.Verbosef("Shuffled JobList: %s", s.JobPoolString())

	// Each player gets drawn N cards (where N is the number of clients connected)
	// Each player is assigned one job card that they are applying for
	numPlayers := len(s.PlayersToSubmittedJobs)
	size := (numPlayers + 1)
	evenSplits := make([][]*pack.Card, 0)
	var j int
	for i := 0; i < len(s.JobPool); i += size {
		j += size
		if j > len(s.JobPool) {
			j = len(s.JobPool)
		}

		evenSplits = append(evenSplits, s.JobPool[i:j])
	}

	i := 0
	for uuid := range s.PlayersToDealtJobs {
		s.PlayersToDealtJobs[uuid] = evenSplits[i]
		i++
	}

	logger.Debugf("%s", s.JobUUIDMapToString(&s.PlayersToDealtJobs))
}

// Retrieves the next player for improv.
func (s *State) GetNextPlayerForImprov() *PlayerState {
	// Shuffle the ordering
	rand.Shuffle(len(s.PlayerImprovOrder), func(i, j int) {
		s.PlayerImprovOrder[i], s.PlayerImprovOrder[j] = s.PlayerImprovOrder[j], s.PlayerImprovOrder[i]
	})

	for _, ps := range s.PlayerImprovOrder {
		logger.Verbosef("%s", ps.UUID.String())
	}

	// Pop the first item from slice
	newPlayer := s.PlayerImprovOrder[:1]
	s.PlayerImprovOrder = s.PlayerImprovOrder[1:]

	return newPlayer[0]
}
