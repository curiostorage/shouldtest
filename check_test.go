package shouldtest

import (
	"sort"
	"testing"
)

func TestMatchesWatch(t *testing.T) {
	watch := []string{"lib/proof", "go.mod", "extern/filecoin-ffi"}

	tests := []struct {
		path string
		want bool
	}{
		{"lib/proof/foo.go", true},
		{"lib/proof", true},
		{"lib/proofother/foo.go", false},
		{"go.mod", true},
		{"go.sum", false},
		{"extern/filecoin-ffi/cgo/blah.c", true},
		{"tasks/pdpv0/task.go", false},
		{"README.md", false},
	}

	for _, tc := range tests {
		if got := matchesWatch(tc.path, watch); got != tc.want {
			t.Errorf("matchesWatch(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestCheck(t *testing.T) {
	watch := []string{"lib/proof", "go.mod", "extern/filecoin-ffi"}

	tests := []struct {
		name    string
		opts    Options
		watch   []string
		wantRun bool
	}{
		{
			name: "always run",
			opts: Options{
				Targets:   []string{"./lib/proof/..."},
				AlwaysRun: true,
			},
			wantRun: true,
		},
		{
			name: "matching change",
			opts: Options{
				Targets: []string{"./lib/proof/..."},
				Changed: []string{"lib/proof/verify_test.go"},
			},
			watch:   watch,
			wantRun: true,
		},
		{
			name: "unrelated change",
			opts: Options{
				Targets: []string{"./lib/proof/..."},
				Changed: []string{"tasks/pdpv0/task.go", "README.md"},
			},
			watch:   watch,
			wantRun: false,
		},
		{
			name: "go.mod change",
			opts: Options{
				Targets: []string{"./lib/proof/..."},
				Changed: []string{"go.mod"},
			},
			watch:   watch,
			wantRun: true,
		},
		{
			name: "no changes",
			opts: Options{
				Targets: []string{"./lib/proof/..."},
				Changed: []string{},
			},
			watch:   watch,
			wantRun: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got Result
			var err error
			if tc.watch != nil {
				got, err = checkWithWatch(tc.opts, tc.watch)
			} else {
				got, err = Check(tc.opts)
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Run != tc.wantRun {
				t.Fatalf("Run = %v, want %v; reason: %s", got.Run, tc.wantRun, got.Reason)
			}
		})
	}
}

func checkWithWatch(opts Options, watch []string) (Result, error) {
	if opts.AlwaysRun {
		return Result{Run: true, Reason: "always run"}, nil
	}
	changed := opts.Changed
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
			Reason:       "matched",
			MatchedPaths: matched,
			WatchPaths:   watch,
		}, nil
	}
	return Result{
		Run:        false,
		Reason:     "no match",
		WatchPaths: watch,
	}, nil
}

func TestNormalizePath(t *testing.T) {
	if got := normalizePath("./lib/proof/foo.go"); got != "lib/proof/foo.go" {
		t.Fatalf("got %q", got)
	}
}

