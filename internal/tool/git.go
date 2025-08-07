package tool

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"

	"github.com/openai/openai-go"
)

type Git struct{}

var _ Tool = &Git{}

func (t *Git) Name() string {
	return "git_command"
}

func (t *Git) Description() string {
	return "Execute a safe Git command to retrieve repository information (e.g., diff, status, log)."
}

func (t *Git) Params() map[string]any {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"args": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"description": "Git command arguments (e.g., ['status', '--porcelain'] or ['log', '--oneline', '-5'])",
			},
		},
		"required": []string{"args"},
	}
}

type GitInput struct {
	Args []string `json:"args"`
}

func (t *Git) Call(input string) (string, error) {
	var gitInput GitInput
	if err := json.Unmarshal([]byte(input), &gitInput); err != nil {
		return "", err
	}

	allowedCommands := []string{"diff", "status", "log", "show", "branch", "rev-list", "ls-files", "rev-parse", "describe", "tag"}

	command := gitInput.Args[0]

	if len(gitInput.Args) == 0 || !slices.Contains(allowedCommands, command) {
		return "", fmt.Errorf("invalid Git command: %s", command)
	}

	args := append([]string{command}, gitInput.Args[1:]...)
	cmd := exec.Command("git", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}
