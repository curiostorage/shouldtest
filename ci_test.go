package shouldtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/curiostorage/shouldtest/config"
)

func setCIEnv(t *testing.T, outPath string, eventName string) {
	t.Helper()
	t.Setenv("GITHUB_OUTPUT", outPath)
	t.Setenv("GITHUB_EVENT_NAME", eventName)
	t.Setenv("GITHUB_REF", "refs/heads/feature")
}

func TestValidateCIEnv_pullRequest(t *testing.T) {
	setCIEnv(t, "/tmp/out", "pull_request")
	t.Setenv("GITHUB_BASE_SHA", "abc")
	t.Setenv("GITHUB_HEAD_SHA", "def")
	if err := validateCIEnv(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCIEnv_pullRequestMissingSHA(t *testing.T) {
	setCIEnv(t, "/tmp/out", "pull_request")
	t.Setenv("GITHUB_BASE_SHA", "")
	t.Setenv("GITHUB_HEAD_SHA", "")
	if err := validateCIEnv(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunCI_mainPushAlwaysRuns(t *testing.T) {
	dir := t.TempDir()
	profileDir := filepath.Join(dir, "pkg")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profileDir, profileFileName), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "output")
	if err := os.WriteFile(outPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outPath)
	t.Setenv("GITHUB_EVENT_NAME", "push")
	t.Setenv("GITHUB_REF", "refs/heads/main")

	res, err := RunCI(CIConfig{
		ProfileDir: profileDir,
		Repo: config.Repo{
			DefaultBranch:                  "main",
			AlwaysRunOnPushToDefaultBranch: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Run {
		t.Fatal("expected run on main push")
	}
	out, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "run=true\n" {
		t.Fatalf("GITHUB_OUTPUT: %q", out)
	}
}
