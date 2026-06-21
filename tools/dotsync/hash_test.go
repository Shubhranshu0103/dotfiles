package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "file.txt")
	os.WriteFile(f, []byte("hello"), 0644)

	h, err := hashFile(f)
	if err != nil {
		t.Fatal(err)
	}
	// SHA256("hello") = 2cf24dba...
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if h != want {
		t.Errorf("got %s, want %s", h, want)
	}
}

func TestHashStepSingleFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Brewfile"), []byte("brew \"git\""), 0644)

	topHash, files, err := HashStep(dir, []string{"Brewfile"})
	if err != nil {
		t.Fatal(err)
	}
	if topHash == "" {
		t.Error("expected non-empty top hash")
	}
	if _, ok := files["Brewfile"]; !ok {
		t.Error("expected Brewfile in files map")
	}
}

func TestHashStepDirectory(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "zsh")
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, ".zshrc"), []byte("# zsh"), 0644)
	os.WriteFile(filepath.Join(subdir, ".zprofile"), []byte("# zprofile"), 0644)

	topHash, files, err := HashStep(dir, []string{"zsh/"})
	if err != nil {
		t.Fatal(err)
	}
	if topHash == "" {
		t.Error("expected non-empty top hash")
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestHashStepDeterministic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("bbb"), 0644)

	h1, _, _ := HashStep(dir, []string{"a.txt", "b.txt"})
	h2, _, _ := HashStep(dir, []string{"b.txt", "a.txt"})
	if h1 != h2 {
		t.Error("hash should be order-independent")
	}
}
