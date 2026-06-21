package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dotsync: %v\n", err)
		os.Exit(2)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(2)
	}

	switch args[0] {
	case "check":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: dotsync check <step>")
			os.Exit(2)
		}
		err := cmdCheck(repoRoot, args[1])
		if err == errStale {
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "dotsync check: %v\n", err)
			os.Exit(2)
		}
	case "record":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: dotsync record <step>")
			os.Exit(2)
		}
		if err := cmdRecord(repoRoot, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "dotsync record: %v\n", err)
			os.Exit(2)
		}
	case "status":
		if err := cmdStatus(repoRoot); err != nil {
			fmt.Fprintf(os.Stderr, "dotsync status: %v\n", err)
			os.Exit(2)
		}
	case "rollback":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: dotsync rollback <commit>")
			os.Exit(2)
		}
		if err := cmdRollback(repoRoot, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "dotsync rollback: %v\n", err)
			os.Exit(2)
		}
	case "gc":
		if err := cmdGC(repoRoot); err != nil {
			fmt.Fprintf(os.Stderr, "dotsync gc: %v\n", err)
			os.Exit(2)
		}
	default:
		fmt.Fprintf(os.Stderr, "dotsync: unknown command %q\n", args[0])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: dotsync <command> [args]

commands:
  check <step>      exit 0 if up-to-date, 1 if stale
  record <step>     record current hash for step
  status            show all steps and drift
  rollback <commit> restore state from snapshot
  gc                remove orphaned state entries`)
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "dotsync.toml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find dotsync.toml in any parent directory")
		}
		dir = parent
	}
}

func currentGitCommit(repoRoot string) (string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown", nil // non-fatal: no git, detached HEAD, etc.
	}
	return strings.TrimSpace(string(out)), nil
}

// Stub implementations for subcommands (to be implemented in later tasks)

func cmdStatus(repoRoot string) error {
	return fmt.Errorf("not implemented")
}

func cmdRollback(repoRoot, commit string) error {
	return fmt.Errorf("not implemented")
}

func cmdGC(repoRoot string) error {
	return fmt.Errorf("not implemented")
}
