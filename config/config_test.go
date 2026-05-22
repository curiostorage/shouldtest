package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".github"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".github", RepoFileName), []byte(`{
		"defaultBranch": "main",
		"alwaysRunOnPushToDefaultBranch": true,
		"ciInvoke": "go run example.com/tool"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	repo, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if repo.DefaultBranch != "main" {
		t.Fatalf("defaultBranch: %q", repo.DefaultBranch)
	}
	if !repo.AlwaysRunOnPushToDefaultBranch {
		t.Fatal("expected alwaysRunOnPushToDefaultBranch")
	}
	if repo.CIInvoke != "go run example.com/tool" {
		t.Fatalf("ciInvoke: %q", repo.CIInvoke)
	}
}
