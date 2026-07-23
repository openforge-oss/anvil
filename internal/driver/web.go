package driver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Web struct{ base }

func (Web) Name() string { return "web" }

func (w Web) Steps(phase Phase, _ BuildOptions) ([]Step, bool) {
	scripts := readScripts(w.root)
	switch phase {
	case Deps:
		return []Step{jsInstall(w.root)}, true
	case Analyze:
		if _, ok := scripts["lint"]; ok {
			return []Step{pmRun(w.root, "lint")}, true
		}
		if hasEslint(w.root) {
			return []Step{{Name: "npx eslint .", Argv: []string{"npx", "eslint", "."}}}, true
		}
		return nil, false
	case Test:
		if hasJest(w.root) {
			return []Step{{Name: "npx jest --ci", Argv: []string{"npx", "jest", "--ci"}}}, true
		}
		if hasVitest(w.root) {
			return []Step{{Name: "npx vitest run", Argv: []string{"npx", "vitest", "run"}}}, true
		}
		if t, ok := scripts["test"]; ok && !strings.Contains(t, "no test specified") {
			s := pmRun(w.root, "test")
			s.Env = []string{"CI=true"}
			return []Step{s}, true
		}
		return nil, false
	case Build:
		if _, ok := scripts["build"]; ok {
			return []Step{pmRun(w.root, "build")}, true
		}
		return nil, false
	}
	return nil, false
}

func readScripts(root string) map[string]string {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	_ = json.Unmarshal(data, &pkg)
	return pkg.Scripts
}

func pmRun(root, script string) Step {
	switch {
	case fileExists(filepath.Join(root, "yarn.lock")):
		return Step{Name: "yarn " + script, Argv: []string{"yarn", script}}
	case fileExists(filepath.Join(root, "pnpm-lock.yaml")):
		return Step{Name: "pnpm run " + script, Argv: []string{"pnpm", "run", script}}
	default:
		return Step{Name: "npm run " + script, Argv: []string{"npm", "run", script}}
	}
}

func hasVitest(root string) bool {
	for _, n := range []string{"vitest.config.js", "vitest.config.ts", "vitest.config.mjs", "vitest.config.cjs"} {
		if fileExists(filepath.Join(root, n)) {
			return true
		}
	}
	return false
}
