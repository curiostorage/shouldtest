package shouldtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ChangedFiles returns repo-relative paths changed between base and head.
// When base is empty, uses merge-base with defaultBranch and HEAD. Local uncommitted
// and untracked files are included only when base was resolved via merge-base.
func ChangedFiles(base, head, defaultBranch string) ([]string, error) {
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	localMode := base == ""
	if localMode {
		out, err := exec.Command("git", "merge-base", defaultBranch, "HEAD").Output()
		if err != nil {
			return nil, err
		}
		base = strings.TrimSpace(string(out))
	}
	if head == "" {
		head = "HEAD"
	}

	committed, err := gitLines("git", "diff", "--name-only", base, head)
	if err != nil {
		return nil, err
	}

	var extra [][]string
	if localMode {
		staged, _ := gitLines("git", "diff", "--name-only", "--cached")
		unstaged, _ := gitLines("git", "diff", "--name-only")
		untracked, _ := gitLines("git", "ls-files", "--others", "--exclude-standard")
		extra = [][]string{staged, unstaged, untracked}
	}

	seen := make(map[string]bool)
	var paths []string
	for _, batch := range append([][]string{committed}, extra...) {
		for _, p := range batch {
			p = normalizePath(p)
			if p == "" || seen[p] {
				continue
			}
			if _, err := os.Stat(p); err != nil {
				continue
			}
			seen[p] = true
			paths = append(paths, p)
		}
	}
	return paths, nil
}

func gitLines(cmd string, args ...string) ([]string, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return nil, err
	}
	var lines []string
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func normalizePath(p string) string {
	p = filepath.ToSlash(strings.TrimSpace(p))
	p = strings.TrimPrefix(p, "./")
	return p
}
