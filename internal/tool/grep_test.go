package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGrep_Call(t *testing.T) {
	tempDir := t.TempDir()

	createTestFile(t, tempDir, "file1.txt", "hello world\nfoo bar")
	createTestFile(t, tempDir, "file2.txt", "hello golang\nbaz qux")
	createTestFile(t, tempDir, "subdir/file3.txt", "hello there\nno match")

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid pattern and single file with match",
			input:   `{"pattern":"hello","path":"` + filepath.Join(tempDir, "file1.txt") + `"}`,
			want:    `["` + filepath.Join(tempDir, "file1.txt") + `:hello world"]`,
			wantErr: false,
		},
		{
			name:    "Valid pattern and directory with multiple matches",
			input:   `{"pattern":"hello","path":"` + tempDir + `"}`,
			want:    `["` + filepath.Join(tempDir, "file1.txt") + `:hello world","` + filepath.Join(tempDir, "file2.txt") + `:hello golang","` + filepath.Join(tempDir, "subdir", "file3.txt") + `:hello there"]`,
			wantErr: false,
		},
		{
			name:    "No matches found",
			input:   `{"pattern":"nonexistent","path":"` + filepath.Join(tempDir, "file1.txt") + `"}`,
			want:    "",
			wantErr: true,
			errMsg:  "nothing found based on nonexistent",
		},
		{
			name:    "Invalid regular expression",
			input:   `{"pattern":"[a-z","path":"` + filepath.Join(tempDir, "file1.txt") + `"}`,
			want:    "",
			wantErr: true,
			errMsg:  "failed to compile regular expression: error parsing regexp: missing closing ]: `[a-z`",
		},
		{
			name:    "Non-existent path",
			input:   `{"pattern":"hello","path":"` + filepath.Join(tempDir, "nonexistent.txt") + `"}`,
			want:    "",
			wantErr: true,
			errMsg:  "path " + filepath.Join(tempDir, "nonexistent.txt") + " doesnt exist",
		},
		{
			name:    "Invalid JSON input",
			input:   `{"pattern":"hello","path":`,
			want:    "",
			wantErr: true,
			errMsg:  "failed to unmarshal input: unexpected end of JSON input",
		},
	}

	grep := &Grep{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := grep.Call(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Call() expected error, result none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Call() error = %v, wantErr %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Call() unexpected error: %v", err)
				return
			}

			if result != tt.want {
				t.Errorf("Call() = %v, want %v", result, tt.want)
			}
		})
	}
}

func createTestFile(t *testing.T, tempDir, filePath, content string) {
	t.Helper()
	fullPath := filepath.Join(tempDir, filePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", fullPath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", fullPath, err)
	}
}
