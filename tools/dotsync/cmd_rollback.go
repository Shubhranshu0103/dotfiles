package main

import (
	"fmt"
	"time"
)

func cmdRollback(repoRoot, commit string) error {
	snaps, err := LoadSnapshots(repoRoot)
	if err != nil {
		return err
	}

	snap, ok := snaps.FindByCommit(commit)
	if !ok {
		return fmt.Errorf("no snapshot found for commit %q", commit)
	}

	state, err := LoadState(repoRoot)
	if err != nil {
		return err
	}

	// Restore snapshot steps into current state
	state.Steps = make(map[string]*StepState)
	for name, s := range snap.Steps {
		state.Steps[name] = &StepState{
			Hash:       s.Hash,
			Files:      s.Files,
			RecordedAt: snap.RecordedAt,
			Commit:     snap.Commit,
			Status:     "ok",
		}
	}
	state.Meta.CurrentCommit = commit
	state.Meta.CurrentCommitAt = time.Now().UTC().Format(time.RFC3339)

	fmt.Printf("Rolled back to commit %s\n", commit)
	return SaveState(repoRoot, state)
}
