package driver

type Android struct{ base }

func (Android) Name() string { return "android" }

func (Android) Steps(phase Phase, opts BuildOptions) ([]Step, bool) {
	switch phase {
	case Deps:
		return nil, false
	case Analyze:
		return []Step{{Name: "gradlew lint", Argv: []string{gradlew(), "lint"}}}, true
	case Test:
		return []Step{{Name: "gradlew test", Argv: []string{gradlew(), "test"}}}, true
	case Build:
		task := buildTask(opts)
		return []Step{{Name: "gradlew " + task, Argv: []string{gradlew(), task}}}, true
	}
	return nil, false
}

func buildTask(opts BuildOptions) string {
	verb := "assemble"
	if opts.Target == "aab" || opts.Target == "appbundle" || opts.Target == "bundle" {
		verb = "bundle"
	}
	variant := "Debug"
	if opts.Release {
		variant = "Release"
	}
	return verb + title(opts.Flavor) + variant
}
