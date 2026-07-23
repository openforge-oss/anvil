package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// webFrameworks maps a dependency to a subtype, most specific first so that, for
// example, Next wins over a bare React and SvelteKit over plain Svelte.
var webFrameworks = []struct{ dep, subtype string }{
	{"next", "next"},
	{"nuxt", "nuxt"},
	{"@sveltejs/kit", "sveltekit"},
	{"@angular/core", "angular"},
	{"@angular/cli", "angular"},
	{"astro", "astro"},
	{"@remix-run/dev", "remix"},
	{"remix", "remix"},
	{"gatsby", "gatsby"},
	{"react-scripts", "cra"},
	{"@vue/cli-service", "vue"},
	{"vue", "vue"},
	{"svelte", "svelte"},
	{"vite", "vite"},
}

func detectWeb(dir string) *Project {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	if pkg.has("react-native") || pkg.has("expo") {
		return nil
	}
	if isWorkspaceRoot(dir, pkg) {
		return nil
	}

	if fw := webFramework(pkg); fw != "" {
		return &Project{Path: dir, Stack: Web, Subtype: fw, Confidence: 0.9}
	}
	if _, ok := pkg.Scripts["build"]; ok {
		return &Project{Path: dir, Stack: Web, Subtype: "web-app", Confidence: 0.8}
	}
	if hasWebTooling(dir, pkg) {
		return &Project{Path: dir, Stack: Web, Subtype: "library", Confidence: 0.65}
	}
	return nil
}

// isWorkspaceRoot reports whether the package.json marks a monorepo root, which
// should be descended into rather than claimed. Checked before the build-script
// test because orchestrator roots often have a root build script. turbo.json
// alone is not a trigger (Turbo supports single-package repos).
func isWorkspaceRoot(dir string, pkg packageJSON) bool {
	if ws := strings.TrimSpace(string(pkg.Workspaces)); ws != "" && ws != "null" && ws != "[]" && ws != "{}" {
		return true
	}
	for _, f := range []string{"pnpm-workspace.yaml", "lerna.json", "nx.json"} {
		if fileExists(filepath.Join(dir, f)) {
			return true
		}
	}
	return false
}

func webFramework(pkg packageJSON) string {
	for _, fw := range webFrameworks {
		if pkg.has(fw.dep) {
			return fw.subtype
		}
	}
	return ""
}

func hasWebTooling(dir string, pkg packageJSON) bool {
	if _, ok := pkg.Scripts["test"]; ok {
		return true
	}
	if _, ok := pkg.Scripts["lint"]; ok {
		return true
	}
	for _, n := range []string{
		".eslintrc", ".eslintrc.js", ".eslintrc.cjs", ".eslintrc.json", "eslint.config.js",
		"jest.config.js", "jest.config.ts", "vitest.config.js", "vitest.config.ts",
	} {
		if fileExists(filepath.Join(dir, n)) {
			return true
		}
	}
	return false
}
