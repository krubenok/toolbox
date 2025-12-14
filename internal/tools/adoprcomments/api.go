package adoprcomments

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyrubeno/toolbox/internal/auth"
)

// Client handles Azure DevOps API requests.
type Client struct {
	auth       *auth.Auth
	httpClient *http.Client
	debug      bool
	debugLog   func(string)
}

// NewClient creates a new ADO API client.
func NewClient(auth *auth.Auth, debug bool, debugLog func(string)) *Client {
	return &Client{
		auth:       auth,
		httpClient: &http.Client{},
		debug:      debug,
		debugLog:   debugLog,
	}
}

// fetchJSON performs a GET request and decodes the JSON response.
func (c *Client) fetchJSON(apiURL string, result any) error {
	if c.debug && c.debugLog != nil {
		c.debugLog(fmt.Sprintf("Fetching: %s", apiURL))
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.auth.AuthorizationHeader())
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed (%s): %s", resp.Status, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// threadsURL builds the PR threads API URL.
func threadsURL(pr *ParsedPR) string {
	return fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/git/repositories/%s/pullRequests/%s/threads?api-version=7.1-preview.1",
		url.PathEscape(pr.Organization),
		url.PathEscape(pr.Project),
		url.PathEscape(pr.Repository),
		url.PathEscape(pr.PRID),
	)
}

// prURL builds the PR details API URL.
func prURL(pr *ParsedPR) string {
	return fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/git/pullRequests/%s?api-version=7.1-preview.1",
		url.PathEscape(pr.Organization),
		url.PathEscape(pr.Project),
		url.PathEscape(pr.PRID),
	)
}

// threadsURLByRepoID builds the PR threads API URL using repository ID.
func threadsURLByRepoID(pr *ParsedPR, repoID string) string {
	return fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/git/repositories/%s/pullRequests/%s/threads?api-version=7.1-preview.1",
		url.PathEscape(pr.Organization),
		url.PathEscape(pr.Project),
		url.PathEscape(repoID),
		url.PathEscape(pr.PRID),
	)
}

// ThreadsResponse represents the API response for PR threads.
type ThreadsResponse struct {
	Value []Thread `json:"value"`
}

// Thread represents a comment thread on a PR.
type Thread struct {
	ID            int            `json:"id"`
	Status        string         `json:"status"`
	ThreadContext *ThreadContext `json:"threadContext"`
	Properties    *ThreadProps   `json:"properties"`
	Comments      []Comment      `json:"comments"`
}

// ThreadContext contains file location information.
type ThreadContext struct {
	FilePath       string        `json:"filePath"`
	RightFileStart *FilePosition `json:"rightFileStart"`
	RightFileEnd   *FilePosition `json:"rightFileEnd"`
}

// FilePosition represents a line position in a file.
type FilePosition struct {
	Line int `json:"line"`
}

// ThreadProps contains additional thread properties.
type ThreadProps struct {
	FilePath *PropValue `json:"FilePath"`
}

// PropValue represents a property value wrapper.
type PropValue struct {
	Value string `json:"$value"`
}

// Comment represents a single comment in a thread.
type Comment struct {
	ID              int     `json:"id"`
	Content         string  `json:"content"`
	CommentType     string  `json:"commentType"`
	PublishedDate   string  `json:"publishedDate"`
	LastUpdatedDate string  `json:"lastUpdatedDate"`
	Author          *Author `json:"author"`
}

// Author represents a comment author.
type Author struct {
	DisplayName string `json:"displayName"`
}

// PRResponse represents the PR details API response.
type PRResponse struct {
	Repository *RepoInfo `json:"repository"`
}

// RepoInfo contains repository information.
type RepoInfo struct {
	ID string `json:"id"`
}

// FetchThreads retrieves PR comment threads from Azure DevOps.
// It handles 404 errors by looking up the PR to get the repository ID.
func (c *Client) FetchThreads(pr *ParsedPR) ([]Thread, error) {
	var threadsResp ThreadsResponse

	// Try fetching threads directly
	err := c.fetchJSON(threadsURL(pr), &threadsResp)
	if err == nil {
		return threadsResp.Value, nil
	}

	// If not a 404, return the error
	if !strings.Contains(err.Error(), "404") {
		return nil, err
	}

	if c.debug && c.debugLog != nil {
		c.debugLog("Thread fetch 404; attempting to resolve PR and retry")
	}

	// Fetch PR details to get repository ID
	var prResp PRResponse
	if err := c.fetchJSON(prURL(pr), &prResp); err != nil {
		return nil, err
	}

	if prResp.Repository == nil || prResp.Repository.ID == "" {
		return nil, fmt.Errorf("PR response missing repository.id")
	}

	// Retry with repository ID
	if err := c.fetchJSON(threadsURLByRepoID(pr, prResp.Repository.ID), &threadsResp); err != nil {
		return nil, err
	}

	return threadsResp.Value, nil
}
