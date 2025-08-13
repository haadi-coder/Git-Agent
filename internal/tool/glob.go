package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	var args struct {
		Pattern string `json:"pattern"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %w", err)
	}

	matches, err := filepath.Glob(args.Pattern)
	if err != nil {
		return "", fmt.Errorf("failed to execute pattern '%s': %w", args.Pattern, err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no files match the patttern: %s", args.Pattern)
	}

	bytes, err := json.Marshal(matches)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bytes: %w", err)
	}

	return string(bytes), nil
}
