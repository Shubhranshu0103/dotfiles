package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const DotsyncVersion = "0.1.0"
const stateFileName = ".dotfiles-state"

type DotfilesState struct {
	Meta  StateMeta             `toml:"meta"`
	Steps map[string]*StepState `toml:"steps"`
}

type StateMeta struct {
	CurrentCommit   string `toml:"current_commit"`
	CurrentCommitAt string `toml:"current_commit_at"`
	DotsyncVersion  string `toml:"dotsync_version"`
}

type StepState struct {
	Hash       string            `toml:"hash"`
	RecordedAt string            `toml:"recorded_at"`
	Commit     string            `toml:"commit"`
	Status     string            `toml:"status"` // "ok" | "never_run"
	Files      map[string]string `toml:"files"`
}

func LoadState(repoRoot string) (*DotfilesState, error) {
	path := filepath.Join(repoRoot, stateFileName)
	state := &DotfilesState{
		Steps: make(map[string]*StepState),
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return state, nil
	}
	if _, err := toml.DecodeFile(path, state); err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}
	if state.Steps == nil {
		state.Steps = make(map[string]*StepState)
	}
	return state, nil
}

func SaveState(repoRoot string, state *DotfilesState) error {
	path := filepath.Join(repoRoot, stateFileName)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("saving state: %w", err)
	}
	defer f.Close()
	f.WriteString("# Managed by dotsync — do not edit manually.\n")
	f.WriteString("# Run `dotsync status` for a pretty view.\n\n")
	return toml.NewEncoder(f).Encode(state)
}
