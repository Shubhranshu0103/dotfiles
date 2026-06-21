package main

import (
	"fmt"
	"testing"
	"time"
)

func TestStateRoundTrip(t *testing.T) {
	dir := t.TempDir()

	state := &DotfilesState{
		Meta: StateMeta{
			CurrentCommit:   "abc1234",
			CurrentCommitAt: time.Now().UTC().Format(time.RFC3339),
			DotsyncVersion:  DotsyncVersion,
		},
		Steps: map[string]*StepState{
			"brew": {
				Hash:       "deadbeef",
				RecordedAt: time.Now().UTC().Format(time.RFC3339),
				Commit:     "abc1234",
				Status:     "ok",
				Files:      map[string]string{"Brewfile": "deadbeef"},
			},
		},
	}

	if err := SaveState(dir, state); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Meta.CurrentCommit != "abc1234" {
		t.Errorf("got %q", loaded.Meta.CurrentCommit)
	}
	step, ok := loaded.Steps["brew"]
	if !ok {
		t.Fatal("expected brew step")
	}
	if step.Hash != "deadbeef" {
		t.Errorf("got hash %q", step.Hash)
	}
}

func TestLoadStateEmpty(t *testing.T) {
	dir := t.TempDir()
	state, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Steps) != 0 {
		t.Error("expected empty steps on missing state file")
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	dir := t.TempDir()

	snap := Snapshot{
		Commit:     "abc1234",
		RecordedAt: time.Now().UTC().Format(time.RFC3339),
		Steps: map[string]SnapshotStep{
			"brew": {Hash: "deadbeef", Files: map[string]string{"Brewfile": "deadbeef"}},
		},
	}

	if err := AppendSnapshot(dir, snap); err != nil {
		t.Fatal(err)
	}

	snaps, err := LoadSnapshots(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
	found, ok := snaps.FindByCommit("abc1234")
	if !ok {
		t.Fatal("expected to find snapshot by commit")
	}
	if found.Steps["brew"].Hash != "deadbeef" {
		t.Errorf("got hash %q", found.Steps["brew"].Hash)
	}
}

func TestSnapshotCap(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 12; i++ {
		snap := Snapshot{
			Commit:     fmt.Sprintf("commit%02d", i),
			RecordedAt: time.Now().UTC().Format(time.RFC3339),
			Steps:      map[string]SnapshotStep{},
		}
		if err := AppendSnapshot(dir, snap); err != nil {
			t.Fatal(err)
		}
	}

	snaps, _ := LoadSnapshots(dir)
	if len(snaps) != maxSnapshots {
		t.Errorf("expected %d snapshots, got %d", maxSnapshots, len(snaps))
	}
	// Oldest entries should be pruned
	_, ok := snaps.FindByCommit("commit00")
	if ok {
		t.Error("expected commit00 to be pruned")
	}
	_, ok = snaps.FindByCommit("commit11")
	if !ok {
		t.Error("expected commit11 to be present")
	}
}
