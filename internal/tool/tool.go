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

func cleanPath(path string) (string, error) {
	if path == "" {
		return ".", nil
	}

	cleaned := filepath.Clean(path)

	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("path traversal found: %s", path)
	}

	return cleaned, nil
}
