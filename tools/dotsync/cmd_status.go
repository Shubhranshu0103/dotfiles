package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	symOK      = "✓ up-to-date"
	symStale   = "✗ stale"
	symNever   = "— never run"
	symDrifted = "⚠ drifted"
)

func cmdStatus(repoRoot string) error {
	cfg, err := LoadConfig(repoRoot)
	if err != nil {
		return err
	}

	state, err := LoadState(repoRoot)
	if err != nil {
		return err
	}

	snaps, _ := LoadSnapshots(repoRoot)

	fmt.Printf("%-22s %-16s %-22s %-10s %s\n", "Step", "Status", "Last Run", "Commit", "Drift")
	fmt.Println(strings.Repeat("─", 90))

	for _, stepDef := range cfg.Steps {
		recorded, exists := state.Steps[stepDef.Name]

		status := symNever
		lastRun := "—"
		commit := "—"
		drift := "—"
		var staleFiles []string

		if exists {
			currentHash, currentFiles, err := HashStep(repoRoot, stepDef.Inputs)
			if err != nil {
				status = "error"
			} else if currentHash == recorded.Hash {
				status = symOK
			} else {
				status = symStale
				for path, h := range currentFiles {
					if recorded.Files[path] != h {
						staleFiles = append(staleFiles, path)
					}
				}
			}
			lastRun = recorded.RecordedAt
			if len(lastRun) > 16 {
				lastRun = lastRun[:16]
			}
			commit = recorded.Commit
		}

		if stepDef.DriftCheck == "vscode" && exists {
			if d := checkVSCodeDrift(repoRoot, stepDef); d != "" {
				drift = symDrifted
				fmt.Printf("%-22s %-16s %-22s %-10s %s\n", stepDef.Name, status, lastRun, commit, drift)
				fmt.Printf("  └ %s (run: dots vscode-export)\n", d)
				continue
			}
		}

		fmt.Printf("%-22s %-16s %-22s %-10s %s\n", stepDef.Name, status, lastRun, commit, drift)
		for _, f := range staleFiles {
			fmt.Printf("  └ %s modified\n", f)
		}
	}

	if len(snaps) > 0 {
		fmt.Printf("\nSnapshots: %d stored", len(snaps))
		if len(snaps) >= 2 {
			fmt.Printf("  →  rollback available to %s", snaps[0].Commit)
		}
		fmt.Println()
	}

	return nil
}

// checkVSCodeDrift exports each VS Code profile to a temp file and compares
// against the committed .code-profile. Returns a description if drifted, "" if clean.
func checkVSCodeDrift(repoRoot string, step StepDef) string {
	if _, err := exec.LookPath("code"); err != nil {
		return "" // VS Code CLI not available — skip drift check gracefully
	}

	profiles := []string{"ML", "Java", "WebDev", "Rust"}
	for _, profile := range profiles {
		tmp, err := os.CreateTemp("", fmt.Sprintf("dotsync-%s-*.code-profile", profile))
		if err != nil {
			continue
		}
		tmp.Close()
		defer os.Remove(tmp.Name())

		cmd := exec.Command("code", "--profile", profile, "--export-profile", tmp.Name())
		if err := cmd.Run(); err != nil {
			continue // profile may not exist yet — skip
		}

		repoFile := fmt.Sprintf("%s/vscode/profiles/%s.code-profile", repoRoot, strings.ToLower(profile))
		repoData, err := os.ReadFile(repoFile)
		if err != nil {
			continue
		}
		liveData, err := os.ReadFile(tmp.Name())
		if err != nil {
			continue
		}

		if !profilesEqual(repoData, liveData) {
			return fmt.Sprintf("%s.code-profile has UI changes", strings.ToLower(profile))
		}
	}
	return ""
}

// profilesEqual compares the settings and extensions fields of two .code-profile JSON blobs.
// Ignores metadata fields like version and timestamps.
func profilesEqual(a, b []byte) bool {
	extract := func(data []byte) map[string]any {
		var full map[string]any
		if err := json.Unmarshal(data, &full); err != nil {
			return nil
		}
		return map[string]any{
			"settings":   full["settings"],
			"extensions": full["extensions"],
		}
	}

	aMap := extract(a)
	bMap := extract(b)
	if aMap == nil || bMap == nil {
		return true // can't compare — assume clean
	}

	aJSON, _ := json.Marshal(aMap)
	bJSON, _ := json.Marshal(bMap)
	return string(aJSON) == string(bJSON)
}
