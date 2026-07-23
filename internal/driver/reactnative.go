package driver

import (
	"os"
	"path/filepath"
	"strings"
)

type ReactNative struct{ base }

func (ReactNative) Name() string { return "react-native" }

func (r ReactNative) Steps(phase Phase, opts BuildOptions) ([]Step, bool) {
	switch phase {
	case Deps:
		steps := []Step{jsInstall(r.root)}
		if fileExists(filepath.Join(r.root, "ios", "Podfile")) {
			steps = append(steps, Step{Name: "pod install", Argv: []string{"pod", "install"}, Dir: "ios"})
		}
		return steps, true
	case Analyze:
		if !hasEslint(r.root) {
			return nil, false
		}
		return []Step{{Name: "eslint", Argv: []string{"npx", "eslint", "."}}}, true
	case Test:
		if !hasJest(r.root) {
			return nil, false
		}
		return []Step{{Name: "jest", Argv: []string{"npx", "jest", "--ci"}}}, true
	case Build:
		if target(opts, "android") == "ios" {
			ws := firstGlob(filepath.Join(r.root, "ios"), "*.xcworkspace")
			if ws == "" {
				return nil, false
			}
			scheme := strings.TrimSuffix(ws, ".xcworkspace")
			if opts.Flavor != "" {
				scheme = opts.Flavor
			}
			return []Step{{
				Name: "xcodebuild build",
				Dir:  "ios",
				Argv: []string{"xcodebuild", "build", "-workspace", ws, "-scheme", scheme,
					"-destination", "generic/platform=iOS Simulator", "CODE_SIGNING_ALLOWED=NO"},
			}}, true
		}
		variant := "Debug"
		if opts.Release {
			variant = "Release"
		}
		task := "assemble" + title(opts.Flavor) + variant
		return []Step{{Name: "gradlew " + task, Dir: "android", Argv: []string{gradlew(), task}}}, true
	case Sign:
		if opts.Signing.ExportPlist == "" {
			return nil, false
		}
		ws := firstGlob(filepath.Join(r.root, "ios"), "*.xcworkspace")
		if ws == "" {
			return nil, false
		}
		scheme := strings.TrimSuffix(ws, ".xcworkspace")
		if opts.Flavor != "" {
			scheme = opts.Flavor
		}
		archive := "build/anvil/" + scheme + ".xcarchive"
		archiveArgs := []string{"xcodebuild", "-workspace", ws, "-scheme", scheme,
			"-configuration", "Release", "-archivePath", archive, "archive", "-allowProvisioningUpdates"}
		if opts.Signing.TeamID != "" {
			archiveArgs = append(archiveArgs, "DEVELOPMENT_TEAM="+opts.Signing.TeamID)
		}
		return []Step{
			{Name: "xcodebuild archive", Dir: "ios", Argv: archiveArgs},
			{Name: "xcodebuild -exportArchive", Dir: "ios", Argv: []string{"xcodebuild", "-exportArchive",
				"-archivePath", archive, "-exportPath", "build/anvil/ipa", "-exportOptionsPlist", opts.Signing.ExportPlist}},
		}, true
	}
	return nil, false
}

func jsInstall(root string) Step {
	switch {
	case fileExists(filepath.Join(root, "yarn.lock")):
		return Step{Name: "yarn install", Argv: []string{"yarn", "install", "--frozen-lockfile"}}
	case fileExists(filepath.Join(root, "pnpm-lock.yaml")):
		return Step{Name: "pnpm install", Argv: []string{"pnpm", "install", "--frozen-lockfile"}}
	case fileExists(filepath.Join(root, "package-lock.json")):
		return Step{Name: "npm ci", Argv: []string{"npm", "ci"}}
	default:
		return Step{Name: "npm install", Argv: []string{"npm", "install"}}
	}
}

func hasEslint(root string) bool {
	for _, n := range []string{
		".eslintrc", ".eslintrc.js", ".eslintrc.cjs", ".eslintrc.json",
		".eslintrc.yaml", ".eslintrc.yml", "eslint.config.js",
		"eslint.config.mjs", "eslint.config.cjs",
	} {
		if fileExists(filepath.Join(root, n)) {
			return true
		}
	}
	return false
}

func hasJest(root string) bool {
	for _, n := range []string{"jest.config.js", "jest.config.ts", "jest.config.cjs", "jest.config.mjs", "jest.config.json"} {
		if fileExists(filepath.Join(root, n)) {
			return true
		}
	}
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	return err == nil && strings.Contains(string(data), "\"jest\"")
}
