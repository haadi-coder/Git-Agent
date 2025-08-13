package tool

import (
	"bytes"
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

	stdErr := bytes.Buffer{}
	stdOutput := bytes.Buffer{}
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOutput

	err := cmd.Run()

	resp := GitOutput{
		Output: string(stdOutput.String()),
	}

	if err != nil {
		if cmd.ProcessState.Exited() {
			resp.ExitCode = cmd.ProcessState.ExitCode()
		}
		resp.Error = stdErr.String()
	}

	output, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal output: %w", err)
	}

	return string(output), nil
}
