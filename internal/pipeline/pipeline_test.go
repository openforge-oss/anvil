package pipeline

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/openforge-oss/anvil/internal/driver"
)

// TestHelperProcess is not a real test; it is re-executed as the subprocess for
// each pipeline Step so the runner can be exercised with deterministic output
// and exit codes.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if out := os.Getenv("HELPER_OUT"); out != "" {
		fmt.Fprintln(os.Stdout, out)
	}
	code, _ := strconv.Atoi(os.Getenv("HELPER_EXIT"))
	os.Exit(code)
}

func helperStep(name, out string, exit int) driver.Step {
	return driver.Step{
		Name: name,
		Argv: []string{os.Args[0], "-test.run=TestHelperProcess"},
		Env:  []string{"GO_WANT_HELPER_PROCESS=1", "HELPER_OUT=" + out, "HELPER_EXIT=" + strconv.Itoa(exit)},
	}
}

type fakeDriver struct {
	steps map[driver.Phase][]driver.Step
}

func (fakeDriver) Name() string { return "fake" }

func (f fakeDriver) Steps(p driver.Phase, _ driver.BuildOptions) ([]driver.Step, bool) {
	s := f.steps[p]
	return s, len(s) > 0
}

func (fakeDriver) Classify(_ driver.Phase, code int, _ []byte) driver.Status {
	if code == 0 {
		return driver.OK
	}
	return driver.Failed
}

func (fakeDriver) Artifacts(_ driver.Phase, _ []byte) []string { return nil }

func TestRunFailFastAndSkip(t *testing.T) {
	d := fakeDriver{steps: map[driver.Phase][]driver.Step{
		driver.Deps:  {helperStep("deps", "deps-line", 0)},
		driver.Test:  {helperStep("test", "boom", 1)},
		driver.Build: {helperStep("build", "should-not-run", 0)},
	}}

	events := make(chan Event, 128)
	res := Run(context.Background(), t.TempDir(), d, driver.BuildOptions{}, events)

	done := map[int]driver.Status{}
	started := map[int]bool{}
	var lines []string
	for e := range events {
		switch e.Kind {
		case KindStepStart:
			started[e.Index] = true
		case KindStepDone:
			done[e.Index] = e.Status
		case KindLine:
			lines = append(lines, e.Line)
		}
	}

	if res.Success {
		t.Error("expected failure")
	}
	// Plan order: 0 deps, 1 analyze (skipped), 2 test, 3 build.
	if done[0] != driver.OK {
		t.Errorf("deps status = %v, want ok", done[0])
	}
	if done[1] != driver.Skipped {
		t.Errorf("analyze status = %v, want skipped", done[1])
	}
	if done[2] != driver.Failed {
		t.Errorf("test status = %v, want failed", done[2])
	}
	if started[3] {
		t.Error("build should not start after a failure (fail-fast)")
	}
	if !contains(lines, "deps-line") || !contains(lines, "boom") {
		t.Errorf("missing streamed lines, got %v", lines)
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if strings.TrimSpace(s) == want {
			return true
		}
	}
	return false
}
