package main

import (
	"fmt"
)

// errStale is returned when a step is stale or never run.
// main() converts this to exit code 1.
var errStale = fmt.Errorf("stale")

func cmdCheck(repoRoot, stepName string) error {
	cfg, err := LoadConfig(repoRoot)
	if err != nil {
		return err
	}

	step, ok := cfg.FindStep(stepName)
	if !ok {
		// Unknown step: treat as stale so install.sh runs it and record adds it
		return errStale
	}

	state, err := LoadState(repoRoot)
	if err != nil {
		return err
	}

	recorded, ok := state.Steps[stepName]
	if !ok {
		return errStale
	}

	currentHash, _, err := HashStep(repoRoot, step.Inputs)
	if err != nil {
		return err
	}

	if currentHash != recorded.Hash {
		return errStale
	}
	return nil
}
