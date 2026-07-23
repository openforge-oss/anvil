package upload

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type IOS struct {
	KeyID    string
	IssuerID string
	KeyPath  string
	Artifact string
}

func (IOS) Platform() Platform { return PlatformIOS }

func (u IOS) Validate() error {
	switch {
	case u.KeyID == "" || u.IssuerID == "":
		return fmt.Errorf("iOS upload needs the App Store Connect key id and issuer id (--api-key-id/--api-issuer-id or ASC_KEY_ID/ASC_ISSUER_ID)")
	case u.KeyPath == "":
		return fmt.Errorf("iOS upload needs the .p8 key (--api-key-path or ASC_KEY_P8_BASE64)")
	case u.Artifact == "":
		return fmt.Errorf("iOS upload needs --artifact <ipa>")
	}
	return nil
}

func (u IOS) Describe() []string {
	return []string{
		fmt.Sprintf("stage AuthKey_%s.p8 into a temporary private_keys dir (0600)", u.KeyID),
		fmt.Sprintf("xcrun altool --upload-app -f %s -t ios --apiKey %s --apiIssuer %s", u.Artifact, u.KeyID, u.IssuerID),
	}
}

func (u IOS) Run(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "anvil-asc")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	key, err := os.ReadFile(u.KeyPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "AuthKey_"+u.KeyID+".p8"), key, 0o600); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "xcrun", "altool", "--upload-app",
		"-f", u.Artifact, "-t", "ios", "--apiKey", u.KeyID, "--apiIssuer", u.IssuerID)
	cmd.Env = append(os.Environ(), "API_PRIVATE_KEYS_DIR="+dir)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
