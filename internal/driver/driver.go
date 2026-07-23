// Package driver defines the build lifecycle contract and the per-stack drivers.
// A driver only describes what to run per phase and how to interpret the result;
// the pipeline package owns execution, streaming, and status.
package driver

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Phase int

const (
	Deps Phase = iota
	Analyze
	Test
	Build
)

var Phases = []Phase{Deps, Analyze, Test, Build}

func (p Phase) String() string {
	switch p {
	case Deps:
		return "deps"
	case Analyze:
		return "analyze"
	case Test:
		return "test"
	case Build:
		return "build"
	default:
		return "unknown"
	}
}

type Status int

const (
	Pending Status = iota
	Running
	OK
	Failed
	Skipped
)

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Running:
		return "running"
	case OK:
		return "ok"
	case Failed:
		return "failed"
	case Skipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// Step is one command in a phase. Dir is relative to the project root; empty
// means the root itself.
type Step struct {
	Name string
	Argv []string
	Dir  string
	Env  []string
}

// BuildOptions carries user choices for the Build phase. Target is
// stack-specific (for example apk, appbundle, ios); empty selects the driver
// default. Flavor selects a build flavor, product flavor, or scheme where the
// stack supports it. Release requests a release build where applicable.
type BuildOptions struct {
	Target  string
	Flavor  string
	Release bool
}

type Driver interface {
	Name() string
	Steps(phase Phase, opts BuildOptions) (steps []Step, applicable bool)
	Classify(phase Phase, exitCode int, output []byte) Status
	Artifacts(phase Phase, output []byte) []string
}

// base provides the default Classify and Artifacts so drivers only override the
// exceptions.
type base struct{ root string }

func (base) Classify(_ Phase, exitCode int, _ []byte) Status {
	if exitCode == 0 {
		return OK
	}
	return Failed
}

func (base) Artifacts(_ Phase, _ []byte) []string { return nil }

func gradlew() string {
	if runtime.GOOS == "windows" {
		return "gradlew.bat"
	}
	return "./gradlew"
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func firstGlob(dir, pattern string) string {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil || len(matches) == 0 {
		return ""
	}
	return filepath.Base(matches[0])
}
