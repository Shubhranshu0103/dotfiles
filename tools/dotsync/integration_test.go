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
