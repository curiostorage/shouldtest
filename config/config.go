// Package config loads per-repository settings from .github/shouldtest.json.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/curiostorage/shouldtest/registry"
)

const RepoFileName = "shouldtest.json"

// Repo is repository-wide shouldtest configuration (.github/shouldtest.json).
type Repo struct {
	DefaultBranch                  string   `json:"defaultBranch"`
	AlwaysRunOnPushToDefaultBranch bool     `json:"alwaysRunOnPushToDefaultBranch"`
	Workflows                      []string `json:"workflows"`
	WalkSkipDirs                   []string `json:"walkSkipDirs"`
	IgnoreProfileUnder             []string `json:"ignoreProfileUnder"`
	CIInvoke                       string   `json:"ciInvoke"`
}

// Load reads .github/shouldtest.json from moduleRoot.
func Load(moduleRoot string) (Repo, error) {
	path := filepath.Join(moduleRoot, ".github", RepoFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Repo{}, fmt.Errorf("read repo config %s: %w", path, err)
	}
	var repo Repo
	if err := json.Unmarshal(data, &repo); err != nil {
		return Repo{}, fmt.Errorf("parse repo config %s: %w", path, err)
	}
	repo.applyDefaults()
	return repo, nil
}

// LoadFromModule discovers the module root and loads repo config.
func LoadFromModule() (Repo, string, error) {
	root, err := ModuleRoot()
	if err != nil {
		return Repo{}, "", err
	}
	repo, err := Load(root)
	return repo, root, err
}

func (r *Repo) applyDefaults() {
	if r.DefaultBranch == "" {
		r.DefaultBranch = "main"
	}
	if len(r.Workflows) == 0 {
		r.Workflows = []string{".github/workflows/ci.yml"}
	}
	if len(r.WalkSkipDirs) == 0 {
		r.WalkSkipDirs = []string{".git", "node_modules", "vendor"}
	}
	if r.CIInvoke == "" {
		r.CIInvoke = "shouldtest"
	}
}

// RegistryConfig converts repo settings for CI profile registry checks.
func (r Repo) RegistryConfig() registry.Config {
	return registry.Config{
		ProfileFileName:    "shouldtest.json",
		Workflows:          append([]string(nil), r.Workflows...),
		WalkSkipDirs:       append([]string(nil), r.WalkSkipDirs...),
		IgnoreProfileUnder: append([]string(nil), r.IgnoreProfileUnder...),
		CIInvoke:           r.CIInvoke,
	}
}

// DefaultBranchRef returns refs/heads/<defaultBranch> for GitHub Actions.
func (r Repo) DefaultBranchRef() string {
	return "refs/heads/" + r.DefaultBranch
}

// ModuleRoot returns the directory containing go.mod.
func ModuleRoot() (string, error) {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", err
	}
	mod := strings.TrimSpace(string(out))
	if mod == "" || mod == "/dev/null" {
		return "", fmt.Errorf("not in a go module")
	}
	return filepath.Clean(filepath.Dir(mod)), nil
}
