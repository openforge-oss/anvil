package driver

import "strings"

type Swift struct{ base }

func (Swift) Name() string { return "swift" }

func (Swift) Steps(phase Phase, opts BuildOptions) ([]Step, bool) {
	switch phase {
	case Deps:
		return []Step{{Name: "swift package resolve", Argv: []string{"swift", "package", "resolve"}}}, true
	case Analyze:
		return nil, false
	case Test:
		return []Step{{Name: "swift test", Argv: []string{"swift", "test"}}}, true
	case Build:
		args := []string{"build"}
		if opts.Release {
			args = append(args, "-c", "release")
		}
		return []Step{{Name: "swift " + strings.Join(args, " "), Argv: append([]string{"swift"}, args...)}}, true
	}
	return nil, false
}
