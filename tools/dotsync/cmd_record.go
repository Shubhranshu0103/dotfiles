package main

import (
	"fmt"
	"time"
)

func cmdRecord(repoRoot, stepName string) error {
	cfg, err := LoadConfig(repoRoot)
	if err != nil {
		return err
	}

	step, ok := cfg.FindStep(stepName)
	if !ok {
		return fmt.Errorf("step %q not found in dotsync.toml", stepName)
	}

	topHash, files, err := HashStep(repoRoot, step.Inputs)
	if err != nil {
		return err
	}

	commit, err := currentGitCommit(repoRoot)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)

	state, err := LoadState(repoRoot)
	if err != nil {
		return err
	}

	state.Steps[stepName] = &StepState{
		Hash:       topHash,
		RecordedAt: now,
		Commit:     commit,
		Status:     "ok",
		Files:      files,
	}
	state.Meta.CurrentCommit = commit
	state.Meta.CurrentCommitAt = now
	state.Meta.DotsyncVersion = DotsyncVersion

	if err := SaveState(repoRoot, state); err != nil {
		return err
	}

	// Build snapshot from current full state
	snap := Snapshot{
		Commit:     commit,
		RecordedAt: now,
		Steps:      make(map[string]SnapshotStep),
	}
	for name, s := range state.Steps {
		snap.Steps[name] = SnapshotStep{Hash: s.Hash, Files: s.Files}
	}

	return AppendSnapshot(repoRoot, snap)
}
