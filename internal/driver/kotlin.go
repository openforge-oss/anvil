package driver

type Kotlin struct{ base }

func (Kotlin) Name() string { return "kotlin" }

func (Kotlin) Steps(phase Phase, _ BuildOptions) ([]Step, bool) {
	switch phase {
	case Test:
		return []Step{{Name: "gradlew test", Argv: []string{gradlew(), "test"}}}, true
	case Build:
		return []Step{{Name: "gradlew build", Argv: []string{gradlew(), "build"}}}, true
	}
	return nil, false
}
