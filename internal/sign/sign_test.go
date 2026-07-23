package sign

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeF(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateKeystoreAndKeyProperties(t *testing.T) {
	if _, err := exec.LookPath("keytool"); err != nil {
		t.Skip("keytool not available")
	}
	dir := t.TempDir()
	ks := Keystore{
		Path: filepath.Join(dir, "ks.jks"), Alias: "upload",
		StorePass: "testpass123", KeyPass: "testpass123", DName: "CN=test, O=test, C=US",
	}
	if err := GenerateKeystore(ks); err != nil {
		t.Fatal(err)
	}
	if !fileExists(ks.Path) {
		t.Fatal("keystore not created")
	}
	out, err := exec.Command("keytool", "-list", "-keystore", ks.Path, "-storepass", ks.StorePass, "-storetype", "PKCS12").CombinedOutput()
	if err != nil {
		t.Fatalf("keytool -list failed: %v: %s", err, out)
	}
	if !strings.Contains(string(out), "upload") {
		t.Errorf("alias not in keystore listing: %s", out)
	}
	if err := GenerateKeystore(ks); err != nil {
		t.Fatalf("second GenerateKeystore (should be idempotent): %v", err)
	}
	path, err := WriteKeyProperties(dir, ks)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	for _, want := range []string{"storeFile=", "keyAlias=upload", "storePassword=testpass123"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("key.properties missing %q", want)
		}
	}
}

func TestAndroidLayout(t *testing.T) {
	dir := t.TempDir()
	writeF(t, filepath.Join(dir, "android/app/build.gradle"), "")
	gr, app, kts, ok := AndroidLayout(dir)
	if !ok || kts || gr != filepath.Join(dir, "android") || app != filepath.Join(dir, "android/app/build.gradle") {
		t.Errorf("flutter layout: gr=%s app=%s kts=%v ok=%v", gr, app, kts, ok)
	}

	dir2 := t.TempDir()
	writeF(t, filepath.Join(dir2, "app/build.gradle.kts"), "")
	gr2, _, kts2, ok2 := AndroidLayout(dir2)
	if !ok2 || !kts2 || gr2 != dir2 {
		t.Errorf("native kts layout: gr=%s kts=%v ok=%v", gr2, kts2, ok2)
	}
}

func TestWireAndroidGradleGroovyIdempotent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "build.gradle")
	writeF(t, f, "plugins { id 'com.android.application' }\nandroid { namespace 'com.x' }\n")

	wired, err := WireAndroidGradle(f, false)
	if err != nil || !wired {
		t.Fatalf("wire: wired=%v err=%v", wired, err)
	}
	data, _ := os.ReadFile(f)
	if !strings.Contains(string(data), "signingConfigs") || !strings.Contains(string(data), gradleMarker) {
		t.Errorf("gradle not wired:\n%s", data)
	}

	if _, err := WireAndroidGradle(f, false); err != nil {
		t.Fatal(err)
	}
	data2, _ := os.ReadFile(f)
	if string(data) != string(data2) {
		t.Errorf("not idempotent: file changed on the second wire")
	}
}

func TestWireAndroidGradleKotlinSkipped(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "build.gradle.kts")
	writeF(t, f, "plugins { id(\"com.android.application\") }\n")
	if wired, err := WireAndroidGradle(f, true); err != nil || wired {
		t.Errorf("kotlin DSL should not auto-wire: wired=%v err=%v", wired, err)
	}
}

func TestEnsureGitignoreIdempotent(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 2; i++ {
		if err := EnsureGitignore(dir, []string{"key.properties", "*.jks"}); err != nil {
			t.Fatal(err)
		}
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if n := strings.Count(string(data), "key.properties"); n != 1 {
		t.Errorf("duplicate gitignore entries (%d):\n%s", n, data)
	}
}

func TestWriteExportOptions(t *testing.T) {
	dir := t.TempDir()
	path, err := WriteExportOptions(dir, "TEAM123", "ad-hoc")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	for _, want := range []string{"<string>ad-hoc</string>", "<string>TEAM123</string>", "automatic"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("ExportOptions missing %q", want)
		}
	}
}
