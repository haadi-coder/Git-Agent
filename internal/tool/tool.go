package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tool interface {
	Name() string
	Description() string
	Params() map[string]any
	Call(ctx context.Context, input string) (string, error)
}

func cleanpath(inputpath string) (string, error) {
	if inputpath == "" {
		return ".", nil
	}

	cleaned := filepath.Clean(inputpath)
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute paths is not allowed: %s", inputpath)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	fullpath := filepath.Join(workDir, inputpath)

	if !strings.HasPrefix(fullpath, workDir) || strings.Contains(inputpath, "..\\") {
		return "", fmt.Errorf("path traversal found: %s", inputpath)
	}

	relPath, err := filepath.Rel(workDir, fullpath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	return relPath, nil
}
