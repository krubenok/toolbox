package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	toolboxmcp "github.com/krubenok/toolbox/internal/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := toolboxmcp.NewServer()

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
