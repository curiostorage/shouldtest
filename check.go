package shouldtest

import (
	"fmt"
	"sort"
	"strings"

	"github.com/curiostorage/shouldtest/config"
)

// Options configures whether a test target should run based on changed files.
type Options struct {
	Targets       []string
	BuildTags     []string
	ExtraPaths    []string
	Changed       []string // optional override (for tests); skips git
	BaseRef       string
	HeadRef       string
	DefaultBranch string
	AlwaysRun     bool
}

// Result describes the shouldtest decision.
type Result struct {
	Run          bool
	Reason       string
	MatchedPaths []string
	WatchPaths   []string
}

// Check decides whether tests for Targets should run given changed files.
func Check(opts Options) (Result, error) {
	if len(opts.Targets) == 0 {
		return Result{}, fmt.Errorf("shouldtest: at least one target is required")
	}

	if opts.AlwaysRun {
		return Result{Run: true, Reason: "always run"}, nil
	}

	moduleRoot, err := config.ModuleRoot()
	if err != nil {
		return Result{}, fmt.Errorf("module root: %w", err)
	}

	watch, err := watchPaths(opts, moduleRoot)
	if err != nil {
		return Result{}, err
	}

	changed := opts.Changed
	if changed == nil {
		changed, err = ChangedFiles(opts.BaseRef, opts.HeadRef, opts.DefaultBranch)
		if err != nil {
			return Result{}, fmt.Errorf("changed files: %w", err)
		}
	}
	for i, p := range changed {
		changed[i] = normalizePath(p)
	}

	if len(changed) == 0 {
		return Result{Run: false, Reason: "no changed files", WatchPaths: watch}, nil
	}

	var matched []string
	for _, ch := range changed {
		if matchesWatch(ch, watch) {
			matched = append(matched, ch)
		}
	}

	if len(matched) > 0 {
		sort.Strings(matched)
		return Result{
			Run:          true,
			Reason:       fmt.Sprintf("%d changed file(s) match watch paths", len(matched)),
			MatchedPaths: matched,
			WatchPaths:   watch,
		}, nil
	}

	return Result{
		Run:        false,
		Reason:     "no changed files match watch paths",
		WatchPaths: watch,
	}, nil
}

func watchPaths(opts Options, moduleRoot string) ([]string, error) {
	deps, err := DependencyDirs(opts.Targets, opts.BuildTags, moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("dependency dirs: %w", err)
	}

	seen := make(map[string]bool)
	var watch []string
	add := func(p string) {
		p = normalizePath(p)
		p = strings.TrimSuffix(p, "/")
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		watch = append(watch, p)
	}

	for _, d := range deps {
		add(d)
	}
	for _, p := range opts.ExtraPaths {
		add(p)
	}

	sort.Strings(watch)
	return watch, nil
}

func matchesWatch(changed string, watch []string) bool {
	changed = normalizePath(changed)
	for _, w := range watch {
		w = normalizePath(w)
		if changed == w {
			return true
		}
		if strings.HasPrefix(changed, w+"/") {
			return true
		}
	}
	return false
}
