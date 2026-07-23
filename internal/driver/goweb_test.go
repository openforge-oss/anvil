package driver

import (
	"strings"
	"testing"
)

func TestGoSteps(t *testing.T) {
	d := Go{}
	wantCmds(t, mustSteps(t, d, Deps, BuildOptions{}), "go mod download")
	wantCmds(t, mustSteps(t, d, Analyze, BuildOptions{}), "go vet ./...", "gofmt -l .")
	wantCmds(t, mustSteps(t, d, Test, BuildOptions{}), "go test ./...")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{}), "go build ./...")
}

func TestGoClassifyAnalyze(t *testing.T) {
	d := Go{}
	if d.Classify(Analyze, 0, []byte("")) != OK {
		t.Error("clean analyze should be OK")
	}
	if d.Classify(Analyze, 0, []byte("main.go\n")) != Failed {
		t.Error("gofmt listing files (exit 0) should be Failed")
	}
	if d.Classify(Test, 0, nil) != OK || d.Classify(Test, 1, nil) != Failed {
		t.Error("non-analyze phases classify by exit code")
	}
}

func TestWebSteps(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"package-lock.json": "",
		"package.json":      `{"scripts":{"build":"vite build","test":"vitest run","lint":"eslint ."}}`,
	})
	d := Web{base{root}}
	wantCmds(t, mustSteps(t, d, Deps, BuildOptions{}), "npm ci")
	wantCmds(t, mustSteps(t, d, Analyze, BuildOptions{}), "npm run lint")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{}), "npm run build")
	steps := mustSteps(t, d, Test, BuildOptions{})
	if len(steps) != 1 || strings.Join(steps[0].Argv, " ") != "npm run test" {
		t.Errorf("web test: %v", cmds(steps))
	}
}

func TestWebSkipsWhenNoScripts(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"package.json": `{"name":"x"}`})
	d := Web{base{root}}
	if _, ok := d.Steps(Build, BuildOptions{}); ok {
		t.Error("build should be skipped with no build script")
	}
	if _, ok := d.Steps(Analyze, BuildOptions{}); ok {
		t.Error("analyze should be skipped with no lint script or eslint config")
	}
}
