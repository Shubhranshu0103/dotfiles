package main

import (
	"fmt"
	"strings"
)

const (
	symOK    = "✓ up-to-date"
	symStale = "✗ stale"
	symNever = "— never run"
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

	fmt.Printf("%-22s %-16s %-22s %s\n", "Step", "Status", "Last Run", "Commit")
	fmt.Println(strings.Repeat("─", 75))

	for _, stepDef := range cfg.Steps {
		recorded, exists := state.Steps[stepDef.Name]

		status := symNever
		lastRun := "—"
		commit := "—"
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

		fmt.Printf("%-22s %-16s %-22s %s\n", stepDef.Name, status, lastRun, commit)
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
