package tool

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
		return "", fmt.Errorf("failed to unmarshal input: %w", err)
	}

	rgx, err := regexp.Compile(args.Pattern)
	if err != nil {
		return "", fmt.Errorf("failed to compile regular expression: %w", err)
	}

	path, err := cleanPath(args.Path)
	if err != nil {
		return "", fmt.Errorf("failed to clean path: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("path %s doesnt exist", args.Path)
		}

		return "", fmt.Errorf("failed to check path: %w", err)
	}

	var matches []string

	walkFunc := func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		file, _ := os.Open(filePath)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if rgx.MatchString(scanner.Text()) {
				matches = append(matches, fmt.Sprintf("%s:%s", filePath, scanner.Text()))
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

	if len(matches) == 0 {
		return "", fmt.Errorf("nothing found based on %s", args.Pattern)
	}

	output, err := json.Marshal(matches)
	if err != nil {
		return "", fmt.Errorf("failed to marshal output json: %w", err)
	}

	return string(output), nil
}
