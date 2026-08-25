package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/openforge-oss/anvil/internal/driver"
)

func TestSelectBuildPhases(t *testing.T) {
	available := []driver.Phase{driver.Deps, driver.Analyze, driver.Test, driver.Build}

	tests := []struct {
		name    string
		until   string
		only    string
		want    []driver.Phase
		wantErr string
	}{
		{name: "all phases by default", want: available},
		{name: "inclusive prefix", until: "analyze", want: []driver.Phase{driver.Deps, driver.Analyze}},
		{name: "one phase", only: "test", want: []driver.Phase{driver.Test}},
		{name: "conflicting selectors", until: "test", only: "analyze", wantErr: "cannot be used together"},
		{name: "unknown phase", only: "sign", wantErr: "available phases: deps, analyze, test, build"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectBuildPhases(available, tt.until, tt.only)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("selectBuildPhases() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("phases = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectBuildPhasesIncludesOptionalSign(t *testing.T) {
	available := append(append([]driver.Phase{}, driver.Phases...), driver.Sign)

	got, err := selectBuildPhases(available, "", "sign")
	if err != nil {
		t.Fatalf("selectBuildPhases() error = %v", err)
	}
	if !reflect.DeepEqual(got, []driver.Phase{driver.Sign}) {
		t.Fatalf("phases = %v, want [sign]", got)
	}
}
