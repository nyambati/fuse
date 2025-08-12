package init_test

import (
	"path/filepath"
	"testing"

	initer "github.com/nyambati/fuse/internal/init"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitProject(t *testing.T) {
	tests := []struct {
		name    string
		options initer.InitOptions
		wants   struct {
			files []string
			dirs  []string
		}
	}{
		{
			name:    "TestInitProject_EmptyDir",
			options: initer.InitOptions{},
			wants: struct {
				files []string
				dirs  []string
			}{
				files: []string{".fuse.yaml", "global/global.yaml", "global/silence_windows.yaml", "teams/README.md"},
				dirs:  []string{"dist"},
			},
		},
		{
			name: "TestInitProject_NoSample",
			options: initer.InitOptions{
				NoSample: true,
			},
			wants: struct {
				files []string
				dirs  []string
			}{
				files: []string{".fuse.yaml"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.options.Path = tempDir
			err := initer.InitProject(tt.options)
			require.NoError(t, err)

			for _, file := range tt.wants.files {
				assert.FileExists(t, filepath.Join(tempDir, file))
			}
			for _, dir := range tt.wants.dirs {
				assert.DirExists(t, filepath.Join(tempDir, dir))
			}
		})
	}
}
