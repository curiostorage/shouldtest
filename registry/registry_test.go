package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	root := t.TempDir()
	writeWorkflow(t, root, ".github/workflows/ci.yml", `
run: mytool --ci --verbose ./pkg/foo
`)
	if err := os.MkdirAll(filepath.Join(root, "pkg/foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg/foo/shouldtest.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		Workflows:    []string{".github/workflows/ci.yml"},
		CIInvoke:     "mytool",
		WalkSkipDirs: []string{".git"},
	}
	if err := Validate(root, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_missingCIReference(t *testing.T) {
	root := t.TempDir()
	writeWorkflow(t, root, ".github/workflows/ci.yml", "run: mytool --ci ./other\n")
	if err := os.MkdirAll(filepath.Join(root, "pkg/foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg/foo/shouldtest.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Validate(root, Config{
		Workflows: []string{".github/workflows/ci.yml"},
		CIInvoke:  "mytool",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDiscoverProfileDirs_skipDotGithub(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{".github/shouldtest.json", "pkg/a/shouldtest.json"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, p)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, p), []byte(`{}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	dirs, err := DiscoverProfileDirs(root, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 1 || dirs[0] != "pkg/a" {
		t.Fatalf("got %v", dirs)
	}
}

func TestDiscoverProfileDirs_ignorePrefix(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"pkg/a/shouldtest.json", "tools/shouldtest/shouldtest.json"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, p)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, p), []byte(`{}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	dirs, err := DiscoverProfileDirs(root, Config{
		IgnoreProfileUnder: []string{"tools/shouldtest"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 1 || dirs[0] != "pkg/a" {
		t.Fatalf("got %v", dirs)
	}
}

func TestParseReferencedDirs(t *testing.T) {
	yaml := `
      run: go run github.com/curiostorage/shouldtest/cmd/shouldtest --ci --verbose ./lib/proof
`
	got := ParseReferencedDirs(yaml, "go run github.com/curiostorage/shouldtest/cmd/shouldtest")
	if len(got) != 1 || got[0] != "lib/proof" {
		t.Fatalf("got %v", got)
	}
}

func writeWorkflow(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
