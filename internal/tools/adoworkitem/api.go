package adoworkitem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/krubenok/toolbox/internal/auth"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	defaultBaseURL     = "https://dev.azure.com"
)

// HTTPError represents a non-2xx response from the Azure DevOps API.
type HTTPError struct {
	StatusCode int
	Status     string
	URL        string
	Body       string
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "http error"
	}
	if e.Body == "" {
		return fmt.Sprintf("request failed (%s): %s", e.Status, e.URL)
	}
	return fmt.Sprintf("request failed (%s): %s: %s", e.Status, e.URL, e.Body)
}

type Client struct {
	auth       *auth.Auth
	baseURL    string
	httpClient *http.Client
	debug      bool
	debugLog   func(string)
}

func NewClient(azAuth *auth.Auth, debug bool, debugLog func(string)) *Client {
	return NewClientWithBaseURL(azAuth, defaultBaseURL, debug, debugLog)
}

func NewClientWithBaseURL(azAuth *auth.Auth, baseURL string, debug bool, debugLog func(string)) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		auth:       azAuth,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		debug:      debug,
		debugLog:   debugLog,
	}
}

func (c *Client) fetchJSON(ctx context.Context, apiURL string, result any) error {
	if c.debug && c.debugLog != nil {
		c.debugLog(fmt.Sprintf("Fetching: %s", apiURL))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.auth.AuthorizationHeader())
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			URL:        apiURL,
			Body:       strings.TrimSpace(string(bodyBytes)),
		}
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) WorkItemURL(parsed *ParsedWorkItem) string {
	return fmt.Sprintf(
		"%s/%s/%s/_apis/wit/workitems/%d?$expand=relations&api-version=7.1-preview.3",
		c.baseURL,
		url.PathEscape(parsed.Organization),
		url.PathEscape(parsed.Project),
		parsed.ID,
	)
}

func (c *Client) WorkItemCommentsURL(parsed *ParsedWorkItem, top int, continuationToken string) string {
	u := fmt.Sprintf(
		"%s/%s/%s/_apis/wit/workItems/%d/comments?api-version=7.1-preview.3",
		c.baseURL,
		url.PathEscape(parsed.Organization),
		url.PathEscape(parsed.Project),
		parsed.ID,
	)
	if top > 0 {
		u += "&$top=" + strconv.Itoa(top)
	}
	if continuationToken != "" {
		u += "&continuationToken=" + url.QueryEscape(continuationToken)
	}
	return u
}

func UIWorkItemURL(baseURL, org, project string, id int) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return baseURL + "/" + url.PathEscape(org) + "/" + url.PathEscape(project) + "/_workitems/edit/" + strconv.Itoa(id)
}

type WorkItemResponse struct {
	ID        int                `json:"id"`
	Rev       int                `json:"rev"`
	URL       string             `json:"url"`
	Fields    map[string]any     `json:"fields"`
	Relations []WorkItemRelation `json:"relations"`
}

type WorkItemRelation struct {
	Rel        string         `json:"rel"`
	URL        string         `json:"url"`
	Attributes map[string]any `json:"attributes"`
}

type WorkItemCommentsResponse struct {
	TotalCount        int               `json:"totalCount"`
	Count             int               `json:"count"`
	Comments          []WorkItemComment `json:"comments"`
	ContinuationToken string            `json:"continuationToken"`
}

type WorkItemComment struct {
	ID           int          `json:"id"`
	Text         string       `json:"text"`
	CreatedDate  string       `json:"createdDate"`
	CreatedBy    *IdentityRef `json:"createdBy"`
	ModifiedDate string       `json:"modifiedDate"`
	ModifiedBy   *IdentityRef `json:"modifiedBy"`
}

type IdentityRef struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
}

func (c *Client) FetchWorkItem(ctx context.Context, parsed *ParsedWorkItem) (*WorkItemResponse, error) {
	var wi WorkItemResponse
	if err := c.fetchJSON(ctx, c.WorkItemURL(parsed), &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}

func (c *Client) FetchAllComments(ctx context.Context, parsed *ParsedWorkItem, maxComments int) ([]WorkItemComment, error) {
	const defaultTop = 200

	var all []WorkItemComment
	var token string
	for {
		var resp WorkItemCommentsResponse
		if err := c.fetchJSON(ctx, c.WorkItemCommentsURL(parsed, defaultTop, token), &resp); err != nil {
			return nil, err
		}

		all = append(all, resp.Comments...)
		if maxComments > 0 && len(all) >= maxComments {
			return all[:maxComments], nil
		}

		if resp.ContinuationToken == "" {
			return all, nil
		}
		token = resp.ContinuationToken
	}
}
