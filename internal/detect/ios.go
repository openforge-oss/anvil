package detect

import "path/filepath"

func detectIOS(dir string) *Project {
	if hasGlob(dir, "*.xcworkspace") || hasGlob(dir, "*.xcodeproj") {
		return &Project{Path: dir, Stack: IOS, Subtype: "app", Confidence: 0.85}
	}
	if fileExists(filepath.Join(dir, "Package.swift")) {
		subtype := "library"
		if fileContains(filepath.Join(dir, "Package.swift"), ".executable") {
			subtype = "app"
		}
		return &Project{Path: dir, Stack: IOS, Subtype: subtype, Confidence: 0.7}
	}
	if fileExists(filepath.Join(dir, "Podfile")) {
		return &Project{Path: dir, Stack: IOS, Subtype: "app", Confidence: 0.5}
	}
	return nil
}
