package detect

import "path/filepath"

func detectGo(dir string) *Project {
	if !fileExists(filepath.Join(dir, "go.mod")) {
		return nil
	}
	subtype := "library"
	if hasGoMain(dir) {
		subtype = "app"
	}
	return &Project{Path: dir, Stack: Go, Subtype: subtype, Confidence: 0.95}
}

// hasGoMain reports whether a main package exists in the module root or under
// cmd/*, without invoking the toolchain.
func hasGoMain(dir string) bool {
	patterns := []string{
		filepath.Join(dir, "*.go"),
		filepath.Join(dir, "cmd", "*", "*.go"),
	}
	for _, pat := range patterns {
		files, _ := filepath.Glob(pat)
		for _, f := range files {
			if fileContains(f, "package main") {
				return true
			}
		}
	}
	return false
}
