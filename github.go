package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	githubAPIBase = "https://api.github.com"
	rawBase       = "https://raw.githubusercontent.com"
)

const (
	defaultBranch  = "master"
	httpTimeoutSec = 10
)

type TreeResponse struct {
	Sha  string       `json:"sha"`
	Tree []TreeObject `json:"tree"`
}

type TreeObject struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Sha  string `json:"sha"`
	URL  string `json:"url"`
	Size int    `json:"size"`
}

type Branch struct {
	Name    string `json:"name"`
	Default bool   `json:"-"`
}

type repoInfo struct {
	DefaultBranch string `json:"default_branch"`
}

// parseRepoURL extrahiert owner und repo aus z.B.
// https://github.com/owner/repo
// https://github.com/owner/repo/
// https://github.com/owner/repo/tree/main
func parseRepoURL(repoURL string) (owner, repo string, err error) {
	u, err := url.Parse(strings.TrimSpace(repoURL))
	if err != nil {
		return "", "", err
	}
	if u.Host != "github.com" {
		return "", "", errors.New("URL must be a github.com URL")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", errors.New("URL must have the form https://github.com/owner/repo")
	}
	return parts[0], parts[1], nil
}

// buildRawURL erzeugt einen Raw-Link für eine Datei.
func buildRawURL(owner, repo, branch, filePath string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", rawBase, owner, repo, branch, filePath)
}

// fetchTree nutzt die GitHub Git Trees API mit ?recursive=1.
func fetchTree(owner, repo, branch string) ([]TreeObject, error) {
	client := &http.Client{Timeout: httpTimeoutSec * time.Second}
	apiURL := fmt.Sprintf("%s/repos/%s/%s/git/trees/%s?recursive=1", githubAPIBase, owner, repo, branch)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Optional: User-Agent setzen, manche Setups verlangen das
	req.Header.Set("User-Agent", "github-repo-raw-urls-ui")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Repository or branch not found (HTTP 404)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var tr TreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return tr.Tree, nil
}

func ListRawURLsWithOwnerRepo(owner, repo, branch string) ([]string, error) {
	if branch == "" {
		branch = defaultBranch
	}
	tree, err := fetchTree(owner, repo, branch)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, obj := range tree {
		if obj.Type == "blob" && !strings.HasSuffix(obj.Path, "/") {
			urls = append(urls, buildRawURL(owner, repo, branch, obj.Path))
		}
	}
	return urls, nil
}

func ListRawURLs(repoURL, branch string) ([]string, error) {
	owner, repo, err := parseRepoURL(repoURL)
	if err != nil {
		return nil, err
	}
	return ListRawURLsWithOwnerRepo(owner, repo, branch)
}

func getRepoInfo(owner, repo string) (*repoInfo, error) {
	client := &http.Client{Timeout: httpTimeoutSec * time.Second}
	apiURL := fmt.Sprintf("%s/repos/%s/%s", githubAPIBase, owner, repo)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "github-repo-raw-urls-ui")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Repository not found (HTTP 404)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (repo info): %s – %s", resp.Status, string(body))
	}

	var info repoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// getBranches holt alle Branches und markiert den Default-Branch.
func getBranches(owner, repo string) ([]Branch, string, error) {
	client := &http.Client{Timeout: httpTimeoutSec * time.Second}

	// Erst Default-Branch holen
	info, err := getRepoInfo(owner, repo)
	if err != nil {
		return nil, "", err
	}
	defaultBranch := info.DefaultBranch

	apiURL := fmt.Sprintf("%s/repos/%s/%s/branches?per_page=100", githubAPIBase, owner, repo)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "github-repo-raw-urls-ui")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, "", fmt.Errorf("Branches not found (HTTP 404)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("GitHub API error (branches): %s – %s", resp.Status, string(body))
	}

	var tmp []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tmp); err != nil {
		return nil, "", err
	}

	branches := make([]Branch, 0, len(tmp))
	for _, b := range tmp {
		branches = append(branches, Branch{
			Name:    b.Name,
			Default: b.Name == defaultBranch,
		})
	}
	return branches, defaultBranch, nil
}
