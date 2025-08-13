package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type GitOutput struct {
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Error    string `json:"error"`
}

func (t *Git) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		Args []string `json:"args"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %w", err)
	}

	if len(args.Args) == 0 {
		return "", fmt.Errorf("there should be at least 1 subcommand")
	}

	subcommand := args.Args[0]
	if !slices.Contains(readOnlySubcommands, subcommand) {
		return "", fmt.Errorf("only readOnly commands available to use: %s", strings.Join(readOnlySubcommands, ","))
	}

	cmd := exec.CommandContext(ctx, "git", args.Args...)

	stderr := bytes.Buffer{}
	stdout := bytes.Buffer{}
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return "", fmt.Errorf("failed to exec git cmd: %w", err)
		}
	}

	output := GitOutput{
		ExitCode: cmd.ProcessState.ExitCode(),
		Output:   stdout.String(),
		Error:    stderr.String(),
	}

	bytes, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bytes: %w", err)
	}

	return string(bytes), nil
}
