package detect

import "path/filepath"

func detectAndroid(dir string) *Project {
	hasSettings := fileExists(filepath.Join(dir, "settings.gradle")) ||
		fileExists(filepath.Join(dir, "settings.gradle.kts"))
	hasBuild := fileExists(filepath.Join(dir, "build.gradle")) ||
		fileExists(filepath.Join(dir, "build.gradle.kts"))
	if !hasSettings && !hasBuild {
		return nil
	}

	confidence := 0.6
	if hasSettings {
		confidence = 0.9
	}

	subtype := "library"
	if gradleContains(dir, "com.android.application") {
		subtype = "app"
	}

	var flags []string
	if gradleContains(dir, "kotlin(\"multiplatform\")") ||
		gradleContains(dir, "org.jetbrains.kotlin.multiplatform") {
		flags = append(flags, "kmp")
	}

	return &Project{Path: dir, Stack: Android, Subtype: subtype, Confidence: confidence, Flags: flags}
}

// gradleContains scans the build files at dir and in its immediate submodules
// for a token, covering the common case where the application plugin lives in an
// app/ module rather than the root build file.
func gradleContains(dir, token string) bool {
	candidates := []string{
		filepath.Join(dir, "build.gradle"),
		filepath.Join(dir, "build.gradle.kts"),
	}
	if entries, err := filepath.Glob(filepath.Join(dir, "*", "build.gradle")); err == nil {
		candidates = append(candidates, entries...)
	}
	if entries, err := filepath.Glob(filepath.Join(dir, "*", "build.gradle.kts")); err == nil {
		candidates = append(candidates, entries...)
	}
	for _, c := range candidates {
		if fileContains(c, token) {
			return true
		}
	}
	return false
}
