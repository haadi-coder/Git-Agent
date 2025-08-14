package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLS(t *testing.T) {
	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	require.NoError(t, err)

	createTestFileStructure(t, tempDir)

	tests := []struct {
		name          string
		inputpath     string
		expectedFiles []string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid path with files and directories",
			inputpath:     ".",
			expectedFiles: []string{"file1.txt", "file2.txt", "subdir/", "subdir/nested.txt"},
			expectError:   false,
		},
		{
			name:          "Empty path (current directory)",
			inputpath:     "",
			expectedFiles: []string{"file1.txt", "file2.txt", "subdir/", "subdir/nested.txt"},
			expectError:   false,
		},
		{
			name:          "Non-existent path",
			inputpath:     "./non/existent/path",
			expectError:   true,
			errorContains: "failed to walk through files",
		},
		{
			name:          "Path traversal",
			inputpath:     "../" + tempDir,
			expectError:   true,
			errorContains: "path traversal found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := Tool.Call(&LS{}, context.Background(), fmt.Sprintf(`{"path":"%s"}`, tt.inputpath))

			if tt.expectError {
				assert.Error(t, err, "Expected an error")
				assert.Contains(t, err.Error(), tt.errorContains, "Error message should contain expected substring")
				return
			}

			assert.NoError(t, err, "Expected no error")
			var files []string
			err = json.Unmarshal([]byte(result), &files)
			assert.NoError(t, err, "Failed to unmarshal result")
			assert.ElementsMatch(t, tt.expectedFiles, files, "Listed files should match expected")
		})
	}
}

func createTestFileStructure(t *testing.T, tempDir string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
	assert.NoError(t, err, "Failed to create file1.txt")

	err = os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content2"), 0644)
	assert.NoError(t, err, "Failed to create file2.txt")

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	assert.NoError(t, err, "Failed to create subdir")

	err = os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644)
	assert.NoError(t, err, "Failed to create nested.txt")

	originalPath, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalPath)
	}()
}
