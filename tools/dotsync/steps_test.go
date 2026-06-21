package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	content := `
[[steps]]
name   = "brew"
inputs = ["Brewfile"]

[[steps]]
name        = "vscode_profiles"
inputs      = ["vscode/profiles/"]
drift_check = "vscode"
`
	os.WriteFile(filepath.Join(dir, "dotsync.toml"), []byte(content), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(cfg.Steps))
	}
	if cfg.Steps[0].Name != "brew" {
		t.Errorf("expected first step name 'brew', got %q", cfg.Steps[0].Name)
	}
}

func TestFindStep(t *testing.T) {
	cfg := &DotsyncConfig{
		Steps: []StepDef{
			{Name: "brew", Inputs: []string{"Brewfile"}},
			{Name: "stow_zsh", Inputs: []string{"zsh/"}},
		},
	}

	step, ok := cfg.FindStep("stow_zsh")
	if !ok {
		t.Fatal("expected to find stow_zsh")
	}
	if step.Name != "stow_zsh" {
		t.Errorf("got %q", step.Name)
	}

	_, ok = cfg.FindStep("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}
