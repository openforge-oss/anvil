package driver

import "strings"

type Flutter struct{ base }

func (Flutter) Name() string { return "flutter" }

func (Flutter) Steps(phase Phase, opts BuildOptions) ([]Step, bool) {
	switch phase {
	case Deps:
		return []Step{{Name: "flutter pub get", Argv: []string{"flutter", "pub", "get"}}}, true
	case Analyze:
		return []Step{{Name: "flutter analyze", Argv: []string{"flutter", "analyze"}}}, true
	case Test:
		return []Step{{Name: "flutter test", Argv: []string{"flutter", "test"}}}, true
	case Build:
		var args []string
		switch target(opts, "apk") {
		case "apk":
			args = []string{"build", "apk", "--release"}
		case "appbundle", "aab":
			args = []string{"build", "appbundle", "--release"}
		case "ios":
			args = []string{"build", "ios", "--release", "--no-codesign"}
		default:
			return nil, false
		}
		if opts.Flavor != "" {
			args = append(args, "--flavor", opts.Flavor)
		}
		return []Step{{Name: "flutter " + strings.Join(args, " "), Argv: append([]string{"flutter"}, args...)}}, true
	case Sign:
		if opts.Signing.ExportPlist == "" {
			return nil, false
		}
		args := []string{"build", "ipa", "--export-options-plist", opts.Signing.ExportPlist}
		if opts.Flavor != "" {
			args = append(args, "--flavor", opts.Flavor)
		}
		return []Step{{Name: "flutter " + strings.Join(args, " "), Argv: append([]string{"flutter"}, args...)}}, true
	}
	return nil, false
}

func (Flutter) Artifacts(phase Phase, output []byte) []string {
	if phase != Build {
		return nil
	}
	var paths []string
	for _, line := range strings.Split(string(output), "\n") {
		i := strings.Index(line, "Built ")
		if i < 0 {
			continue
		}
		p := strings.TrimSpace(line[i+len("Built "):])
		if j := strings.Index(p, " ("); j >= 0 {
			p = p[:j]
		}
		if p != "" {
			paths = append(paths, strings.TrimSuffix(p, "."))
		}
	}
	return paths
}

func target(opts BuildOptions, fallback string) string {
	if opts.Target == "" {
		return fallback
	}
	return opts.Target
}
