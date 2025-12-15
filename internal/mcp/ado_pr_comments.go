package mcp

import (
	"context"

	"github.com/kyrubeno/toolbox/internal/tools/adoprcomments"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AdoPRCommentsInput defines the input schema for the ado_pr_comments tool.
type AdoPRCommentsInput struct {
	// Azure DevOps PR URL (required)
	PRURL string `json:"pr_url" jsonschema:"Azure DevOps PR URL"`
	// Filter by thread status (optional)
	Statuses []string `json:"statuses,omitempty" jsonschema:"Returns only comment threads with this status. If ommited returns only threads with status 'active'. Values: active/fixed/closed/byDesign/pending/wontFix."`
	// Output format: toon (default) or json
	Format string `json:"format,omitempty" jsonschema:"Defaults to a more efficient output format. Set to 'json' to get raw JSON output."`
	// Disable content filtering
	NoFilter bool `json:"no_filter,omitempty" jsonschema:"Response is filtered by default to remove null or emtpy fiels and other low-value data. Set no_filter to true to disable this behavior."`
}

// registerAdoPRCommentsTool registers the ado_pr_comments tool with the server.
func registerAdoPRCommentsTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ado_pr_comments",
		Description: "Fetch pull request comments from Azure DevOps. Returns comment threads with author, content, status, and file location information. By default, returns only comments with status 'actvive'.",
	}, handleAdoPRComments)
}

// handleAdoPRComments handles the ado_pr_comments tool invocation.
func handleAdoPRComments(ctx context.Context, req *mcp.CallToolRequest, input AdoPRCommentsInput) (*mcp.CallToolResult, any, error) {
	if input.PRURL == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "error: pr_url is required"},
			},
			IsError: true,
		}, nil, nil
	}

	opts := adoprcomments.Options{
		Ctx:        ctx,
		PRURL:      input.PRURL,
		Statuses:   input.Statuses,
		OutputJSON: input.Format == "json",
		NoFilter:   input.NoFilter,
	}

	result, err := adoprcomments.Run(opts)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	contents := make([]mcp.Content, 0, 2)
	if input.Format != "json" && result.Summary != "" {
		contents = append(contents, &mcp.TextContent{Text: result.Summary})
	}
	contents = append(contents, &mcp.TextContent{Text: result.Output})

	return &mcp.CallToolResult{
		Content: contents,
	}, nil, nil
}
