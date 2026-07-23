// Package tui renders pipeline events, either as an interactive Bubble Tea view
// or as plain line-oriented output for non-TTY and CI use.
package tui

import (
	"fmt"
	"io"

	"github.com/openforge-oss/anvil/internal/driver"
	"github.com/openforge-oss/anvil/internal/pipeline"
)

// Plain consumes pipeline events and writes line-oriented output. It returns
// whether the run succeeded.
func Plain(events <-chan pipeline.Event, out io.Writer) bool {
	var items []pipeline.Item
	success := true
	for e := range events {
		switch e.Kind {
		case pipeline.KindPlan:
			items = e.Items
		case pipeline.KindStepStart:
			fmt.Fprintf(out, "==> %s: %s\n", items[e.Index].Phase, items[e.Index].Name)
		case pipeline.KindLine:
			fmt.Fprintf(out, "    %s\n", e.Line)
		case pipeline.KindStepDone:
			switch e.Status {
			case driver.Skipped:
				fmt.Fprintf(out, "--- %s: skipped\n", items[e.Index].Phase)
			case driver.OK:
				fmt.Fprintln(out, "    ok")
				for _, a := range e.Artifacts {
					fmt.Fprintf(out, "    artifact: %s\n", a)
				}
			case driver.Failed:
				fmt.Fprintf(out, "    failed (exit %d)\n", e.ExitCode)
				success = false
			}
		case pipeline.KindDone:
			success = e.Success
		}
	}
	return success
}
