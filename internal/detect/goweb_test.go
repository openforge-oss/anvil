package detect

import (
	"reflect"
	"sort"
	"testing"
)

func TestScanGoWeb(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"go app", map[string]string{"go.mod": "module x\n", "main.go": "package main\nfunc main() {}"}, []string{".|go|app"}},
		{"go library", map[string]string{"go.mod": "module x\n", "lib.go": "package lib\n"}, []string{".|go|library"}},
		{"go cmd main", map[string]string{"go.mod": "module x\n", "cmd/srv/main.go": "package main\nfunc main() {}"}, []string{".|go|app"}},
		{"go.work descends to member", map[string]string{"go.work": "go 1.22\n", "a/go.mod": "module a\n", "a/main.go": "package main\nfunc main() {}"}, []string{"a|go|app"}},
		{"next app", map[string]string{"package.json": `{"dependencies":{"next":"14"},"scripts":{"build":"next build"}}`}, []string{".|web|next"}},
		{"vite app", map[string]string{"package.json": `{"devDependencies":{"vite":"5"},"scripts":{"build":"vite build"}}`}, []string{".|web|vite"}},
		{"cra app", map[string]string{"package.json": `{"dependencies":{"react-scripts":"5"},"scripts":{"build":"react-scripts build"}}`}, []string{".|web|cra"}},
		{"web library (test only)", map[string]string{"package.json": `{"scripts":{"test":"vitest run"}}`}, []string{".|web|library"}},
		{"pnpm workspace descends to web member", map[string]string{
			"package.json":        `{"name":"root"}`,
			"pnpm-workspace.yaml": "packages:\n  - app\n",
			"app/package.json":    `{"dependencies":{"vite":"5"},"scripts":{"build":"vite build"}}`,
		}, []string{"app|web|vite"}},
		{"rn app with build script stays rn", map[string]string{
			"package.json":     `{"dependencies":{"react-native":"0.74.0"},"scripts":{"build":"x"}}`,
			"android/.gitkeep": "",
			"ios/.gitkeep":     "",
		}, []string{".|react-native|bare-rn"}},
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

func TestScanBareAndTurbo(t *testing.T) {
	bare := writeTree(t, map[string]string{"package.json": `{"name":"x","dependencies":{"lodash":"4"}}`})
	if got, _ := Scan(bare, DefaultOptions()); len(got) != 0 {
		t.Errorf("bare package.json should not be claimed, got %v", summarize(bare, got))
	}
	turbo := writeTree(t, map[string]string{"turbo.json": "{}", "package.json": `{"dependencies":{"vite":"5"},"scripts":{"build":"vite build"}}`})
	got, _ := Scan(turbo, DefaultOptions())
	if s := summarize(turbo, got); len(s) != 1 || s[0] != ".|web|vite" {
		t.Errorf("single-package turbo app should be claimed as web, got %v", s)
	}
}
