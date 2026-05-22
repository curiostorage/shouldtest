package shouldtest

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/curiostorage/shouldtest/config"
)

type listPkg struct {
	Dir string `json:"Dir"`
}

// DependencyDirs runs go list for targets and returns in-repo package directories.
// Includes both direct targets and their -deps -test closure.
func DependencyDirs(targets []string, tags []string, moduleRoot string) ([]string, error) {
	seen := make(map[string]bool)
	var dirs []string

	for _, withDeps := range []bool{false, true} {
		listDirs, err := goListDirs(targets, tags, withDeps)
		if err != nil {
			return nil, err
		}
		for _, d := range listDirs {
			if !underModule(d, moduleRoot) {
				continue
			}
			rel, err := filepath.Rel(moduleRoot, d)
			if err != nil {
				continue
			}
			rel = filepath.ToSlash(rel)
			if rel == "." {
				continue
			}
			if seen[rel] {
				continue
			}
			seen[rel] = true
			dirs = append(dirs, rel)
		}
	}
	return dirs, nil
}

func goListDirs(targets []string, tags []string, withDeps bool) ([]string, error) {
	moduleRoot, err := config.ModuleRoot()
	if err != nil {
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if wd != moduleRoot {
		if err := os.Chdir(moduleRoot); err != nil {
			return nil, err
		}
		defer func() { _ = os.Chdir(wd) }()
	}

	args := []string{"list", "-json"}
	if withDeps {
		args = append(args, "-deps", "-test")
	}
	if len(tags) > 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	args = append(args, targets...)

	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var dirs []string
	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var pkg listPkg
		if err := dec.Decode(&pkg); err != nil {
			return nil, err
		}
		if pkg.Dir != "" {
			dirs = append(dirs, filepath.Clean(pkg.Dir))
		}
	}
	return dirs, nil
}

func underModule(dir, moduleRoot string) bool {
	dir = filepath.Clean(dir)
	moduleRoot = filepath.Clean(moduleRoot)
	if dir == moduleRoot {
		return true
	}
	rel, err := filepath.Rel(moduleRoot, dir)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
