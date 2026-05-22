package shouldtest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/curiostorage/shouldtest/config"
)

const profileFileName = "shouldtest.json"

// Profile is JSON config beside a package's tests (<dir>/shouldtest.json).
type Profile struct {
	BuildTags    []string `json:"buildTags"`
	ExtraPaths   []string `json:"extraPaths"`
	CoverageFile string   `json:"coverageFile"`
}

// LoadProfile reads <dir>/shouldtest.json.
func LoadProfile(dir string) (Profile, error) {
	if dir == "" {
		return Profile{}, fmt.Errorf("profile directory is required")
	}
	path := filepath.Join(dir, profileFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, fmt.Errorf("read profile %s: %w", path, err)
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return Profile{}, fmt.Errorf("parse profile %s: %w", path, err)
	}
	return p, nil
}

func listTargetPattern(profileDir string) (string, error) {
	rel := normalizePath(profileDir)
	if filepath.IsAbs(profileDir) {
		root, err := config.ModuleRoot()
		if err != nil {
			return "", err
		}
		rel, err = filepath.Rel(root, profileDir)
		if err != nil {
			return "", err
		}
		rel = normalizePath(rel)
	}
	return "./" + rel + "/...", nil
}

func optionsForProfile(profileDir string, p Profile, overrides Options) (Options, error) {
	target, err := listTargetPattern(profileDir)
	if err != nil {
		return Options{}, err
	}
	return Options{
		Targets:       []string{target},
		BuildTags:     append([]string(nil), p.BuildTags...),
		ExtraPaths:    append([]string(nil), p.ExtraPaths...),
		BaseRef:       overrides.BaseRef,
		HeadRef:       overrides.HeadRef,
		DefaultBranch: overrides.DefaultBranch,
		Changed:       overrides.Changed,
		AlwaysRun:     overrides.AlwaysRun,
	}, nil
}

// OptionsForProfile builds Check options from <dir>/shouldtest.json.
func OptionsForProfile(dir string, overrides Options) (Options, error) {
	p, err := LoadProfile(dir)
	if err != nil {
		return Options{}, err
	}
	return optionsForProfile(dir, p, overrides)
}
