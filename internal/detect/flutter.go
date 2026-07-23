package detect

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type pubspecDoc struct {
	Dependencies map[string]any `yaml:"dependencies"`
	Flutter      map[string]any `yaml:"flutter"`
}

func detectFlutterOrDart(dir string) *Project {
	path := filepath.Join(dir, "pubspec.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var doc pubspecDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return &Project{Path: dir, Stack: Dart, Subtype: "package", Confidence: 0.5}
	}

	_, hasFlutterDep := doc.Dependencies["flutter"]
	isFlutter := hasFlutterDep || doc.Flutter != nil
	if !isFlutter {
		return &Project{Path: dir, Stack: Dart, Subtype: "package", Confidence: 0.9}
	}

	subtype := "app"
	if doc.Flutter != nil {
		if _, ok := doc.Flutter["module"]; ok {
			subtype = "module"
		} else if _, ok := doc.Flutter["plugin"]; ok {
			subtype = "plugin"
		}
	}
	if subtype == "app" && !isRunnableFlutterApp(dir) {
		subtype = "package"
	}
	return &Project{Path: dir, Stack: Flutter, Subtype: subtype, Confidence: 0.95}
}

func isRunnableFlutterApp(dir string) bool {
	return dirExists(filepath.Join(dir, "android")) ||
		dirExists(filepath.Join(dir, "ios")) ||
		fileExists(filepath.Join(dir, "lib", "main.dart"))
}
