package tool

import (
	"context"
	"path/filepath"
	"testing"
)

func TestGlob_Call(t *testing.T) {
	tempDir := t.TempDir()

	createTestFile(t, tempDir, "file1.txt", "content1")
	createTestFile(t, tempDir, "file2.txt", "content2")
	createTestFile(t, tempDir, "file3.doc", "content3")
	createTestFile(t, tempDir, "src/test1_test.go", "test content1")
	createTestFile(t, tempDir, "src/test2_test.go", "test content2")

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Match all txt files",
			input:   `{"pattern":"` + filepath.Join(tempDir, "*.txt") + `"}`,
			want:    filepath.Join(tempDir, "file1.txt") + "\n" + filepath.Join(tempDir, "file2.txt"),
			wantErr: false,
		},
		{
			name:    "Match test files in src directory",
			input:   `{"pattern":"` + filepath.Join(tempDir, "src", "*_test.go") + `"}`,
			want:    filepath.Join(tempDir, "src", "test1_test.go") + "\n" + filepath.Join(tempDir, "src", "test2_test.go"),
			wantErr: false,
		},
		{
			name:    "No matching files",
			input:   `{"pattern":"` + filepath.Join(tempDir, "*.xyz") + `"}`,
			want:    "",
			wantErr: true,
			errMsg:  "no files match the patttern: " + filepath.Join(tempDir, "*.xyz"),
		},
		{
			name:    "Invalid glob pattern",
			input:   `{"pattern":"` + filepath.Join(tempDir, "[a-z") + `"}`,
			want:    "",
			wantErr: true,
			errMsg:  "failed to execute pattern '" + filepath.Join(tempDir, "[a-z") + "': syntax error in pattern",
		},
		{
			name:    "Invalid JSON input",
			input:   `{"pattern":"*.txt"`,
			want:    "",
			wantErr: true,
			errMsg:  "failed to unmarshal input: unexpected end of JSON input",
		},
	}

	glob := &Glob{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := glob.Call(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Call() expected error, got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Call() error = %v, wantErr %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Call() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Call() = %v, want %v", got, tt.want)
			}
		})
	}
}
