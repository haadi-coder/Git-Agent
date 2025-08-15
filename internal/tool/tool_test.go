package tool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanPath(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		shouldFail bool
	}{
		{
			name:       "simple file",
			input:      "file.txt",
			shouldFail: false,
		},
		{
			name:       "file in subdirectory",
			input:      "dir/file.txt",
			shouldFail: false,
		},
		{
			name:       "file in parent directory",
			input:      "../file.txt",
			shouldFail: true,
		},
		{
			name:       "absolute path",
			input:      "/etc/passwd",
			shouldFail: true,
		},
		{
			name:       "multiple path traversal",
			input:      "../../../../etc/passwd",
			shouldFail: true,
		},
		{
			name:       "Windows path traversal",
			input:      "..\\..\\windows",
			shouldFail: true,
		},
		{
			name:       "empty path",
			input:      "",
			shouldFail: false,
		},
		{
			name:       "noraml relative path",
			input:      "./normal/path",
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleaned, err := cleanPath(tc.input)

			fmt.Print(cleaned)
			if tc.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
