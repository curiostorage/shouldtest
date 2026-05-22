package shouldtest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/curiostorage/shouldtest/config"
)

// CIConfig configures GitHub Actions integration.
type CIConfig struct {
	ProfileDir  string
	CoverageDir string
	OutputKey   string
	Verbose     bool
	Repo        config.Repo
}

// CIResult is the outcome of a CI decision step.
type CIResult struct {
	Result
	ProfileDir   string
	SkippedOnCI  bool
	CoveragePath string
}

// RunCI decides whether tests should run and applies GitHub Actions side effects.
func RunCI(cfg CIConfig) (CIResult, error) {
	if err := validateCIEnv(); err != nil {
		return CIResult{}, err
	}
	if cfg.ProfileDir == "" {
		return CIResult{}, fmt.Errorf("profile directory is required")
	}

	repo := cfg.Repo
	if repo.DefaultBranch == "" {
		var err error
		repo, _, err = config.LoadFromModule()
		if err != nil {
			return CIResult{}, err
		}
	}

	p, err := LoadProfile(cfg.ProfileDir)
	if err != nil {
		return CIResult{}, err
	}

	covDir := cfg.CoverageDir
	if covDir == "" {
		covDir = "coverage"
	}
	outKey := cfg.OutputKey
	if outKey == "" {
		outKey = "run"
	}

	overrides := Options{
		BaseRef:       os.Getenv("GITHUB_BASE_SHA"),
		HeadRef:       os.Getenv("GITHUB_HEAD_SHA"),
		DefaultBranch: repo.DefaultBranch,
	}

	var result Result
	if alwaysRunOnDefaultBranchPush(repo) {
		result = Result{Run: true, Reason: "push to " + repo.DefaultBranch}
	} else {
		opts, err := optionsForProfile(cfg.ProfileDir, p, overrides)
		if err != nil {
			return CIResult{}, err
		}
		result, err = Check(opts)
		if err != nil {
			return CIResult{}, err
		}
	}

	if cfg.Verbose {
		LogResult(cfg.ProfileDir, result)
	}

	ci := CIResult{Result: result, ProfileDir: cfg.ProfileDir}
	if err := writeGitHubOutput(outKey, result.Run); err != nil {
		return ci, err
	}

	if !result.Run && p.CoverageFile != "" {
		if err := os.MkdirAll(covDir, 0o755); err != nil {
			return ci, err
		}
		path := filepath.Join(covDir, p.CoverageFile)
		if err := os.WriteFile(path, []byte("mode: atomic\n"), 0o644); err != nil {
			return ci, err
		}
		ci.SkippedOnCI = true
		ci.CoveragePath = path
	}

	return ci, nil
}

func validateCIEnv() error {
	var missing []string
	for _, key := range []string{"GITHUB_OUTPUT", "GITHUB_EVENT_NAME", "GITHUB_REF"} {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if os.Getenv("GITHUB_EVENT_NAME") == "pull_request" {
		for _, key := range []string{"GITHUB_BASE_SHA", "GITHUB_HEAD_SHA"} {
			if os.Getenv(key) == "" {
				missing = append(missing, key)
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables for --ci: %s", strings.Join(missing, ", "))
	}
	return nil
}

func alwaysRunOnDefaultBranchPush(repo config.Repo) bool {
	if !repo.AlwaysRunOnPushToDefaultBranch {
		return false
	}
	return os.Getenv("GITHUB_EVENT_NAME") == "push" &&
		os.Getenv("GITHUB_REF") == repo.DefaultBranchRef()
}

func writeGitHubOutput(key string, run bool) error {
	path := os.Getenv("GITHUB_OUTPUT")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	val := "false"
	if run {
		val = "true"
	}
	_, err = fmt.Fprintf(f, "%s=%s\n", key, val)
	return err
}

// LogResult prints the decision to stderr.
func LogResult(profileDir string, r Result) {
	fmt.Fprintf(os.Stderr, "shouldtest %s run=%v: %s\n", profileDir, r.Run, r.Reason)
	if len(r.MatchedPaths) > 0 {
		fmt.Fprintf(os.Stderr, "matched changes:\n")
		for _, p := range r.MatchedPaths {
			fmt.Fprintf(os.Stderr, "  %s\n", p)
		}
	}
	if len(r.WatchPaths) > 0 {
		fmt.Fprintf(os.Stderr, "watch paths (%d):\n", len(r.WatchPaths))
		for _, p := range r.WatchPaths {
			fmt.Fprintf(os.Stderr, "  %s\n", p)
		}
	}
}
