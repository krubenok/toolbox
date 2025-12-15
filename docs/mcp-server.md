# toolbox-mcp

MCP (Model Context Protocol) server exposing toolbox utilities for AI assistants.

## Build

```bash
make build-mcp
```

This produces `bin/toolbox-mcp`.

## Configuration

### VS Code

Add the server to your workspace `.vscode/mcp.json`:

```json
{
  "servers": {
    "toolbox": {
      "command": "toolbox-mcp"
    }
  }
}
```

If `toolbox-mcp` is not on your `PATH`, use an absolute path (for example, `.../toolbox/bin/toolbox-mcp`).

Or install via the VS Code CLI:

```bash
code --add-mcp '{"name":"toolbox","command":"toolbox-mcp"}'
```

### Codex CLI

Add to `~/.codex/config.toml`:

```toml
[mcp_servers.toolbox]
command = "toolbox-mcp"
```

Or use the built-in config helper:

```bash
codex mcp add toolbox -- toolbox-mcp
```

### Claude Code

Add to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "toolbox": {
      "command": "toolbox-mcp"
    }
  }
}
```

## Available Tools

### ado_pr_comments

Fetch pull request comments from Azure DevOps.

#### Parameters

| Parameter   | Type       | Required | Description                                                     |
| ----------- | ---------- | -------- | --------------------------------------------------------------- |
| `pr_url`    | `string`   | Yes      | Azure DevOps PR URL                                             |
| `statuses`  | `string[]` | No       | Filter by status (active/fixed/closed/byDesign/pending/wontFix) |
| `format`    | `string`   | No       | Output format: `toon` (default) or `json`                       |
| `no_filter` | `boolean`  | No       | Disable content filtering from config                           |

#### Example Usage

```json
{
  "pr_url": "https://dev.azure.com/org/project/_git/repo/pullrequest/123",
  "statuses": ["active"],
  "format": "toon"
}
```

#### Authentication

The tool uses the same authentication as the CLI:

1. **Environment variables**: `AZDO_PAT` or `ADO_PAT` for Personal Access Token
2. **Azure CLI**: Automatic Bearer token via `az login`

See [ado-pr-comments.md](./ado-pr-comments.md) for detailed authentication setup.

## Adding New Tools

To add a new tool to the MCP server:

1. Create a handler in `internal/mcp/<toolname>.go`
2. Register it in `internal/mcp/server.go` via `registerXxxTool(server)`

The handler should follow the pattern in `ado_pr_comments.go`:
- Define an input struct with `jsonschema` tags
- Use `mcp.AddTool` with a typed handler function
- Return errors via `IsError: true` in the result (not as Go errors)
