package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type GithubRepo struct {
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	HtmlUrl     string   `json:"html_url"`
	Topics      []string `json:"topics"`
	Private     bool     `json:"private"`
}

// FetchAllGithubRepos mengambil semua repo publik dari username di .env
func FetchAllGithubRepos() ([]GithubRepo, error) {

	username := os.Getenv("GITHUB_USERNAME")
	if username == "" {
		return nil, fmt.Errorf("GITHUB_USERNAME is not set")
	}

	// Accept header khusus untuk dapat topics dari GitHub API
	apiUrl := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&sort=updated", username)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "portfolio-app")
	req.Header.Set("Accept", "application/vnd.github.mercy-preview+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch github repos: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("github user not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var repos []GithubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode github response: %v", err)
	}

	// Filter hanya repo publik
	var publicRepos []GithubRepo
	for _, repo := range repos {
		if !repo.Private {
			publicRepos = append(publicRepos, repo)
		}
	}

	return publicRepos, nil
}
