package detect

import "path/filepath"

func detectSwift(dir string) *Project {
	path := filepath.Join(dir, "Package.swift")
	if !fileExists(path) {
		return nil
	}
	subtype := "library"
	if fileContains(path, ".executable") {
		subtype = "executable"
	}
	return &Project{Path: dir, Stack: Swift, Subtype: subtype, Confidence: 0.8}
}
