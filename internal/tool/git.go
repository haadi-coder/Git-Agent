package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/openai/openai-go"
)

var readOnlySubcommands = []string{"diff", "status", "log", "show", "branch", "rev-list", "ls-files", "rev-parse", "describe", "tag"}

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

type GitResponse struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error"`
}

func (t *Git) Call(ctx context.Context, input string) (string, error) {
	var params struct {
		Args []string `json:"args"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to unmarshal input json: %w", err)
	}

	subcommand := params.Args[0]

	if len(params.Args) < 2 {
		return "", fmt.Errorf("there should be at least 1 subcommand %s", subcommand)
	}

	if !slices.Contains(readOnlySubcommands, subcommand) {
		return "", fmt.Errorf("command is not permitted %s", strings.Join(params.Args, " "))
	}

	cmd := exec.CommandContext(ctx, "git", params.Args...)

	output, err := cmd.CombinedOutput()

	resp := GitResponse{
		Output: string(output),
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitError.ExitCode()
			resp.Error = err.Error()
		}
		resp.Error = err.Error()
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal output json: %w", err)
	}

	return string(marshalled), nil
}
