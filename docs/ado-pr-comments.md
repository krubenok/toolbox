# ado-pr-comments

Fetch and display pull request comments from Azure DevOps.

## Usage

```bash
toolbox ado-pr-comments <PR_URL> [flags]
```

### Flags

| Flag          | Description                                                                 |
| ------------- | --------------------------------------------------------------------------- |
| `--status`    | Filter by thread status (comma-separated or repeated, e.g., `--status active,fixed`) |
| `--json`      | Output JSON instead of TOON format                                          |
| `--no-filter` | Disable content filtering                                                   |
| `--debug`     | Print debug info to stderr                                                  |

### Examples

```bash
# Fetch all comments from a PR
toolbox ado-pr-comments https://dev.azure.com/org/project/_git/repo/pullrequest/123

# Fetch only active/unresolved threads
toolbox ado-pr-comments https://dev.azure.com/org/project/_git/repo/pullrequest/123 --status active

# Fetch active and pending threads
toolbox ado-pr-comments https://dev.azure.com/org/project/_git/repo/pullrequest/123 --status active,pending

# Output as JSON instead of TOON
toolbox ado-pr-comments https://dev.azure.com/org/project/_git/repo/pullrequest/123 --json

# Disable content filtering
toolbox ado-pr-comments https://dev.azure.com/org/project/_git/repo/pullrequest/123 --no-filter
```

## Supported URL Formats

Both Azure DevOps URL formats are supported:

- `https://dev.azure.com/{org}/{project}/_git/{repo}/pullrequest/{id}`
- `https://{org}.visualstudio.com/{project}/_git/{repo}/pullrequest/{id}`

## Authentication

Authentication is handled automatically in the following order:

1. **Environment variables**: If `AZDO_PAT` or `ADO_PAT` is set, it's used as a Personal Access Token (Basic auth)
2. **Azure CLI**: If logged in via `az login`, a Bearer token is obtained automatically

### Using a Personal Access Token (PAT)

1. Generate a PAT in Azure DevOps with **Code > Read** scope
2. Set the environment variable:
   ```bash
   export AZDO_PAT="your-token-here"
   ```

### Using Azure CLI

1. Install the Azure CLI: `brew install azure-cli`
2. Log in: `az login`
3. The tool will automatically obtain tokens for Azure DevOps

## Output Formats

### TOON (default)

