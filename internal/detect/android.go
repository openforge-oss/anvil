package detect

import "path/filepath"

func detectAndroid(dir string) *Project {
	if !hasGradleFiles(dir) {
		return nil
	}
	isApp := gradleContains(dir, "com.android.application")
	isLibrary := gradleContains(dir, "com.android.library")
	if !isApp && !isLibrary {
		return nil
	}

	subtype := "library"
	if isApp {
		subtype = "app"
	}

	var flags []string
	if gradleContains(dir, "kotlin(\"multiplatform\")") ||
		gradleContains(dir, "org.jetbrains.kotlin.multiplatform") {
		flags = append(flags, "kmp")
	}

	return &Project{Path: dir, Stack: Android, Subtype: subtype, Confidence: 0.9, Flags: flags}
}

func hasGradleFiles(dir string) bool {
	for _, n := range []string{"settings.gradle", "settings.gradle.kts", "build.gradle", "build.gradle.kts"} {
		if fileExists(filepath.Join(dir, n)) {
			return true
		}
	}
	return false
}

// gradleContains scans the build files at dir and in its immediate submodules
// for a token, covering the common case where a plugin lives in a submodule
// (for example an app/ or androidApp/ module) rather than the root build file.
func gradleContains(dir, token string) bool {
	candidates := []string{
		filepath.Join(dir, "build.gradle"),
		filepath.Join(dir, "build.gradle.kts"),
	}
	for _, pat := range []string{"*/build.gradle", "*/build.gradle.kts"} {
		if entries, err := filepath.Glob(filepath.Join(dir, pat)); err == nil {
			candidates = append(candidates, entries...)
		}
	}
	for _, c := range candidates {
		if fileContains(c, token) {
			return true
		}
	}
	return false
}
