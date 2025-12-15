package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer creates and configures the MCP server with all available tools.
func NewServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "toolbox",
		Version: "0.1.0",
	}, nil)

	// Register tools
	registerAdoPRCommentsTool(server)

	return server
}
