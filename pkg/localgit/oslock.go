package localgit

import (
	"fmt"
	"os"
	"path/filepath"
)

const mutationLockName = "gitboard.mutation.lock"

func openMutationLockFile(commonDir string) (*os.File, error) {
	commonDir = filepath.Clean(commonDir)
	if commonDir == "" || commonDir == "." {
		return nil, ErrCommonDirRequired
	}
	if err := os.MkdirAll(commonDir, 0o755); err != nil {
		return nil, fmt.Errorf("mutation lock dir: %w", err)
	}
	path := filepath.Join(commonDir, mutationLockName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open mutation lock: %w", err)
	}
	return f, nil
}
