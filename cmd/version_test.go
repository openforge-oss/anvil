package cmd

import (
	"strings"
	"testing"
)

func TestVersionStringHasPrefix(t *testing.T) {
	got := versionString()
	if !strings.HasPrefix(got, "anvil ") {
		t.Fatalf("versionString() = %q, want it to start with %q", got, "anvil ")
	}
}
