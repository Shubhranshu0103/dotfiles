package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HashStep computes a top-level hash and per-file map for a step's inputs.
// Inputs ending in "/" are walked recursively relative to repoRoot.
func HashStep(repoRoot string, inputs []string) (string, map[string]string, error) {
	files := make(map[string]string)

	for _, input := range inputs {
		fullPath := filepath.Join(repoRoot, input)
		if strings.HasSuffix(input, "/") {
			dirFiles, err := hashDir(repoRoot, fullPath)
			if err != nil {
				return "", nil, err
			}
			for k, v := range dirFiles {
				files[k] = v
			}
		} else {
			// Skip files that don't exist yet (e.g. vscode/settings.json before
			// it's added to the repo). A step with all-missing inputs hashes to
			// an empty string, which never matches stored state → always stale.
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				continue
			}
			h, err := hashFile(fullPath)
			if err != nil {
				return "", nil, fmt.Errorf("hashing %s: %w", input, err)
			}
			files[input] = h
		}
	}

	topHash, err := combineHashes(files)
	if err != nil {
		return "", nil, err
	}
	return topHash, files, nil
}

// hashDir walks root recursively, returning a map of repo-relative paths to SHA256 hashes.
func hashDir(repoRoot, root string) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		h, err := hashFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		files[rel] = h
		return nil
	})
	return files, err
}

// combineHashes produces a single SHA256 from a map of path→hash pairs,
// sorted by path so the result is deterministic.
func combineHashes(files map[string]string) (string, error) {
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s:%s\n", k, files[k])
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// hashFile returns the SHA256 hex digest of a file's contents.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
