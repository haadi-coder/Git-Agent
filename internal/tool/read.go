package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type Read struct{}

var _ Tool = &Read{}

func (t *Read) Name() string {
	return "read_file"
}

func (t *Read) Description() string {
	return "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names."
}

func (t *Read) Params() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The relative path of a file in the working directory",
			},
		},
		"required": []string{"path"},
	}
}

func (t *Read) Call(ctx context.Context, input string) (string, error) {
	var params struct {
		Path string `json:"path"`
	}

	err := json.Unmarshal([]byte(input), &params)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal input json: %w", err)
	}

	path, err := validatePath(params.Path)
	if err != nil {
		return "", fmt.Errorf("failed to validate path: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}
