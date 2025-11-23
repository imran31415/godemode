package githubtools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var (
	githubToken string
	httpClient  = &http.Client{}
)

// InitGitHub initializes the GitHub client with token
func InitGitHub(token string) {
	githubToken = token
}

// GetToken returns the current GitHub token
func GetToken() string {
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	}
	return githubToken
}

func makeRequest(method, endpoint string, body interface{}) (map[string]interface{}, error) {
	token := GetToken()
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN not set")
	}

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, "https://api.github.com"+endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Try array response
		var arrayResult []interface{}
		if err2 := json.Unmarshal(respBody, &arrayResult); err2 == nil {
			return map[string]interface{}{
				"items": arrayResult,
				"count": len(arrayResult),
			}, nil
		}
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

func getString(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(args map[string]interface{}, key string) int {
	if v, ok := args[key]; ok {
		switch i := v.(type) {
		case int:
			return i
		case float64:
			return int(i)
		case string:
			n, _ := strconv.Atoi(i)
			return n
		}
	}
	return 0
}

func searchRepositories(args map[string]interface{}) (interface{}, error) {
	query := getString(args, "query")
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	params := url.Values{}
	params.Set("q", query)
	if perPage := getInt(args, "per_page"); perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}
	if page := getInt(args, "page"); page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	return makeRequest("GET", "/search/repositories?"+params.Encode(), nil)
}

func getRepository(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	return makeRequest("GET", fmt.Sprintf("/repos/%s/%s", owner, repo), nil)
}

func listIssues(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	params := url.Values{}
	if state := getString(args, "state"); state != "" {
		params.Set("state", state)
	}
	if labels := getString(args, "labels"); labels != "" {
		params.Set("labels", labels)
	}
	if perPage := getInt(args, "per_page"); perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	return makeRequest("GET", endpoint, nil)
}

func getIssue(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	issueNumber := getInt(args, "issue_number")
	if owner == "" || repo == "" || issueNumber == 0 {
		return nil, fmt.Errorf("owner, repo, and issue_number are required")
	}

	return makeRequest("GET", fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, issueNumber), nil)
}

// NOTE: Write operations (createIssue, updateIssue, addIssueComment)
// have been intentionally removed for safety. This benchmark is READ-ONLY.

func listIssueComments(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	issueNumber := getInt(args, "issue_number")
	if owner == "" || repo == "" || issueNumber == 0 {
		return nil, fmt.Errorf("owner, repo, and issue_number are required")
	}

	return makeRequest("GET", fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, issueNumber), nil)
}

func searchIssues(args map[string]interface{}) (interface{}, error) {
	query := getString(args, "query")
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	params := url.Values{}
	params.Set("q", query)
	if perPage := getInt(args, "per_page"); perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}

	return makeRequest("GET", "/search/issues?"+params.Encode(), nil)
}

func listPullRequests(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	params := url.Values{}
	if state := getString(args, "state"); state != "" {
		params.Set("state", state)
	}
	if perPage := getInt(args, "per_page"); perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	return makeRequest("GET", endpoint, nil)
}

func getPullRequest(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	pullNumber := getInt(args, "pull_number")
	if owner == "" || repo == "" || pullNumber == 0 {
		return nil, fmt.Errorf("owner, repo, and pull_number are required")
	}

	return makeRequest("GET", fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, pullNumber), nil)
}

func listCommits(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	params := url.Values{}
	if sha := getString(args, "sha"); sha != "" {
		params.Set("sha", sha)
	}
	if perPage := getInt(args, "per_page"); perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/commits", owner, repo)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	return makeRequest("GET", endpoint, nil)
}

func getFileContents(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	path := getString(args, "path")
	if owner == "" || repo == "" || path == "" {
		return nil, fmt.Errorf("owner, repo, and path are required")
	}

	params := url.Values{}
	if ref := getString(args, "ref"); ref != "" {
		params.Set("ref", ref)
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	return makeRequest("GET", endpoint, nil)
}

func searchCode(args map[string]interface{}) (interface{}, error) {
	query := getString(args, "query")
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	params := url.Values{}
	params.Set("q", query)
	if perPage := getInt(args, "per_page"); perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}

	return makeRequest("GET", "/search/code?"+params.Encode(), nil)
}

func listLabels(args map[string]interface{}) (interface{}, error) {
	owner := getString(args, "owner")
	repo := getString(args, "repo")
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	return makeRequest("GET", fmt.Sprintf("/repos/%s/%s/labels", owner, repo), nil)
}
