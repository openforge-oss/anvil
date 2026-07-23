// Package sign provides guided release-signing setup: it generates and wires
// signing material without committing secrets. Interactive prompting lives in
// the cmd layer; these functions take resolved values so they can be tested.
package sign

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Keystore describes an Android signing keystore.
type Keystore struct {
	Path      string
	Alias     string
	StorePass string
	KeyPass   string
	DName     string
}

// GenerateKeystore creates a PKCS12 keystore with keytool if Path does not
// already exist. It is idempotent.
func GenerateKeystore(ks Keystore) error {
	if fileExists(ks.Path) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(ks.Path), 0o755); err != nil {
		return err
	}
	args := []string{
		"-genkeypair", "-noprompt",
		"-keystore", ks.Path, "-alias", ks.Alias,
		"-keyalg", "RSA", "-keysize", "2048", "-validity", "10000",
		"-storetype", "PKCS12",
		"-storepass", ks.StorePass, "-keypass", ks.KeyPass,
		"-dname", ks.DName,
	}
	if out, err := exec.Command("keytool", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("keytool genkeypair: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// WriteKeyProperties writes key.properties into the Gradle root and returns its
// path. Gradle reads it via rootProject.file('key.properties').
func WriteKeyProperties(gradleRoot string, ks Keystore) (string, error) {
	path := filepath.Join(gradleRoot, "key.properties")
	content := fmt.Sprintf("storeFile=%s\nstorePassword=%s\nkeyAlias=%s\nkeyPassword=%s\n",
		ks.Path, ks.StorePass, ks.Alias, ks.KeyPass)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// AndroidLayout locates the app-module build file and the Gradle root for the
// common project shapes: Flutter/RN under android/, and native at the repo root.
func AndroidLayout(root string) (gradleRoot, appBuildFile string, kotlinDSL, ok bool) {
	candidates := []struct {
		gradleRoot, app string
	}{
		{filepath.Join(root, "android"), filepath.Join(root, "android", "app")},
		{root, filepath.Join(root, "app")},
		{root, root},
	}
	for _, c := range candidates {
		if f := filepath.Join(c.app, "build.gradle"); fileExists(f) {
			return c.gradleRoot, f, false, true
		}
		if f := filepath.Join(c.app, "build.gradle.kts"); fileExists(f) {
			return c.gradleRoot, f, true, true
		}
	}
	return "", "", false, false
}

const gradleMarker = "anvil-signing"

// WireAndroidGradle appends a Groovy signingConfigs block that reads
// key.properties, unless it is already present. Kotlin DSL build files are left
// untouched (wired reports false) so the caller can print manual instructions.
func WireAndroidGradle(appBuildFile string, kotlinDSL bool) (bool, error) {
	data, err := os.ReadFile(appBuildFile)
	if err != nil {
		return false, err
	}
	if strings.Contains(string(data), gradleMarker) {
		return true, nil
	}
	if kotlinDSL {
		return false, nil
	}
	block := "\n// " + gradleMarker + " (managed by anvil; reads key.properties)\n" +
		"def anvilProps = new Properties()\n" +
		"def anvilPropsFile = rootProject.file('key.properties')\n" +
		"if (anvilPropsFile.exists()) {\n" +
		"    anvilPropsFile.withInputStream { anvilProps.load(it) }\n" +
		"    android {\n" +
		"        signingConfigs {\n" +
		"            release {\n" +
		"                storeFile file(anvilProps['storeFile'])\n" +
		"                storePassword anvilProps['storePassword']\n" +
		"                keyAlias anvilProps['keyAlias']\n" +
		"                keyPassword anvilProps['keyPassword']\n" +
		"            }\n" +
		"        }\n" +
		"        buildTypes { release { signingConfig signingConfigs.release } }\n" +
		"    }\n" +
		"}\n// end " + gradleMarker + "\n"
	f, err := os.OpenFile(appBuildFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return false, err
	}
	defer f.Close()
	if _, err := f.WriteString(block); err != nil {
		return false, err
	}
	return true, nil
}

// WriteExportOptions writes an iOS ExportOptions.plist for automatic signing and
// returns its path.
func WriteExportOptions(root, teamID, method string) (string, error) {
	path := filepath.Join(root, "ExportOptions.plist")
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>method</key>
	<string>%s</string>
	<key>teamID</key>
	<string>%s</string>
	<key>signingStyle</key>
	<string>automatic</string>
</dict>
</plist>
`, method, teamID)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// EnsureGitignore appends any missing patterns to the project .gitignore under
// an anvil header. It is idempotent.
func EnsureGitignore(root string, patterns []string) error {
	path := filepath.Join(root, ".gitignore")
	existing, _ := os.ReadFile(path)
	lines := map[string]bool{}
	for _, l := range strings.Split(string(existing), "\n") {
		lines[strings.TrimSpace(l)] = true
	}
	var add []string
	for _, p := range patterns {
		if !lines[p] {
			add = append(add, p)
		}
	}
	if len(add) == 0 {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n# anvil signing secrets\n" + strings.Join(add, "\n") + "\n")
	return err
}

// SecretPatterns are the signing files that must never be committed.
var SecretPatterns = []string{"key.properties", "*.jks", "*.keystore", "*.p12", "*.p8", "ExportOptions.plist"}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
