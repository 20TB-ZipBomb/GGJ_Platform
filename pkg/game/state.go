package game

import (
	"strings"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils"
	"github.com/google/uuid"
)

type State struct {
	JobPool                []*string
	JobInputsPerPlayer     int
	PlayersToSubmittedJobs map[uuid.UUID][]*string
}

func CreateGameState(numPlayers int) *State {
	// Players are required to come up with N+1 jobs
	numRequiredJobInputs := numPlayers + 1

	return &State{
		JobPool:                make([]*string, 0),
		JobInputsPerPlayer:     numRequiredJobInputs,
		PlayersToSubmittedJobs: make(map[uuid.UUID][]*string),
	}
}

func (s *State) Reset() {
	s.JobPool = make([]*string, 0)
	s.JobInputsPerPlayer = 0
	s.PlayersToSubmittedJobs = make(map[uuid.UUID][]*string)
}

func (s *State) AddJob(uuid uuid.UUID, sj *string) {
	if _, ok := s.PlayersToSubmittedJobs[uuid]; !ok {
		s.PlayersToSubmittedJobs[uuid] = make([]*string, 0)
	}

	s.PlayersToSubmittedJobs[uuid] = append(s.PlayersToSubmittedJobs[uuid], sj)
	s.JobPool = append(s.JobPool, sj)

	if !utils.IsProductionEnv() {
		jobs := make([]string, len(s.PlayersToSubmittedJobs[uuid]))
		for i, job := range s.PlayersToSubmittedJobs[uuid] {
			jobs[i] = *job
		}
		logger.Debugf("[server] %s has submitted { %s }", uuid, strings.Join(jobs, ", "))
	}
}
