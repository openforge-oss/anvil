package driver

import (
	"path/filepath"
	"strings"
)

type IOS struct{ base }

func (IOS) Name() string { return "ios" }

func (d IOS) Steps(phase Phase, opts BuildOptions) ([]Step, bool) {
	switch phase {
	case Deps:
		if fileExists(filepath.Join(d.root, "Podfile")) {
			return []Step{{Name: "pod install", Argv: []string{"pod", "install"}}}, true
		}
		return nil, false
	case Analyze:
		return nil, false
	case Test:
		flag, value, scheme := d.container(opts)
		if scheme == "" {
			return nil, false
		}
		return []Step{{
			Name: "xcodebuild test",
			Argv: []string{"xcodebuild", "test", flag, value, "-scheme", scheme,
				"-destination", "platform=iOS Simulator,name=iPhone 15"},
		}}, true
	case Build:
		flag, value, scheme := d.container(opts)
		if scheme == "" {
			return nil, false
		}
		return []Step{{
			Name: "xcodebuild build",
			Argv: []string{"xcodebuild", "build", flag, value, "-scheme", scheme,
				"-destination", "generic/platform=iOS Simulator", "CODE_SIGNING_ALLOWED=NO"},
		}}, true
	}
	return nil, false
}

// container prefers a CocoaPods workspace over a bare project and derives the
// scheme from its name, the common single-scheme convention.
func (d IOS) container(opts BuildOptions) (flag, value, scheme string) {
	if ws := firstGlob(d.root, "*.xcworkspace"); ws != "" {
		flag, value, scheme = "-workspace", ws, strings.TrimSuffix(ws, ".xcworkspace")
	} else if pj := firstGlob(d.root, "*.xcodeproj"); pj != "" {
		flag, value, scheme = "-project", pj, strings.TrimSuffix(pj, ".xcodeproj")
	} else {
		return "", "", ""
	}
	if opts.Flavor != "" {
		scheme = opts.Flavor
	}
	return flag, value, scheme
}
