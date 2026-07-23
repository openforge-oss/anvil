package detect

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func summarize(root string, ps []Project) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		rel, _ := filepath.Rel(root, p.Path)
		s := filepath.ToSlash(rel) + "|" + string(p.Stack) + "|" + p.Subtype
		if len(p.Flags) > 0 {
			f := append([]string(nil), p.Flags...)
			sort.Strings(f)
			s += "|" + strings.Join(f, ",")
		}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

const flutterAppPubspec = `name: demo
environment:
  sdk: ">=3.0.0 <4.0.0"
dependencies:
  flutter:
    sdk: flutter
`

func TestScan(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "flutter app absorbs native folders",
			files: map[string]string{
				"pubspec.yaml":         flutterAppPubspec,
				"lib/main.dart":        "void main() {}",
				"android/.gitkeep":     "",
				"ios/.gitkeep":         "",
				"ios/Podfile":          "platform :ios",
				"android/build.gradle": "",
			},
			want: []string{".|flutter|app"},
		},
		{
			name:  "pure dart package is not flutter",
			files: map[string]string{"pubspec.yaml": "name: mylib\nenvironment:\n  sdk: \">=3.0.0 <4.0.0\"\ndependencies:\n  meta: ^1.0.0\n"},
			want:  []string{".|dart|package"},
		},
		{
			name: "flutter module",
			files: map[string]string{
				"pubspec.yaml": "name: mod\ndependencies:\n  flutter:\n    sdk: flutter\nflutter:\n  module:\n    androidPackage: com.example.mod\n",
			},
			want: []string{".|flutter|module"},
		},
		{
			name: "flutter plugin prunes its example app",
			files: map[string]string{
				"pubspec.yaml":          "name: plug\ndependencies:\n  flutter:\n    sdk: flutter\nflutter:\n  plugin:\n    platforms:\n      android:\n        package: com.example.plug\n",
				"android/.gitkeep":      "",
				"ios/.gitkeep":          "",
				"example/pubspec.yaml":  flutterAppPubspec,
				"example/lib/main.dart": "void main() {}",
			},
			want: []string{".|flutter|plugin"},
		},
		{
			name: "bare react native",
			files: map[string]string{
				"package.json":     `{"name":"rn","dependencies":{"react-native":"0.74.0"}}`,
				"android/.gitkeep": "",
				"ios/.gitkeep":     "",
			},
			want: []string{".|react-native|bare-rn"},
		},
		{
			name:  "expo managed without native folders",
			files: map[string]string{"package.json": `{"name":"exp","dependencies":{"expo":"51.0.0","react-native":"0.74.0"}}`},
			want:  []string{".|react-native|expo-managed"},
		},
		{
			name: "expo prebuild with native folders",
			files: map[string]string{
				"package.json":     `{"name":"exp","dependencies":{"expo":"51.0.0","react-native":"0.74.0"}}`,
				"android/.gitkeep": "",
				"ios/.gitkeep":     "",
			},
			want: []string{".|react-native|expo-bare"},
		},
		{
			name: "android app with application module",
			files: map[string]string{
				"settings.gradle":  "include ':app'",
				"app/build.gradle": "plugins { id 'com.android.application' }",
			},
			want: []string{".|android|app"},
		},
		{
			name: "android library only",
			files: map[string]string{
				"settings.gradle": "include ':lib'",
				"build.gradle":    "plugins { id 'com.android.library' }",
			},
			want: []string{".|android|library"},
		},
		{
			name: "gradle multi module is one project",
			files: map[string]string{
				"settings.gradle":   "include ':app', ':core'",
				"app/build.gradle":  "plugins { id 'com.android.application' }",
				"core/build.gradle": "plugins { id 'com.android.library' }",
			},
			want: []string{".|android|app"},
		},
		{
			name: "standalone ios workspace and project count once",
			files: map[string]string{
				"App.xcodeproj/project.pbxproj":            "",
				"App.xcworkspace/contents.xcworkspacedata": "",
				"Podfile": "platform :ios, '13.0'",
			},
			want: []string{".|ios|app"},
		},
		{
			name:  "swift package library",
			files: map[string]string{"Package.swift": "// swift-tools-version:5.9\nlet package = Package(name: \"X\", products: [.library(name: \"X\", targets: [\"X\"])])"},
			want:  []string{".|swift|library"},
		},
		{
			name:  "swift executable package",
			files: map[string]string{"Package.swift": "// swift-tools-version:5.9\nlet package = Package(name: \"cli\", targets: [.executableTarget(name: \"cli\")])"},
			want:  []string{".|swift|executable"},
		},
		{
			name: "kotlin jvm gradle is not android",
			files: map[string]string{
				"settings.gradle.kts": "rootProject.name = \"svc\"",
				"build.gradle.kts":    "plugins { kotlin(\"jvm\") version \"2.0.0\" }",
			},
			want: []string{".|kotlin|jvm"},
		},
		{
			name: "node_modules yields no phantom react native projects",
			files: map[string]string{
				"package.json":                           `{"name":"rn","dependencies":{"react-native":"0.74.0"}}`,
				"android/.gitkeep":                       "",
				"ios/.gitkeep":                           "",
				"node_modules/react-native/package.json": `{"name":"react-native","version":"0.74.0"}`,
				"node_modules/some-lib/package.json":     `{"name":"some-lib","dependencies":{"react-native":"0.74.0"}}`,
			},
			want: []string{".|react-native|bare-rn"},
		},
		{
			name: "mixed monorepo surfaces each project",
			files: map[string]string{
				"package.json":                    `{"name":"mono","private":true,"workspaces":["apps/*","packages/*"]}`,
				"apps/fl/pubspec.yaml":            flutterAppPubspec,
				"apps/fl/lib/main.dart":           "void main() {}",
				"apps/rn/package.json":            `{"name":"rn","dependencies":{"react-native":"0.74.0"}}`,
				"apps/rn/android/.gitkeep":        "",
				"apps/rn/ios/.gitkeep":            "",
				"packages/andlib/settings.gradle": "include ':lib'",
				"packages/andlib/build.gradle":    "plugins { id 'com.android.library' }",
			},
			want: []string{
				"apps/fl|flutter|app",
				"apps/rn|react-native|bare-rn",
				"packages/andlib|android|library",
			},
		},
		{
			name: "kotlin multiplatform flagged, ios app pruned",
			files: map[string]string{
				"settings.gradle.kts":                     "include(\":androidApp\", \":shared\")",
				"build.gradle.kts":                        "plugins { kotlin(\"multiplatform\") }",
				"androidApp/build.gradle.kts":             "plugins { id(\"com.android.application\") }",
				"shared/build.gradle.kts":                 "plugins { kotlin(\"multiplatform\") }",
				"iosApp/iosApp.xcodeproj/project.pbxproj": "",
			},
			want: []string{".|android|app|kmp"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := writeTree(t, tc.files)
			got, err := Scan(root, DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			gotS := summarize(root, got)
			want := append([]string(nil), tc.want...)
			sort.Strings(want)
			if !reflect.DeepEqual(gotS, want) {
				t.Errorf("Scan()\n got: %v\nwant: %v", gotS, want)
			}
		})
	}
}
