package shouldtest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfile(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, profileFileName), []byte(`{
		"buildTags": ["debug"],
		"extraPaths": ["fixtures"],
		"coverageFile": "custom.out"
	}`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	p, err := LoadProfile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.CoverageFile != "custom.out" || len(p.BuildTags) != 1 {
		t.Fatalf("got %+v", p)
	}
}

func TestListTargetPattern(t *testing.T) {
	tests := []struct{ dir, want string }{
		{"./lib/proof", "./lib/proof/..."},
		{"lib/proof", "./lib/proof/..."},
	}
	for _, tc := range tests {
		got, err := listTargetPattern(tc.dir)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("listTargetPattern(%q) = %q, want %q", tc.dir, got, tc.want)
		}
	}
}

func TestOptionsForProfile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, profileFileName), []byte(`{
		"buildTags": ["a", "b"],
		"extraPaths": ["fixtures"]
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	opts, err := OptionsForProfile(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.Targets) != 1 || opts.Targets[0][len(opts.Targets[0])-4:] != "/..." {
		t.Fatalf("targets: %v", opts.Targets)
	}
	if len(opts.BuildTags) != 2 {
		t.Fatalf("build tags: %v", opts.BuildTags)
	}
}
