package driver

import "bytes"

type Go struct{ base }

func (Go) Name() string { return "go" }

func (Go) Steps(phase Phase, _ BuildOptions) ([]Step, bool) {
	switch phase {
	case Deps:
		return []Step{{Name: "go mod download", Argv: []string{"go", "mod", "download"}}}, true
	case Analyze:
		return []Step{
			{Name: "go vet", Argv: []string{"go", "vet", "./..."}},
			{Name: "gofmt -l", Argv: []string{"gofmt", "-l", "."}},
		}, true
	case Test:
		return []Step{{Name: "go test", Argv: []string{"go", "test", "./..."}}}, true
	case Build:
		return []Step{{Name: "go build", Argv: []string{"go", "build", "./..."}}}, true
	}
	return nil, false
}

// Classify treats a non-empty Analyze output as a failure: gofmt -l exits 0 even
// when it lists unformatted files, so exit code alone is not enough.
func (Go) Classify(phase Phase, exitCode int, output []byte) Status {
	if phase == Analyze {
		if exitCode != 0 || len(bytes.TrimSpace(output)) > 0 {
			return Failed
		}
		return OK
	}
	if exitCode == 0 {
		return OK
	}
	return Failed
}
