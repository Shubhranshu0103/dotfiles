package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// checkVSCodeDrift reads VS Code's internal storage directly (no CLI invocation)
// to compare live extension lists against committed .code-profile files.
// Returns a description if drifted, "" if clean.
func checkVSCodeDrift(repoRoot string, step StepDef) string {
	home := os.Getenv("HOME")
	vsCodeBase := filepath.Join(home, "Library/Application Support/Code/User")

	// Read profile name→location ID mapping from VS Code's global storage.
	storageData, err := os.ReadFile(filepath.Join(vsCodeBase, "globalStorage/storage.json"))
	if err != nil {
		return "" // VS Code not installed or never launched
	}

	var storage map[string]json.RawMessage
	if err := json.Unmarshal(storageData, &storage); err != nil {
		return ""
	}

	var profileDefs []struct {
		Location string `json:"location"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(storage["userDataProfiles"], &profileDefs); err != nil {
		return ""
	}

	locationByName := make(map[string]string, len(profileDefs))
	for _, p := range profileDefs {
		locationByName[strings.ToLower(p.Name)] = p.Location
	}

	profiles := []string{"ML", "Java", "WebDev", "Rust"}
	for _, profile := range profiles {
		location, ok := locationByName[strings.ToLower(profile)]
		if !ok {
			continue // profile not created in VS Code yet
		}

		// Read live extensions from internal storage.
		extData, err := os.ReadFile(filepath.Join(vsCodeBase, "profiles", location, "extensions.json"))
		if err != nil {
			continue
		}

		var extList []struct {
			Identifier struct {
				ID string `json:"id"`
			} `json:"identifier"`
		}
		if err := json.Unmarshal(extData, &extList); err != nil {
			continue
		}

		liveIDs := make([]string, 0, len(extList))
		for _, e := range extList {
			liveIDs = append(liveIDs, strings.ToLower(e.Identifier.ID))
		}
		sort.Strings(liveIDs)

		// Read committed profile (extensions stored as space-separated string).
		repoData, err := os.ReadFile(filepath.Join(repoRoot, "vscode/profiles", strings.ToLower(profile)+".code-profile"))
		if err != nil {
			continue
		}

		var repoProfile struct {
			Extensions string `json:"extensions"`
		}
		if err := json.Unmarshal(repoData, &repoProfile); err != nil {
			continue
		}

		repoIDs := make([]string, 0)
		for _, id := range strings.Fields(repoProfile.Extensions) {
			repoIDs = append(repoIDs, strings.ToLower(id))
		}
		sort.Strings(repoIDs)

		if !stringSlicesEqual(liveIDs, repoIDs) {
			return fmt.Sprintf("%s.code-profile has UI changes", strings.ToLower(profile))
		}
	}
	return ""
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
