package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Grep struct{}

var _ Tool = &Grep{}

func (t *Grep) Name() string {
	return "grep"
}

func (t *Grep) Description() string {
	return "Searches text using patterns, including plain strings and regular expressions."
}

func (t *Grep) Params() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "A regular expression for searching the contents of files.",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file or directory to search in. If it is a directory, the search will be recursive.",
			},
		},
		"required": []string{"pattern", "path"},
	}
}

func (t *Grep) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		Pattern string `json:"pattern"`
		Path    string `json:"path"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal grep input: %w", err)
	}

	regexp, err := regexp.Compile(args.Pattern)
	if err != nil {
		return "", fmt.Errorf("failed to compile regular expression from provided pattern: %w", err)
	}

	path, err := cleanPath(args.Path)
	if err != nil {
		return "", fmt.Errorf("failed to validate path: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("path %s doesnt exist", args.Path)
		}
		return "", fmt.Errorf("failed to check path: %w", err)
	}

	var results []string

	walkFunc := func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if regexp.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d:%s", filePath, i+1, line))
			}
		}

		return nil
	}

	if info.IsDir() {
		err = filepath.Walk(args.Path, walkFunc)
	} else {
		err = walkFunc(args.Path, info, nil)
	}

	if err != nil {
		return "", fmt.Errorf("failed to walk through files: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("nothing found based on %s", args.Pattern)
	}

	return strings.Join(results, "\n"), nil
}
