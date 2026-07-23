package upload

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type NPM struct {
	Dir string
}

func (NPM) Platform() Platform { return PlatformNPM }

func (u NPM) Validate() error {
	if os.Getenv("NPM_TOKEN") != "" {
		return nil
	}
	for _, p := range []string{filepath.Join(u.dir(), ".npmrc"), filepath.Join(os.Getenv("HOME"), ".npmrc")} {
		if _, err := os.Stat(p); err == nil {
			return nil
		}
	}
	return fmt.Errorf("npm publish needs NPM_TOKEN or an .npmrc with an auth token")
}

func (u NPM) Describe() []string {
	return []string{fmt.Sprintf("npm publish (in %s)", u.dir())}
}

func (u NPM) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "npm", "publish")
	cmd.Dir = u.dir()
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func (u NPM) dir() string {
	if u.Dir == "" {
		return "."
	}
	return u.Dir
}
