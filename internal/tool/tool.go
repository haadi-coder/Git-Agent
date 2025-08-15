package tool

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type Tool interface {
	Name() string
	Description() string
	Params() map[string]any
	Call(ctx context.Context, input string) (string, error)
}

func cleanPath(inputpath string) (string, error) {
	if inputpath == "" {
		return ".", nil
	}

	cleaned := filepath.Clean(inputpath)
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute paths is not allowed: %s", inputpath)
	}

	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("path traversal found: %s", inputpath)
	}

	return cleaned, nil
}
