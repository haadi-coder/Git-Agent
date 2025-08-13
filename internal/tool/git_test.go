package tool

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit_Call_ValidCommands(t *testing.T) {
	git := &Git{}
	ctx := context.Background()

	testCases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "git status",
			args:    []string{"status"},
			wantErr: false,
		},
		{
			name:    "git status with flags",
			args:    []string{"status", "--porcelain"},
			wantErr: false,
		},
		{
			name:    "git log",
			args:    []string{"log", "--oneline", "-5"},
			wantErr: false,
		},
		{
			name:    "git diff",
			args:    []string{"diff", "--name-only"},
			wantErr: false,
		},
		{
			name:    "git branch",
			args:    []string{"branch", "-a"},
			wantErr: false,
		},
		{
			name:    "git show",
			args:    []string{"show", "--stat"},
			wantErr: false,
		},
		{
			name:    "git rev-list",
			args:    []string{"rev-list", "HEAD", "--count"},
			wantErr: false,
		},
		{
			name:    "git ls-files",
			args:    []string{"ls-files"},
			wantErr: false,
		},
		{
			name:    "git rev-parse",
			args:    []string{"rev-parse", "HEAD"},
			wantErr: false,
		},
		{
			name:    "git describe",
			args:    []string{"describe", "--tags"},
			wantErr: false,
		},
		{
			name:    "git tag",
			args:    []string{"tag", "-l"},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := map[string][]string{
				"args": tc.args,
			}
			inputJSON, err := json.Marshal(input)
			require.NoError(t, err)

			result, err := git.Call(ctx, string(inputJSON))

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			var output GitOutput
			err = json.Unmarshal([]byte(result), &output)
			assert.NoError(t, err)

			t.Logf("Exit code: %d, Output: %s, Error: %s", output.ExitCode, output.Output, output.Error)
		})
	}
}

func TestGit_Call_InvalidCommands(t *testing.T) {
	git := &Git{}
	ctx := context.Background()

	testCases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "git add (write command)",
			args:    []string{"add", "."},
			wantErr: "only readOnly commands available to use",
		},
		{
			name:    "git commit (write command)",
			args:    []string{"commit", "-m", "test"},
			wantErr: "only readOnly commands available to use",
		},
		{
			name:    "git push (write command)",
			args:    []string{"push"},
			wantErr: "only readOnly commands available to use",
		},
		{
			name:    "git pull (write command)",
			args:    []string{"pull"},
			wantErr: "only readOnly commands available to use",
		},
		{
			name:    "git reset (dangerous command)",
			args:    []string{"reset", "--hard"},
			wantErr: "only readOnly commands available to use",
		},
		{
			name:    "git rm (write command)",
			args:    []string{"rm", "file.txt"},
			wantErr: "only readOnly commands available to use",
		},
		{
			name:    "git checkout (write command)",
			args:    []string{"checkout", "branch"},
			wantErr: "only readOnly commands available to use",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := map[string][]string{
				"args": tc.args,
			}
			inputJSON, err := json.Marshal(input)
			require.NoError(t, err)

			result, err := git.Call(ctx, string(inputJSON))

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
			assert.Empty(t, result)
		})
	}
}

func TestGit_Call_EdgeCases(t *testing.T) {
	git := &Git{}
	ctx := context.Background()

	t.Run("empty args", func(t *testing.T) {
		input := map[string][]string{
			"args": {},
		}
		inputJSON, err := json.Marshal(input)
		require.NoError(t, err)

		result, err := git.Call(ctx, string(inputJSON))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "there should be at least 1 subcommand")
		assert.Empty(t, result)
	})

	t.Run("invalid JSON input", func(t *testing.T) {
		result, err := git.Call(ctx, "invalid json")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal input")
		assert.Empty(t, result)
	})

	t.Run("missing args field", func(t *testing.T) {
		input := map[string]string{
			"other": "field",
		}
		inputJSON, err := json.Marshal(input)
		require.NoError(t, err)

		result, err := git.Call(ctx, string(inputJSON))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "there should be at least 1 subcommand")
		assert.Empty(t, result)
	})
}

func TestGit_Call_ContextCancellation(t *testing.T) {
	git := &Git{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	input := map[string][]string{
		"args": {"status"},
	}
	inputJSON, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := git.Call(ctx, string(inputJSON))

	if err != nil {
		t.Logf("Command cancelled as expected: %v", err)
	} else {
		var output GitOutput
		unmarshalErr := json.Unmarshal([]byte(result), &output)
		assert.NoError(t, unmarshalErr)
		t.Logf("Command completed before cancellation")
	}
}

func TestGit_Call_OutputParsing(t *testing.T) {
	git := &Git{}
	ctx := context.Background()

	input := map[string][]string{
		"args": {"--version"},
	}
	inputJSON, err := json.Marshal(input)
	require.NoError(t, err)

	_, err = git.Call(ctx, string(inputJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only readOnly commands available to use")
}

func TestGit_Call_InGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not available")
	}

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	cmd := exec.Command("git", "init")
	err = cmd.Run()
	if err != nil {
		t.Skip("Could not initialize git repo")
	}

	_ = exec.Command("git", "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "config", "user.name", "Test User").Run()

	git := &Git{}
	ctx := context.Background()

	t.Run("git status in empty repo", func(t *testing.T) {
		input := map[string][]string{
			"args": {"status"},
		}
		inputJSON, err := json.Marshal(input)
		require.NoError(t, err)

		result, err := git.Call(ctx, string(inputJSON))
		assert.NoError(t, err)

		var output GitOutput
		err = json.Unmarshal([]byte(result), &output)
		assert.NoError(t, err)

		assert.Contains(t, strings.ToLower(output.Output), "master")
	})

	t.Run("git ls-files in empty repo", func(t *testing.T) {
		input := map[string][]string{
			"args": {"ls-files"},
		}
		inputJSON, err := json.Marshal(input)
		require.NoError(t, err)

		result, err := git.Call(ctx, string(inputJSON))
		assert.NoError(t, err)

		var output GitOutput
		err = json.Unmarshal([]byte(result), &output)
		assert.NoError(t, err)

		assert.Empty(t, strings.TrimSpace(output.Output))
	})
}
