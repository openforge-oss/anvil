package upload

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileCred resolves a file-based credential: an explicit path, else the path in
// pathEnv, else the base64 contents in base64Env decoded to a 0600 temp file
// (the returned cleanup removes it). It refuses a path inside repoRoot so
// secrets are never read from the working tree.
func FileCred(repoRoot, explicitPath, pathEnv, base64Env, tmpName string) (path string, cleanup func(), err error) {
	cleanup = func() {}
	switch {
	case explicitPath != "":
		path = explicitPath
	case os.Getenv(pathEnv) != "":
		path = os.Getenv(pathEnv)
	case os.Getenv(base64Env) != "":
		raw, derr := base64.StdEncoding.DecodeString(strings.TrimSpace(os.Getenv(base64Env)))
		if derr != nil {
			return "", cleanup, fmt.Errorf("%s: %w", base64Env, derr)
		}
		dir, derr := os.MkdirTemp("", "anvil-cred")
		if derr != nil {
			return "", cleanup, derr
		}
		path = filepath.Join(dir, tmpName)
		if werr := os.WriteFile(path, raw, 0o600); werr != nil {
			os.RemoveAll(dir)
			return "", cleanup, werr
		}
		return path, func() { os.RemoveAll(dir) }, nil
	default:
		return "", cleanup, fmt.Errorf("missing credential: set %s, %s, or provide a path", pathEnv, base64Env)
	}

	if inside(repoRoot, path) {
		return "", cleanup, fmt.Errorf("refusing to read a credential inside the repo: %s (keep it outside the working tree)", path)
	}
	if _, e := os.Stat(path); e != nil {
		return "", cleanup, fmt.Errorf("credential not found: %s", path)
	}
	return path, cleanup, nil
}

func inside(root, path string) bool {
	ra, err1 := filepath.Abs(root)
	pa, err2 := filepath.Abs(path)
	if err1 != nil || err2 != nil {
		return false
	}
	rel, err := filepath.Rel(ra, pa)
	return err == nil && !strings.HasPrefix(rel, "..")
}
