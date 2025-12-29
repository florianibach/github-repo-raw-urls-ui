package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace",
			input:   "   ",
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

func TestListRawURLsWithOwnerRepo_FiltersBlobsOnly(t *testing.T) {
	// Arrange: fake GitHub API server
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/git/trees/main", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "github-repo-raw-urls-ui", r.Header.Get("User-Agent"))
		assert.Equal(t, "1", r.URL.Query().Get("recursive"))

		resp := TreeResponse{
			Sha: "abc",
			Tree: []TreeObject{
				{Path: "README.md", Type: "blob"},
				{Path: "dir", Type: "tree"},
				{Path: "dir/main.go", Type: "blob"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Patch bases for test
	oldAPI := githubAPIBase
	oldRaw := rawBase
	githubAPIBase = srv.URL
	rawBase = "https://raw.test"
	defer func() {
		githubAPIBase = oldAPI
		rawBase = oldRaw
	}()

	// Act
	urls, err := ListRawURLsWithOwnerRepo("o", "r", "main")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"https://raw.test/o/r/main/README.md",
		"https://raw.test/o/r/main/dir/main.go",
	}, urls)
}

func TestFetchTree_404(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/git/trees/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldAPI := githubAPIBase
	githubAPIBase = srv.URL
	defer func() { githubAPIBase = oldAPI }()

	_, err := fetchTree("o", "r", "missing")
	assert.Error(t, err)
	// message depends on your english translation; adjust if you change it
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestGetBranches_MarksDefaultBranch(t *testing.T) {
	mux := http.NewServeMux()

	// repo info endpoint
	mux.HandleFunc("/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "github-repo-raw-urls-ui", r.Header.Get("User-Agent"))
		_ = json.NewEncoder(w).Encode(repoInfo{DefaultBranch: "main"})
	})

	// branches endpoint
	mux.HandleFunc("/repos/o/r/branches", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "github-repo-raw-urls-ui", r.Header.Get("User-Agent"))
		_ = json.NewEncoder(w).Encode([]map[string]string{
			{"name": "dev"},
			{"name": "main"},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldAPI := githubAPIBase
	githubAPIBase = srv.URL
	defer func() { githubAPIBase = oldAPI }()

	branches, def, err := getBranches("o", "r")
	assert.NoError(t, err)
	assert.Equal(t, "main", def)
	assert.Len(t, branches, 2)

	// Find main and ensure Default=true
	var mainB *Branch
	for i := range branches {
		if branches[i].Name == "main" {
			mainB = &branches[i]
			break
		}
	}
	if assert.NotNil(t, mainB) {
		assert.True(t, mainB.Default)
	}
}
