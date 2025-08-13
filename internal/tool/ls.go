package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
)

type LS struct{}

var _ Tool = &LS{}

func (t *LS) Name() string {
	return "list_files"
}

func (t *LS) Description() string {
	return "tool to list entries of certain path. If there isnt any path provided, list entries of current directory."
}

func (t *LS) Params() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The relative path of directory to list files. Default value is current directory if any path dont provided",
			},
		},
		"required": []string{"path"},
	}
}

func (t *LS) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %w", err)
	}

	dir, err := cleanPath(args.Path)
	if err != nil {
		return "", fmt.Errorf("failed to clean path: %w", err)
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk through files: %w", err)
	}

	output, err := json.Marshal(files)
	if err != nil {
		return "", fmt.Errorf("failed to marshal output: %w", err)
	}

	return string(output), nil
}
