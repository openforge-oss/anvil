// Package pipeline runs a driver's steps: it executes each phase in order,
// streams combined output line by line, captures exit codes, applies the
// driver's classification, collects artifacts, and stops on the first failure.
package pipeline

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/openforge-oss/anvil/internal/driver"
)

// Item is one entry in the resolved plan: a concrete step, or a skipped phase
// (Step is nil).
type Item struct {
	Phase driver.Phase
	Name  string
	Step  *driver.Step
}

// Plan resolves the full ordered list of steps for a driver, inserting a skipped
// entry for any phase the driver reports as not applicable.
func Plan(d driver.Driver, opts driver.BuildOptions) []Item {
	var items []Item
	for _, ph := range driver.Phases {
		steps, ok := d.Steps(ph, opts)
		if !ok || len(steps) == 0 {
			items = append(items, Item{Phase: ph, Name: ph.String()})
			continue
		}
		for i := range steps {
			s := steps[i]
			items = append(items, Item{Phase: ph, Name: s.Name, Step: &s})
		}
	}
	return items
}

type Kind int

const (
	KindPlan Kind = iota
	KindStepStart
	KindLine
	KindStepDone
	KindDone
)

type Event struct {
	Kind      Kind
	Items     []Item
	Index     int
	Line      string
	Status    driver.Status
	ExitCode  int
	Artifacts []string
	Success   bool
}

type Result struct {
	Success   bool
	Artifacts []string
}

// Run executes the plan for driver d rooted at root, emitting events on the
// channel and closing it when finished. Callers read the channel concurrently.
func Run(ctx context.Context, root string, d driver.Driver, opts driver.BuildOptions, events chan<- Event) Result {
	defer close(events)

	items := Plan(d, opts)
	events <- Event{Kind: KindPlan, Items: items}

	var artifacts []string
	success := true

	for idx, it := range items {
		if it.Step == nil {
			events <- Event{Kind: KindStepDone, Index: idx, Status: driver.Skipped}
			continue
		}
		events <- Event{Kind: KindStepStart, Index: idx}

		code, out := runStep(ctx, root, *it.Step, func(line string) {
			events <- Event{Kind: KindLine, Index: idx, Line: line}
		})

		status := d.Classify(it.Phase, code, out)
		var arts []string
		if status == driver.OK {
			arts = d.Artifacts(it.Phase, out)
			artifacts = append(artifacts, arts...)
		}
		events <- Event{Kind: KindStepDone, Index: idx, Status: status, ExitCode: code, Artifacts: arts}

		if status == driver.Failed {
			success = false
			break
		}
	}

	res := Result{Success: success, Artifacts: artifacts}
	events <- Event{Kind: KindDone, Success: success, Artifacts: artifacts}
	return res
}

func runStep(ctx context.Context, root string, step driver.Step, emit func(string)) (int, []byte) {
	workdir := root
	if step.Dir != "" {
		workdir = filepath.Join(root, step.Dir)
	}

	name := step.Argv[0]
	if !filepath.IsAbs(name) {
		if cand := filepath.Join(workdir, name); fileExists(cand) {
			name = cand
		}
	}

	var full bytes.Buffer
	w := &lineWriter{full: &full, emit: emit}

	cmd := exec.CommandContext(ctx, name, step.Argv[1:]...)
	cmd.Dir = workdir
	if len(step.Env) > 0 {
		cmd.Env = append(os.Environ(), step.Env...)
	}
	cmd.Stdout = w
	cmd.Stderr = w

	err := cmd.Run()
	w.flush()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode(), full.Bytes()
		}
		emit("anvil: " + err.Error())
		return -1, full.Bytes()
	}
	return 0, full.Bytes()
}

// lineWriter splits combined output into lines. exec serializes writes when the
// same writer is used for Stdout and Stderr, so no locking is needed.
type lineWriter struct {
	buf  []byte
	full *bytes.Buffer
	emit func(string)
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.full.Write(p)
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		w.emit(string(bytes.TrimRight(w.buf[:i], "\r")))
		w.buf = w.buf[i+1:]
	}
	return len(p), nil
}

func (w *lineWriter) flush() {
	if len(w.buf) > 0 {
		w.emit(string(w.buf))
		w.buf = nil
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
