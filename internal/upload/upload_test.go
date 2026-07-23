package upload

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileCredExplicitOutsideRepo(t *testing.T) {
	repo := t.TempDir()
	cred := filepath.Join(t.TempDir(), "key.p8")
	if err := os.WriteFile(cred, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, cleanup, err := FileCred(repo, cred, "NOPE_PATH", "NOPE_B64", "k.p8")
	defer cleanup()
	if err != nil || got != cred {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestFileCredRefusesInsideRepo(t *testing.T) {
	repo := t.TempDir()
	cred := filepath.Join(repo, "key.p8")
	if err := os.WriteFile(cred, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, cleanup, err := FileCred(repo, cred, "NOPE_PATH", "NOPE_B64", "k.p8")
	defer cleanup()
	if err == nil || !strings.Contains(err.Error(), "inside the repo") {
		t.Fatalf("expected in-repo refusal, got %v", err)
	}
}

func TestFileCredBase64(t *testing.T) {
	repo := t.TempDir()
	t.Setenv("TEST_CRED_B64", base64.StdEncoding.EncodeToString([]byte("secret")))
	got, cleanup, err := FileCred(repo, "", "TEST_CRED_PATH", "TEST_CRED_B64", "k.p8")
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(got)
	if string(data) != "secret" {
		t.Errorf("decoded %q, want secret", data)
	}
	if inside(repo, got) {
		t.Error("temp credential should be outside the repo")
	}
}

func TestFileCredMissing(t *testing.T) {
	_, cleanup, err := FileCred(t.TempDir(), "", "NOPE_PATH", "NOPE_B64", "k.p8")
	defer cleanup()
	if err == nil {
		t.Fatal("expected an error for a missing credential")
	}
}

func TestValidate(t *testing.T) {
	if err := (IOS{}).Validate(); err == nil {
		t.Error("empty iOS should not validate")
	}
	if err := (IOS{KeyID: "k", IssuerID: "i", KeyPath: "p", Artifact: "a.ipa"}).Validate(); err != nil {
		t.Errorf("valid iOS: %v", err)
	}
	if err := (Android{}).Validate(); err == nil {
		t.Error("empty Android should not validate")
	}
	if err := (Android{ServiceAccount: "sa.json", Package: "com.x", Artifact: "a.aab"}).Validate(); err != nil {
		t.Errorf("valid Android: %v", err)
	}
}

func TestDescribe(t *testing.T) {
	a := strings.Join(Android{ServiceAccount: "sa", Package: "com.x", Track: "beta", Artifact: "a.aab"}.Describe(), "\n")
	if !strings.Contains(a, "com.x") || !strings.Contains(a, "beta") {
		t.Errorf("android describe: %s", a)
	}
	i := strings.Join(IOS{KeyID: "K", IssuerID: "I", KeyPath: "p", Artifact: "a.ipa"}.Describe(), "\n")
	if !strings.Contains(i, "altool") {
		t.Errorf("ios describe: %s", i)
	}
}
