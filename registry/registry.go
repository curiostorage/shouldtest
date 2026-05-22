// Package registry validates that package shouldtest.json profiles are wired in CI workflows.
package registry

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Config controls discovery and CI workflow matching.
type Config struct {
	ProfileFileName    string   // default shouldtest.json
	Workflows          []string // repo-relative workflow paths
	WalkSkipDirs       []string // directory names to skip when walking
	IgnoreProfileUnder []string // repo-relative prefixes; ignore profiles under these paths
	CIInvoke           string   // substring that identifies a shouldtest CI invocation line
}

func (c Config) withDefaults() Config {
	if c.ProfileFileName == "" {
		c.ProfileFileName = "shouldtest.json"
	}
	if c.CIInvoke == "" {
		c.CIInvoke = "shouldtest"
	}
	if len(c.WalkSkipDirs) == 0 {
		c.WalkSkipDirs = []string{".git", "node_modules", "vendor"}
	}
	return c
}

// Validate reports errors when profiles are missing from CI or CI references unknown dirs.
func Validate(moduleRoot string, cfg Config) error {
	cfg = cfg.withDefaults()

	profiles, err := DiscoverProfileDirs(moduleRoot, cfg)
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return fmt.Errorf("no %s files found under %s", cfg.ProfileFileName, moduleRoot)
	}

	referenced, err := LoadReferencedDirs(moduleRoot, cfg)
	if err != nil {
		return err
	}
	if len(referenced) == 0 {
		return fmt.Errorf("no %q invocations found in workflows %v", cfg.CIInvoke, cfg.Workflows)
	}

	var errs []string

	var missing []string
	for _, dir := range profiles {
		if !slices.Contains(referenced, dir) {
			missing = append(missing, dir)
		}
	}
	if len(missing) > 0 {
		errs = append(errs, fmt.Sprintf("profiles not referenced in CI: %v (invoke: %s --ci <dir>)",
			missing, cfg.CIInvoke))
	}

	for _, dir := range referenced {
		path := filepath.Join(moduleRoot, dir, cfg.ProfileFileName)
		if _, err := os.Stat(path); err != nil {
			errs = append(errs, fmt.Sprintf("CI references %q but %s is missing", dir, path))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// DiscoverProfileDirs finds repo-relative directories containing cfg.ProfileFileName.
func DiscoverProfileDirs(moduleRoot string, cfg Config) ([]string, error) {
	cfg = cfg.withDefaults()
	var dirs []string
	err := filepath.WalkDir(moduleRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if slices.Contains(cfg.WalkSkipDirs, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != cfg.ProfileFileName {
			return nil
		}
		relDir, err := filepath.Rel(moduleRoot, filepath.Dir(path))
		if err != nil {
			return err
		}
		relDir = filepath.ToSlash(relDir)
		// .github/shouldtest.json is repo config, not a package profile.
		if relDir == ".github" {
			return nil
		}
		if shouldIgnoreProfile(relDir, cfg.IgnoreProfileUnder) {
			return nil
		}
		dirs = append(dirs, relDir)
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(dirs)
	return dirs, nil
}

func shouldIgnoreProfile(relDir string, ignorePrefixes []string) bool {
	for _, prefix := range ignorePrefixes {
		prefix = strings.TrimPrefix(filepath.ToSlash(prefix), "./")
		if relDir == prefix || strings.HasPrefix(relDir, prefix+"/") {
			return true
		}
	}
	return false
}

// LoadReferencedDirs reads workflow files and returns package dirs passed to shouldtest.
func LoadReferencedDirs(moduleRoot string, cfg Config) ([]string, error) {
	cfg = cfg.withDefaults()
	seen := make(map[string]bool)
	var dirs []string
	for _, wf := range cfg.Workflows {
		data, err := os.ReadFile(filepath.Join(moduleRoot, wf))
		if err != nil {
			return nil, fmt.Errorf("read workflow %s: %w", wf, err)
		}
		for _, dir := range ParseReferencedDirs(string(data), cfg.CIInvoke) {
			if seen[dir] {
				continue
			}
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	slices.Sort(dirs)
	return dirs, nil
}

// ParseReferencedDirs extracts package-dir arguments from workflow lines invoking shouldtest.
func ParseReferencedDirs(workflowYAML, ciInvoke string) []string {
	if ciInvoke == "" {
		ciInvoke = "shouldtest"
	}
	seen := make(map[string]bool)
	var dirs []string
	for line := range strings.SplitSeq(workflowYAML, "\n") {
		if !strings.Contains(line, ciInvoke) {
			continue
		}
		idx := strings.Index(line, ciInvoke)
		rest := strings.TrimSpace(line[idx+len(ciInvoke):])
		var lastPath string
		for f := range strings.FieldsSeq(rest) {
			if strings.HasPrefix(f, "-") {
				continue
			}
			lastPath = f
		}
		if lastPath == "" {
			continue
		}
		dir := strings.TrimPrefix(filepath.ToSlash(lastPath), "./")
		if seen[dir] {
			continue
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}
	slices.Sort(dirs)
	return dirs
}
