package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Minimal dotsync.toml
	os.WriteFile(filepath.Join(dir, "dotsync.toml"), []byte(`
[[steps]]
name   = "brew"
inputs = ["Brewfile"]
`), 0644)

	// A Brewfile to hash
	os.WriteFile(filepath.Join(dir, "Brewfile"), []byte(`brew "git"`), 0644)

	return dir
}

func TestCheckNeverRun(t *testing.T) {
	dir := setupTestRepo(t)

	// No state file exists — check should exit 1 (stale)
	err := cmdCheck(dir, "brew")
	if err == nil {
		t.Error("expected error (exit 1) for never-run step")
	}
}

func TestCheckUpToDate(t *testing.T) {
	dir := setupTestRepo(t)

	// Record the current state manually
	topHash, files, _ := HashStep(dir, []string{"Brewfile"})
	state := &DotfilesState{
		Meta:  StateMeta{CurrentCommit: "abc1234", DotsyncVersion: DotsyncVersion},
		Steps: map[string]*StepState{
			"brew": {
				Hash: topHash, Files: files,
				RecordedAt: time.Now().UTC().Format(time.RFC3339),
				Commit: "abc1234", Status: "ok",
			},
		},
	}
	SaveState(dir, state)

	// check should exit 0
	if err := cmdCheck(dir, "brew"); err != nil {
		t.Errorf("expected up-to-date (nil error), got: %v", err)
	}
}

func TestCheckStale(t *testing.T) {
	dir := setupTestRepo(t)

	// Store a wrong hash
	state := &DotfilesState{
		Meta:  StateMeta{CurrentCommit: "abc1234", DotsyncVersion: DotsyncVersion},
		Steps: map[string]*StepState{
			"brew": {Hash: "oldhash", Files: map[string]string{}, RecordedAt: time.Now().UTC().Format(time.RFC3339), Commit: "abc1234", Status: "ok"},
		},
	}
	SaveState(dir, state)

	err := cmdCheck(dir, "brew")
	if err == nil {
		t.Error("expected stale error")
	}
}

func TestRecordCreatesStateEntry(t *testing.T) {
	dir := setupTestRepo(t)

	if err := cmdRecord(dir, "brew"); err != nil {
		t.Fatal(err)
	}

	state, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	step, ok := state.Steps["brew"]
	if !ok {
		t.Fatal("expected brew entry in state")
	}
	if step.Hash == "" {
		t.Error("expected non-empty hash")
	}
	if step.Status != "ok" {
		t.Errorf("expected status ok, got %q", step.Status)
	}
	if _, ok := step.Files["Brewfile"]; !ok {
		t.Error("expected Brewfile in file map")
	}
}

func TestRecordThenCheckUpToDate(t *testing.T) {
	dir := setupTestRepo(t)

	if err := cmdRecord(dir, "brew"); err != nil {
		t.Fatal(err)
	}
	// Immediately after recording, check should return up-to-date
	if err := cmdCheck(dir, "brew"); err != nil {
		t.Errorf("expected up-to-date after record, got: %v", err)
	}
}

func TestRecordAppendsSnapshot(t *testing.T) {
	dir := setupTestRepo(t)

	if err := cmdRecord(dir, "brew"); err != nil {
		t.Fatal(err)
	}

	snaps, err := LoadSnapshots(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps) == 0 {
		t.Error("expected at least one snapshot after record")
	}
}

func TestRollback(t *testing.T) {
	dir := setupTestRepo(t)

	// Record initial state
	if err := cmdRecord(dir, "brew"); err != nil {
		t.Fatal(err)
	}

	// Get the commit that was recorded
	state, _ := LoadState(dir)
	commit := state.Meta.CurrentCommit

	// Corrupt the current state to simulate drift
	state.Steps["brew"].Hash = "wronghash"
	SaveState(dir, state)

	// Rollback should restore
	if err := cmdRollback(dir, commit); err != nil {
		t.Fatal(err)
	}

	restored, _ := LoadState(dir)
	step := restored.Steps["brew"]
	if step == nil {
		t.Fatal("expected brew step after rollback")
	}
	if step.Hash == "wronghash" {
		t.Error("rollback did not restore hash")
	}
}

func TestRollbackUnknownCommit(t *testing.T) {
	dir := setupTestRepo(t)
	err := cmdRollback(dir, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown commit")
	}
}

func TestGCRemovesOrphans(t *testing.T) {
	dir := setupTestRepo(t)

	// Manually write a state entry for a step not in dotsync.toml
	state := &DotfilesState{
		Meta:  StateMeta{CurrentCommit: "abc", DotsyncVersion: DotsyncVersion},
		Steps: map[string]*StepState{
			"brew":    {Hash: "h1", Status: "ok", Files: map[string]string{}},
			"orphan":  {Hash: "h2", Status: "ok", Files: map[string]string{}},
		},
	}
	SaveState(dir, state)

	if err := cmdGC(dir); err != nil {
		t.Fatal(err)
	}

	after, _ := LoadState(dir)
	if _, ok := after.Steps["orphan"]; ok {
		t.Error("expected orphan to be removed")
	}
	if _, ok := after.Steps["brew"]; !ok {
		t.Error("expected brew to remain")
	}
}
