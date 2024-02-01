package game

import (
	"os"
	"path/filepath"
	"time"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"gopkg.in/yaml.v2"
)

const (
	cfgYamlFileName = "config.yml"
)

type GameConfig struct {
	Limits LimitConfig `yaml:"limits"`
	Times  TimeConfig  `yaml:"times"`
}

type LimitConfig struct {
	MinimumNumberOfPlayers int `yaml:"minimum_number_of_players"`
}

type TimeConfig struct {
	ImprovRoundDurationSeconds   int `yaml:"improv_round_duration_seconds"`
	InterceptionTimeAddedSeconds int `yaml:"interception_time_added_seconds"`
	IntermissionDurationSeconds  int `yaml:"intermission_duration_seconds"`
}

// Attempts to read the local game configuration, returns a default if retrieval fails.
func GetGameConfig() *GameConfig {
	var cfg GameConfig
	err := tryReadConfigFile(&cfg)
	if err != nil {
		logger.Errorf("[config] Failed to load configuration file: %v, using default configuration", err)
		return GetDefaultGameConfig()
	}

	return &cfg
}

// Retrieves the default game configuration.
func GetDefaultGameConfig() *GameConfig {
	return &GameConfig{
		Limits: LimitConfig{
			MinimumNumberOfPlayers: 3,
		},
		Times: TimeConfig{
			ImprovRoundDurationSeconds:   30,
			InterceptionTimeAddedSeconds: 30,
			IntermissionDurationSeconds:  10,
		},
	}
}

// Retrieves the improv round duration as a time.Duration.
func (cfg *GameConfig) GetTypedImprovRoundDurationSeconds() time.Duration {
	return time.Duration(cfg.Times.ImprovRoundDurationSeconds) * time.Second
}

// Retrieves the interception time added as a time.Duration.
func (cfg *GameConfig) GetTypedInterceptionTimeAddedSeconds() time.Duration {
	return time.Duration(cfg.Times.InterceptionTimeAddedSeconds) * time.Second
}

// Retrieves the intermission duration as a time.Duration
func (cfg *GameConfig) GetTypedIntermissionDurationSeconds() time.Duration {
	return time.Duration(cfg.Times.IntermissionDurationSeconds) * time.Second
}

// Tries to read the config file and decode it into the GameConfig struct.
func tryReadConfigFile(cfg *GameConfig) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	path := filepath.Join(wd, cfgYamlFileName)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}

	return nil
}
