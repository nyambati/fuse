package init_test

import (
	"os"
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
		{
			name: "TestInitProject_ForceOverwrite",
			options: initer.InitOptions{
				Force: true,
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

			if tt.options.Force {
				os.WriteFile(filepath.Join(tempDir, ".fuse.yaml"), []byte("old content"), 0644)
			}

			err := initer.InitProject(tt.options)
			require.NoError(t, err)

			if tt.options.Force {
				newContent, _ := os.ReadFile(filepath.Join(tempDir, ".fuse.yaml"))
				assert.NotEqual(t, "old content", string(newContent))
			}

			for _, file := range tt.wants.files {
				assert.FileExists(t, filepath.Join(tempDir, file))
			}

			for _, dir := range tt.wants.dirs {
				assert.DirExists(t, filepath.Join(tempDir, dir))
			}
		})
	}
}

func TestInitTeam(t *testing.T) {
	setupFuseProject := func(t *testing.T) string {
		tempDir := t.TempDir()
		// Create a minimal fuse project structure
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "teams"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".fuse.yaml"), []byte("project: test"), 0644))
		return tempDir
	}

	tests := []struct {
		name    string
		options initer.InitOptions
		setup   func(t *testing.T, dir string)
		wantErr bool
		wants   struct {
			files []string
			dirs  []string
		}
	}{
		{
			name: "TestInitTeam_NewTeam",
			options: initer.InitOptions{
				Team: "myteam",
			},
			wantErr: false,
			wants: struct {
				files []string
				dirs  []string
			}{
				files: []string{
					"teams/myteam/channels.yaml",
					"teams/myteam/flows.yaml",
					"teams/myteam/silence_windows.yaml",
					"teams/myteam/alerts/example.yaml",
					"teams/myteam/templates/README.md",
				},
				dirs: []string{
					"teams/myteam/alerts",
					"teams/myteam/templates",
				},
			},
		},
		{
			name: "TestInitTeam_NoSample",
			options: initer.InitOptions{
				Team:     "myteam",
				NoSample: true,
			},
			wantErr: false,
			wants: struct {
				files []string
				dirs  []string
			}{
				files: []string{
					"teams/myteam/silence_windows.yaml",
					"teams/myteam/alerts/example.yaml",
					"teams/myteam/templates/README.md",
				},
				dirs: []string{
					"teams/myteam/alerts",
					"teams/myteam/templates",
				},
			},
		},
		{
			name: "TestInitTeam_ForceOverwrite",
			options: initer.InitOptions{
				Team:  "myteam",
				Force: true,
			},
			setup: func(t *testing.T, dir string) {
				teamDir := filepath.Join(dir, "teams", "myteam")
				os.MkdirAll(teamDir, 0755)
				os.WriteFile(filepath.Join(teamDir, "channels.yaml"), []byte("old content"), 0644)
			},
			wantErr: false,
			wants: struct {
				files []string
				dirs  []string
			}{
				files: []string{"teams/myteam/channels.yaml"},
			},
		},
		{
			name: "TestInitTeam_NoTeamName",
			options: initer.InitOptions{
				Team: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := setupFuseProject(t)
			tt.options.Path = tempDir

			if tt.setup != nil {
				tt.setup(t, tempDir)
			}

			err := initer.InitTeam(tt.options)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			for _, file := range tt.wants.files {
				assert.FileExists(t, filepath.Join(tempDir, file))
			}

			for _, dir := range tt.wants.dirs {
				assert.DirExists(t, filepath.Join(tempDir, dir))
			}

			if tt.options.Force && tt.setup != nil {
				content, _ := os.ReadFile(filepath.Join(tempDir, "teams/myteam/channels.yaml"))
				assert.NotEqual(t, "old content", string(content))
			}
		})
	}
}
