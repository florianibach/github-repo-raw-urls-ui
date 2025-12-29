package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	githubAPIBase  = "https://api.github.com"
	rawBase        = "https://raw.githubusercontent.com"
	defaultBranch  = "main"
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
		return "", "", errors.New("URL ist keine github.com-URL")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", errors.New("URL muss die Form https://github.com/owner/repo haben")
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
		return nil, fmt.Errorf("Repo oder Branch nicht gefunden (HTTP 404)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub API Fehler: %s", resp.Status)
	}

	var tr TreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return tr.Tree, nil
}

// ListRawURLs gibt alle Raw-Links für Dateien im Repo zurück.
func ListRawURLs(repoURL, branch string) ([]string, error) {
	if branch == "" {
		branch = defaultBranch
	}
	owner, repo, err := parseRepoURL(repoURL)
	if err != nil {
		return nil, err
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
