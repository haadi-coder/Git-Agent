package tool

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
)

type ListFilesTool struct{}

var _ Tool = &ListFilesTool{}

func (t *ListFilesTool) Name() string {
	return "list_files"
}

func (t *ListFilesTool) Description() string {
	return "tool to list entries of certain path. If there isnt any path provided, list entries of current directory."
}

func (t *ListFilesTool) Params() map[string]any {
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

type ListFilesInput struct {
	Path string `json:"path"`
}

func (t *ListFilesTool) Call(input string) (string, error) {
	listFilesInput := ListFilesInput{}
	if err := json.Unmarshal([]byte(input), &listFilesInput); err != nil {
		return "", fmt.Errorf("failed to parse input json: %w", err)
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
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
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
