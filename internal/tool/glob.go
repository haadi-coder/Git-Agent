package tool

import (
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

type GlobInput struct {
	Pattern string `json:"pattern"`
}

func (t *Glob) Call(input string) (string, error) {
	var globInput GlobInput

	if err := json.Unmarshal([]byte(input), &globInput); err != nil {
		return "", fmt.Errorf("failed to parse glob input: %w", err)
	}

	matches, err := filepath.Glob(globInput.Pattern)
	if err != nil {
		return "", fmt.Errorf("failed to execute glob pattern '%s': %w", globInput.Pattern, err)
	}

	if len(matches) == 0 {
		return "No files match the patttern", nil
	}

	return strings.Join(matches, "\n"), nil
}
