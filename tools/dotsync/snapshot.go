package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const snapshotFileName = ".dotfiles-snapshots"
const maxSnapshots = 10

type SnapshotFile []Snapshot

type Snapshot struct {
	Commit     string                  `json:"commit"`
	RecordedAt string                  `json:"recorded_at"`
	Steps      map[string]SnapshotStep `json:"steps"`
}

type SnapshotStep struct {
	Hash  string            `json:"hash"`
	Files map[string]string `json:"files"`
}

func LoadSnapshots(repoRoot string) (SnapshotFile, error) {
	path := filepath.Join(repoRoot, snapshotFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return SnapshotFile{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading snapshots: %w", err)
	}
	var snaps SnapshotFile
	if err := json.Unmarshal(data, &snaps); err != nil {
		return nil, fmt.Errorf("parsing snapshots: %w", err)
	}
	return snaps, nil
}

func SaveSnapshots(repoRoot string, snaps SnapshotFile) error {
	path := filepath.Join(repoRoot, snapshotFileName)
	data, err := json.MarshalIndent(snaps, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding snapshots: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (s SnapshotFile) FindByCommit(commit string) (*Snapshot, bool) {
	for i := range s {
		if s[i].Commit == commit {
			return &s[i], true
		}
	}
	return nil, false
}

func AppendSnapshot(repoRoot string, snap Snapshot) error {
	snaps, err := LoadSnapshots(repoRoot)
	if err != nil {
		return err
	}
	// Replace existing entry for same commit, or append
	for i, s := range snaps {
		if s.Commit == snap.Commit {
			snaps[i] = snap
			return SaveSnapshots(repoRoot, snaps)
		}
	}
	snaps = append(snaps, snap)
	if len(snaps) > maxSnapshots {
		snaps = snaps[len(snaps)-maxSnapshots:]
	}
	return SaveSnapshots(repoRoot, snaps)
}
