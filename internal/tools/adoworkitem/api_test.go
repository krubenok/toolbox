package adoworkitem

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/krubenok/toolbox/internal/auth"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClientFetchWorkItemAndComments(t *testing.T) {
	t.Parallel()

	org := "org"
	project := "project"
	id := 1144734

	wiResp := WorkItemResponse{
		ID:  id,
		Rev: 7,
		URL: "https://dev.azure.com/org/_apis/wit/workItems/1144734",
		Fields: map[string]any{
			"System.Title":        "Example",
			"System.WorkItemType": "User Story",
			"System.State":        "Active",
			"System.Description":  "<div>Hello<br/>World</div>",
			"System.AssignedTo": map[string]any{
				"displayName": "Jane Doe",
			},
		},
		Relations: []WorkItemRelation{
			{
				Rel: "System.LinkTypes.Hierarchy-Forward",
				URL: "https://dev.azure.com/org/_apis/wit/workItems/222",
			},
			{
				Rel: "AttachedFile",
				URL: "vstfs:///WorkItemTracking/Attachment/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				Attributes: map[string]any{
					"name": "file.txt",
				},
			},
		},
	}

	page1 := WorkItemCommentsResponse{
		Count: 1,
		Comments: []WorkItemComment{
			{
				ID:          1,
				Text:        "<div>First</div>",
				CreatedDate: "2025-01-01T00:00:00Z",
				CreatedBy:   &IdentityRef{DisplayName: "A"},
			},
		},
		ContinuationToken: "next",
	}
	page2 := WorkItemCommentsResponse{
		Count: 1,
		Comments: []WorkItemComment{
			{
				ID:          2,
				Text:        "Second",
				CreatedDate: "2025-01-02T00:00:00Z",
				CreatedBy:   &IdentityRef{DisplayName: "B"},
			},
		},
	}

	parsed := &ParsedWorkItem{Organization: org, Project: project, ID: id}
	client := NewClientWithBaseURL(&auth.Auth{Scheme: "Basic", Token: "dummy"}, "https://example.test", false, nil)
	client.httpClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Basic ") && !strings.HasPrefix(got, "Bearer ") {
			t.Fatalf("missing Authorization header: %q", got)
		}

		var payload any
		switch {
		case strings.Contains(r.URL.Path, "/_apis/wit/workitems/"):
			payload = wiResp
		case strings.Contains(r.URL.Path, "/_apis/wit/workItems/") && strings.HasSuffix(r.URL.Path, "/comments"):
			if r.URL.Query().Get("continuationToken") == "next" {
				payload = page2
			} else {
				payload = page1
			}
		default:
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
				Request:    r,
			}, nil
		}

		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(b))),
			Request:    r,
		}, nil
	})

	wi, err := client.FetchWorkItem(context.Background(), parsed)
	if err != nil {
		t.Fatalf("FetchWorkItem: %v", err)
	}
	if wi.ID != id || wi.Rev != 7 {
		t.Fatalf("unexpected work item: %+v", wi)
	}

	comments, err := client.FetchAllComments(context.Background(), parsed, 0)
	if err != nil {
		t.Fatalf("FetchAllComments: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("len(comments)=%d, want 2", len(comments))
	}

	simplified := SimplifyWorkItem(parsed, *wi, comments, defaultBaseURL)
	if simplified.Description != "Hello\nWorld" {
		t.Fatalf("description=%q, want %q", simplified.Description, "Hello\nWorld")
	}
	if len(simplified.Children) != 1 || simplified.Children[0].ID != 222 {
		t.Fatalf("children=%+v, want one child id 222", simplified.Children)
	}
	if len(simplified.Attachments) != 1 || !strings.Contains(simplified.Attachments[0].DownloadURL, "/_apis/wit/attachments/") {
		t.Fatalf("attachments=%+v, want downloadUrl set", simplified.Attachments)
	}
	if len(simplified.Discussion) != 2 || simplified.Discussion[0].Text != "First" {
		t.Fatalf("discussion=%+v, want normalized comment text", simplified.Discussion)
	}
}