[TOON](https://github.com/toon-format/toon) is a token-optimized notation format designed for LLM workflows. It's more compact than JSON while remaining human-readable.

By default, empty or null fields are omitted from TOON output to minimize token usage.

```
[3]:
  - filePath: /src/main.go
    lineStart: 42
    status: active
    comments[1]{author,published,type,content}:
      "John Doe","2025-01-15T10:30:00Z",text,"Please add error handling here"
```

### JSON

Use `--json` for standard JSON output:

```json
[
  {
    "filePath": "/src/main.go",
    "lineStart": 42,
    "status": "active",
    "comments": [
      {
        "author": "John Doe",
        "published": "2025-01-15T10:30:00Z",
        "type": "text",
        "content": "Please add error handling here"
      }
    ]
  }
]
```

## Configuration

Configuration is stored in `~/.toolbox/ado-pr-comments.json`.

### Status Filtering

You can configure default status filtering to only include specific thread statuses. This is useful if you typically only care about active (unresolved) threads.

#### Configuration Structure

```json
{
  "status": {
    "include": ["active"]
  }
}
```

#### Available Statuses

| Status      | Description                     |
| ----------- | ------------------------------- |
| `active`    | Unresolved/open threads         |
| `fixed`     | Resolved as fixed               |
| `closed`    | Closed threads                  |
| `byDesign`  | Resolved as by design           |
| `pending`   | Pending resolution              |
| `wontFix`   | Resolved as won't fix           |

#### Behavior

- **Empty `include` array** (default): All statuses are included
- **Non-empty `include` array**: Only threads with matching statuses are included
- **`--status` flag**: Overrides config for that invocation
- **Helpful empty output** (TOON only): If a status filter matches 0 threads but the PR has threads in other statuses, the tool prints a 1-line status breakdown before the empty thread list.

#### Examples

Include only active threads by default:

```json
{
  "status": {
    "include": ["active"]
  }
}
```

Include active and pending threads:

```json
{
  "status": {
    "include": ["active", "pending"]
  }
}
```

Override config with CLI flag:

```bash
# Config says "active" only, but fetch all statuses for this run
toolbox ado-pr-comments <PR_URL> --status active,fixed,closed,byDesign,pending,wontFix
```

### Output Field Control

You can control which fields are included in the TOON output. Each field can be set to one of three modes:

| Mode       | Description                                    |
| ---------- | ---------------------------------------------- |
| `always`   | Always include the field, even if empty/null   |
| `notEmpty` | Only include if the field has a value (default)|
| `never`    | Never include the field                        |

#### Configuration Structure

```json
{
  "output": {
    "filePath": "notEmpty",
    "lineStart": "notEmpty",
    "lineEnd": "notEmpty",
    "status": "notEmpty",
    "author": "notEmpty",
    "published": "notEmpty",
    "updated": "notEmpty",
    "type": "notEmpty",
    "content": "notEmpty"
  }
}
```

#### Available Fields

**Thread fields:**
- `filePath` - Path to the file
- `lineStart` - Starting line number
- `lineEnd` - Ending line number
- `status` - Thread status (active, closed, etc.)

**Comment fields:**
- `author` - Comment author name
- `published` - Publication timestamp
- `updated` - Last update timestamp
- `type` - Comment type (text, system, etc.)
- `content` - Comment content

#### Example: Minimal Output

To get the most compact output, you might hide timestamps and type:

```json
{
  "output": {
    "published": "never",
    "updated": "never",
    "type": "never"
  }
}
```

### Content Filtering

You can configure regex patterns to strip boilerplate content from comments. This is useful for removing bot-generated noise like feedback links, disclaimers, or promotional content.

#### Configuration Structure

```json
{
  "filter": {
    "cutPatterns": [],
    "scrubPatterns": [],
    "authorPatterns": []
  }
}
```

#### Filter Options

| Option           | Type       | Description                                                                                          |
| ---------------- | ---------- | ---------------------------------------------------------------------------------------------------- |
| `cutPatterns`    | `string[]` | Regex patterns. Content after the first match is removed.                                            |
| `scrubPatterns`  | `string[]` | Regex patterns. All matches are removed from the content.                                            |
| `authorPatterns` | `string[]` | Regex patterns. Filters only apply to comments from matching authors. Empty array means all authors. |

#### How Filtering Works

1. **Author matching**: If `authorPatterns` is specified, filtering only applies to comments from matching authors
2. **Cut patterns**: The first matching `cutPattern` truncates the content at that point
3. **Scrub patterns**: All `scrubPatterns` matches are removed from the remaining content
4. **Cleanup**: Extra whitespace and newlines are normalized

#### Example: Filtering Bot Comments

This example filters a `reviewbot` boilerplate example:

```json
{
  "filter": {
    "cutPatterns": [
      "(?i)Rate this:",
      "(?i)AI-generated content may be incorrect"
    ],
    "scrubPatterns": [
      "(?i)\\bUseful\\b",
      "(?i)\\bNot useful\\b"
    ],
    "authorPatterns": [
      "(?i)^reviewbot$"
    ]
  }
}
```

A complete example is available at [`examples/ado-pr-comments.json`](../examples/ado-pr-comments.json).

#### Regex Tips

- Use `(?i)` at the start for case-insensitive matching
- Use `\\b` for word boundaries (note: double-escaped in JSON)
- Use `\\s*` for optional whitespace
- Use `\\.` to match a literal dot
- Use `^` and `$` for start/end of string anchors

#### Disabling Filtering

To disable filtering entirely for a single command:

```bash
toolbox ado-pr-comments <PR_URL> --no-filter
```

To disable filtering permanently, create an empty filter config:

```json
{
  "filter": {
    "cutPatterns": [],
    "scrubPatterns": [],
    "authorPatterns": []
  }
}
```

## Output Schema

Each thread in the output contains:

| Field       | Type     | Description                                      |
| ----------- | -------- | ------------------------------------------------ |
| `filePath`  | `string` | Path to the file (if file-level comment)         |
| `lineStart` | `int`    | Starting line number (if line-level comment)     |
| `lineEnd`   | `int`    | Ending line number (if range comment)            |
| `status`    | `string` | Thread status: `active`, `closed`, `fixed`, etc. |
| `comments`  | `array`  | Array of comments in the thread                  |

Each comment contains:

| Field       | Type     | Description                                        |
| ----------- | -------- | -------------------------------------------------- |
| `author`    | `string` | Display name of the comment author                 |
| `published` | `string` | ISO 8601 timestamp when published                  |
| `updated`   | `string` | ISO 8601 timestamp when last updated               |
| `type`      | `string` | Comment type: `text`, `system`, etc.               |
| `content`   | `string` | The comment content (HTML converted to plain text) |
