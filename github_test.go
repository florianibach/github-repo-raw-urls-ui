package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "simple",
			input:     "https://github.com/user/repo",
			wantOwner: "user",
			wantRepo:  "repo",
		},
		{
			name:      "trailing slash",
			input:     "https://github.com/user/repo/",
			wantOwner: "user",
			wantRepo:  "repo",
		},
		{
			name:      "with tree path",
			input:     "https://github.com/user/repo/tree/main",
			wantOwner: "user",
			wantRepo:  "repo",
		},
		{
			name:    "invalid host",
			input:   "https://gitlab.com/user/repo",
			wantErr: true,
		},
		{
			name:    "too short path",
			input:   "https://github.com/user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRepoURL(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOwner, owner)
			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}

func TestBuildRawURL(t *testing.T) {
	owner := "user"
	repo := "repo"
	branch := "main"
	filePath := "dir/file.go"

	got := buildRawURL(owner, repo, branch, filePath)
	want := "https://raw.githubusercontent.com/user/repo/main/dir/file.go"
	assert.Equal(t, want, got)
}
