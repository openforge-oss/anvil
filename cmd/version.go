package cmd

import "runtime/debug"

// version is injected at release time via -ldflags "-X ...cmd.version=vX.Y.Z".
// For `go install`ed or `go run` builds it falls back to the module build info.
var version = "dev"

func versionString() string {
	if version != "dev" {
		return "anvil " + version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return "anvil " + v
		}
	}
	return "anvil dev"
}
