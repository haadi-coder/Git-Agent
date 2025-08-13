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

func TestRead(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		fileContent string
		expectError bool
	}{
		{
			name:        "success read",
			filePath:    filepath.Join(t.TempDir(), "hello.txt"),
			fileContent: "hi, bro",
			expectError: false,
		},
		{
			name:        "path traversal",
			filePath:    filepath.Join("..", t.TempDir(), "hello.txt"),
			fileContent: "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = os.WriteFile(tc.filePath, []byte(tc.fileContent), 0644)

			result, err := Tool.Call(&Read{}, context.Background(), fmt.Sprintf(`{"path":"%s"}`, tc.filePath))

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.fileContent, result)
		})
	}
}
