package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/openforge-oss/anvil/internal/driver"
	"github.com/openforge-oss/anvil/internal/pipeline"
)

func TestPlainSuccess(t *testing.T) {
	items := []pipeline.Item{
		{Phase: driver.Deps, Name: "install", Step: &driver.Step{}},
		{Phase: driver.Analyze, Name: "analyze"},
	}
	events := make(chan pipeline.Event, 16)
	events <- pipeline.Event{Kind: pipeline.KindPlan, Items: items}
	events <- pipeline.Event{Kind: pipeline.KindStepStart, Index: 0}
	events <- pipeline.Event{Kind: pipeline.KindLine, Index: 0, Line: "hello"}
	events <- pipeline.Event{Kind: pipeline.KindStepDone, Index: 0, Status: driver.OK, Artifacts: []string{"out.apk"}}
	events <- pipeline.Event{Kind: pipeline.KindStepDone, Index: 1, Status: driver.Skipped}
	events <- pipeline.Event{Kind: pipeline.KindDone, Success: true}
	close(events)

	var buf bytes.Buffer
	if !Plain(events, &buf) {
		t.Fatal("want success")
	}
	for _, want := range []string{"==> deps: install", "hello", "ok", "artifact: out.apk", "analyze: skipped"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q in:\n%s", want, buf.String())
		}
	}
}

func TestPlainFailure(t *testing.T) {
	items := []pipeline.Item{{Phase: driver.Test, Name: "test", Step: &driver.Step{}}}
	events := make(chan pipeline.Event, 8)
	events <- pipeline.Event{Kind: pipeline.KindPlan, Items: items}
	events <- pipeline.Event{Kind: pipeline.KindStepStart, Index: 0}
	events <- pipeline.Event{Kind: pipeline.KindStepDone, Index: 0, Status: driver.Failed, ExitCode: 1}
	events <- pipeline.Event{Kind: pipeline.KindDone, Success: false}
	close(events)

	var buf bytes.Buffer
	if Plain(events, &buf) {
		t.Fatal("want failure")
	}
	if !strings.Contains(buf.String(), "failed (exit 1)") {
		t.Errorf("missing failure line in:\n%s", buf.String())
	}
}
