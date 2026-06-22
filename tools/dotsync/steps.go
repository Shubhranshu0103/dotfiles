package main

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const configFileName = "dotsync.toml"

type StepDef struct {
	Name   string   `toml:"name"`
	Inputs []string `toml:"inputs"`
}

type DotsyncConfig struct {
	Steps []StepDef `toml:"steps"`
}

func LoadConfig(repoRoot string) (*DotsyncConfig, error) {
	path := filepath.Join(repoRoot, configFileName)
	var cfg DotsyncConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("loading dotsync.toml: %w", err)
	}
	return &cfg, nil
}

func (c *DotsyncConfig) FindStep(name string) (*StepDef, bool) {
	for i := range c.Steps {
		if c.Steps[i].Name == name {
			return &c.Steps[i], true
		}
	}
	return nil, false
}
