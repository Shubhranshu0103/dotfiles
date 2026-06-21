package main

import "fmt"

func cmdGC(repoRoot string) error {
	cfg, err := LoadConfig(repoRoot)
	if err != nil {
		return err
	}

	known := make(map[string]bool)
	for _, step := range cfg.Steps {
		known[step.Name] = true
	}

	state, err := LoadState(repoRoot)
	if err != nil {
		return err
	}

	removed := 0
	for name := range state.Steps {
		if !known[name] {
			delete(state.Steps, name)
			removed++
		}
	}

	if removed == 0 {
		fmt.Println("Nothing to clean up.")
		return nil
	}

	fmt.Printf("Removed %d orphaned step(s).\n", removed)
	return SaveState(repoRoot, state)
}
