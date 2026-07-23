// Package detect identifies the stack(s) of a project tree from marker files.
package detect

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Stack string

const (
	Flutter     Stack = "flutter"
	Dart        Stack = "dart"
	ReactNative Stack = "react-native"
	Android     Stack = "android"
	IOS         Stack = "ios"
)

type Project struct {
	Path       string   `json:"path"`
	Stack      Stack    `json:"stack"`
	Subtype    string   `json:"subtype,omitempty"`
	Confidence float64  `json:"confidence"`
	Flags      []string `json:"flags,omitempty"`
}

type Options struct {
	MaxDepth int
	SkipDirs map[string]bool
}

func DefaultSkipDirs() map[string]bool {
	names := []string{
		"node_modules", ".git", "build", ".gradle", ".dart_tool", "Pods",
		"DerivedData", ".expo", ".idea", ".fvm", ".symlinks", "Carthage",
		".build", "out", "dist", "vendor", ".cxx",
	}
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}

func DefaultOptions() Options {
	return Options{MaxDepth: 4, SkipDirs: DefaultSkipDirs()}
}

// Scan walks root and returns one Project per detected stack root. It prunes on
// detect (it does not descend into a detected root), skips DefaultSkipDirs, and
// absorbs android/ios folders that belong to a Flutter or React Native parent.
func Scan(root string, opts Options) ([]Project, error) {
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 4
	}
	if opts.SkipDirs == nil {
		opts.SkipDirs = DefaultSkipDirs()
	}
	root = filepath.Clean(root)

	var found []Project
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if path != root && opts.SkipDirs[d.Name()] {
			return filepath.SkipDir
		}
		depth := depthOf(root, path)
		if depth > opts.MaxDepth {
			return filepath.SkipDir
		}
		if p := detectDir(path); p != nil {
			found = append(found, *p)
			return filepath.SkipDir
		}
		if depth == opts.MaxDepth {
			return filepath.SkipDir
		}
		return nil
	}
	if err := filepath.WalkDir(root, walk); err != nil {
		return nil, err
	}
	return containmentSweep(found), nil
}

func detectDir(dir string) *Project {
	if p := detectFlutterOrDart(dir); p != nil {
		return p
	}
	if p := detectReactNative(dir); p != nil {
		return p
	}
	if p := detectAndroid(dir); p != nil {
		return p
	}
	if p := detectIOS(dir); p != nil {
		return p
	}
	return nil
}

var ownedChildren = map[string]bool{
	"android": true, "ios": true, "macos": true, "windows": true,
	"linux": true, "web": true, ".android": true, ".ios": true,
}

func containmentSweep(ps []Project) []Project {
	sort.SliceStable(ps, func(i, j int) bool { return len(ps[i].Path) < len(ps[j].Path) })
	var out []Project
	for _, p := range ps {
		if p.Stack == Android || p.Stack == IOS {
			if absorbedByParent(p, out) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func absorbedByParent(child Project, kept []Project) bool {
	for _, anc := range kept {
		if anc.Stack != Flutter && anc.Stack != ReactNative {
			continue
		}
		if isDescendant(anc.Path, child.Path) && ownedChildren[filepath.Base(child.Path)] {
			return true
		}
	}
	return false
}

func depthOf(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}

func isDescendant(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, "..")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func hasGlob(dir, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	return err == nil && len(matches) > 0
}

func fileContains(path, substr string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), substr)
}
