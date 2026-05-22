// Command shouldtest decides whether expensive test suites should run based on
// changed files in a PR versus the go list dependency closure defined in shouldtest.json.
//
// Repository settings live in .github/shouldtest.json (default branch, CI registry, etc.).
// Per-package settings live in <package-dir>/shouldtest.json.
//
// Usage:
//
//	go run github.com/curiostorage/shouldtest/cmd/shouldtest [flags] <package-dir>
//
// Examples:
//
//	go run github.com/curiostorage/shouldtest/cmd/shouldtest --ci --verbose ./lib/proof
//	go run github.com/curiostorage/shouldtest/cmd/shouldtest --verbose ./lib/proof
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/curiostorage/shouldtest"
	"github.com/curiostorage/shouldtest/config"
)

func main() {
	var (
		ci          bool
		verbose     bool
		jsonOut     bool
		coverageDir string
	)

	flag.BoolVar(&ci, "ci", false, "GitHub Actions mode (requires GITHUB_* env)")
	flag.BoolVar(&verbose, "verbose", false, "print decision details to stderr")
	flag.BoolVar(&jsonOut, "json", false, "print full result as JSON")
	flag.StringVar(&coverageDir, "coverage-dir", "coverage", "coverage directory for placeholder file (--ci)")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: shouldtest [flags] <package-dir>\n")
		os.Exit(2)
	}
	profileDir := args[0]

	repo, _, err := config.LoadFromModule()
	if err != nil {
		fmt.Fprintf(os.Stderr, "shouldtest: %v\n", err)
		os.Exit(1)
	}

	if ci {
		ciResult, err := shouldtest.RunCI(shouldtest.CIConfig{
			ProfileDir:  profileDir,
			CoverageDir: coverageDir,
			Verbose:     verbose,
			Repo:        repo,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "shouldtest: %v\n", err)
			os.Exit(1)
		}
		emitResult(ciResult.Run, ciResult, jsonOut)
		return
	}

	opts, err := shouldtest.OptionsForProfile(profileDir, shouldtest.Options{DefaultBranch: repo.DefaultBranch})
	if err != nil {
		fmt.Fprintf(os.Stderr, "shouldtest: %v\n", err)
		os.Exit(1)
	}
	result, err := shouldtest.Check(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shouldtest: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		shouldtest.LogResult(profileDir, result)
	}
	emitResult(result.Run, result, jsonOut)
}

func emitResult(run bool, v any, jsonOut bool) {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			fmt.Fprintf(os.Stderr, "shouldtest: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if run {
		fmt.Println("run")
	} else {
		fmt.Println("skip")
	}
}
