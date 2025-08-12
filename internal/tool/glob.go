package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type Glob struct{}

var _ Tool = &Glob{}

func (t *Glob) Name() string {
	return "glob"
}

func (t *Glob) Description() string {
	return "Find files matching glob patterns. Useful for finding test files, configuration files, or files of specific types."
}

func (t *Glob) Params() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "A glob pattern to match files, e.g., '*.txt' for text files or 'src/*_test.*' for test files.",
			},
		},
	}
}

func (t *Glob) Call(ctx context.Context, input string) (string, error) {
	// args
	var args struct {
		Pattern string `json:"pattern"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal glob input: %w", err)
	}

	matches, err := filepath.Glob(args.Pattern)
	if err != nil {
		return "", fmt.Errorf("failed to execute glob pattern '%s': %w", args.Pattern, err)
	}

	if len(matches) == 0 {
		return "No files match the patttern", nil
	}

	return strings.Join(matches, "\n"), nil
}
