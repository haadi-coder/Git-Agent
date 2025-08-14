package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWithTempDir(t *testing.T) {
	originalWD, _ := os.Getwd()

	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalWD)
	}()

	tests := []struct {
		name        string
		filePath    string
		fileContent string
		expectError bool
	}{
		{
			name:        "success read in temp dir",
			filePath:    "hello.txt",
			fileContent: "hi, bro",
			expectError: false,
		},
		{
			name:        "success read in subdirectory",
			filePath:    "subdir/file.txt",
			fileContent: "content in subdir",
			expectError: false,
		},
		{
			name:        "path traversal should fail",
			filePath:    "../../../etc/passwd",
			fileContent: "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fileContent != "" {
				dir := filepath.Dir(tc.filePath)
				if dir != "." {
					err := os.MkdirAll(dir, 0755)
					require.NoError(t, err)
				}

				err := os.WriteFile(tc.filePath, []byte(tc.fileContent), 0644)
				require.NoError(t, err)
			}

			result, err := Tool.Call(&Read{}, context.Background(), fmt.Sprintf(`{"path":"%s"}`, tc.filePath))

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.fileContent, result)
			}
		})
	}
}
