package tool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanpath(t *testing.T) {
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
		{"file in subdirectory", "dir/file.txt", false},
		{"file in parent directory", "../file.txt", true},
		{"absolute path", "/etc/passwd", true},
		{"multiple path traversal", "../../../../etc/passwd", true},
		{"Windows path traversal", "..\\..\\windows", true},
		{"empty path", "", false},
		{"noraml relative path", "./normal/path", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleaned, err := cleanpath(tc.input)

			fmt.Print(cleaned)
			if tc.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
