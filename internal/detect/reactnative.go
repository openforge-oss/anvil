package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
	Workspaces      json.RawMessage   `json:"workspaces"`
}

func detectReactNative(dir string) *Project {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	hasRN := pkg.has("react-native")
	hasExpo := pkg.has("expo")
	if !hasRN && !hasExpo {
		return nil
	}

	native := dirExists(filepath.Join(dir, "ios")) && dirExists(filepath.Join(dir, "android"))
	subtype := "bare-rn"
	switch {
	case hasExpo && !native:
		subtype = "expo-managed"
	case hasExpo && native:
		subtype = "expo-bare"
	}
	return &Project{Path: dir, Stack: ReactNative, Subtype: subtype, Confidence: 0.9}
}

func (p packageJSON) has(dep string) bool {
	if _, ok := p.Dependencies[dep]; ok {
		return true
	}
	_, ok := p.DevDependencies[dep]
	return ok
}
