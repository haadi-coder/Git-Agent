package tool

import (
	"encoding/json"
	"fmt"
	"os"
)

type ReadFile struct{}

var _ Tool = &ReadFile{}

func (t *ReadFile) Name() string {
	return "read_file"
}

func (t *ReadFile) Description() string {
	return "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names."
}

func (t *ReadFile) Params() map[string]any {
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

type ReadFileInput struct {
	Path string `json:"path"`
}

func (t *ReadFile) Call(input string) (string, error) {
	readFileInput := ReadFileInput{}

	err := json.Unmarshal([]byte(input), &readFileInput)
	if err != nil {
		return "", fmt.Errorf("failed to parse input json: %w", err)
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}
