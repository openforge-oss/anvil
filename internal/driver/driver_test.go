package driver

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func cmds(steps []Step) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = strings.Join(s.Argv, " ")
	}
	return out
}

func mustSteps(t *testing.T, d Driver, p Phase, opts BuildOptions) []Step {
	t.Helper()
	steps, ok := d.Steps(p, opts)
	if !ok {
		t.Fatalf("%s %s: not applicable, want applicable", d.Name(), p)
	}
	return steps
}

func wantCmds(t *testing.T, got []Step, want ...string) {
	t.Helper()
	if g := cmds(got); !reflect.DeepEqual(g, want) {
		t.Errorf("got %v, want %v", g, want)
	}
}

func TestFlutterSteps(t *testing.T) {
	d := Flutter{}
	wantCmds(t, mustSteps(t, d, Deps, BuildOptions{}), "flutter pub get")
	wantCmds(t, mustSteps(t, d, Analyze, BuildOptions{}), "flutter analyze")
	wantCmds(t, mustSteps(t, d, Test, BuildOptions{}), "flutter test")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{}), "flutter build apk --release")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{Target: "appbundle"}), "flutter build appbundle --release")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{Target: "ios"}), "flutter build ios --release --no-codesign")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{Flavor: "prod"}), "flutter build apk --release --flavor prod")
}

func TestFlutterArtifacts(t *testing.T) {
	out := []byte("Running Gradle task...\n✓ Built build/app/outputs/flutter-apk/app-release.apk (21.2MB).\n")
	got := Flutter{}.Artifacts(Build, out)
	want := []string{"build/app/outputs/flutter-apk/app-release.apk"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAndroidSteps(t *testing.T) {
	d := Android{}
	if _, ok := d.Steps(Deps, BuildOptions{}); ok {
		t.Error("android deps should be skipped")
	}
	wantCmds(t, mustSteps(t, d, Analyze, BuildOptions{}), gradlew()+" lint")
	wantCmds(t, mustSteps(t, d, Test, BuildOptions{}), gradlew()+" test")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{}), gradlew()+" assembleDebug")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{Release: true}), gradlew()+" assembleRelease")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{Target: "aab", Release: true}), gradlew()+" bundleRelease")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{Flavor: "prod"}), gradlew()+" assembleProdDebug")
}

func TestKotlinSteps(t *testing.T) {
	d := Kotlin{}
	if _, ok := d.Steps(Analyze, BuildOptions{}); ok {
		t.Error("kotlin analyze should be skipped")
	}
	wantCmds(t, mustSteps(t, d, Test, BuildOptions{}), gradlew()+" test")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{}), gradlew()+" build")
}

func TestSwiftSteps(t *testing.T) {
	d := Swift{}
	wantCmds(t, mustSteps(t, d, Deps, BuildOptions{}), "swift package resolve")
	wantCmds(t, mustSteps(t, d, Test, BuildOptions{}), "swift test")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{}), "swift build")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{Release: true}), "swift build -c release")
}

func TestReactNativeSteps(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"yarn.lock":    "",
		"ios/Podfile":  "platform :ios",
		".eslintrc.js": "module.exports = {}",
		"package.json": `{"name":"rn","jest":{},"dependencies":{"react-native":"0.74.0"}}`,
	})
	d := ReactNative{base{root}}

	wantCmds(t, mustSteps(t, d, Deps, BuildOptions{}), "yarn install --frozen-lockfile", "pod install")
	wantCmds(t, mustSteps(t, d, Analyze, BuildOptions{}), "npx eslint .")
	wantCmds(t, mustSteps(t, d, Test, BuildOptions{}), "npx jest --ci")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{}), gradlew()+" assembleDebug")
	wantCmds(t, mustSteps(t, d, Build, BuildOptions{Flavor: "prod", Release: true}), gradlew()+" assembleProdRelease")
}

func TestIOSSteps(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "App.xcworkspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFiles(t, root, map[string]string{"Podfile": "platform :ios"})
	d := IOS{base{root}}

	wantCmds(t, mustSteps(t, d, Deps, BuildOptions{}), "pod install")
	if _, ok := d.Steps(Analyze, BuildOptions{}); ok {
		t.Error("ios analyze should be skipped")
	}
	build := mustSteps(t, d, Build, BuildOptions{})
	if got := strings.Join(build[0].Argv, " "); !strings.Contains(got, "-workspace App.xcworkspace") || !strings.Contains(got, "-scheme App") || !strings.Contains(got, "CODE_SIGNING_ALLOWED=NO") {
		t.Errorf("unexpected ios build: %s", got)
	}
	flavored := mustSteps(t, d, Build, BuildOptions{Flavor: "Staging"})
	if got := strings.Join(flavored[0].Argv, " "); !strings.Contains(got, "-scheme Staging") {
		t.Errorf("flavor should override scheme: %s", got)
	}
}

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
