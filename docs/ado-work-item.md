# ado-work-item

Fetch and display work item details from Azure DevOps, optimized for LLM workflows.

## Usage

```bash
toolbox ado-work-item <WORK_ITEM_URL> [flags]
```

### Flags

| Flag               | Description |
| ------------------ | ----------- |
| `--json`           | Output JSON instead of TOON format |
| `--debug`          | Print debug info to stderr |
| `--no-description` | Do not include the work item description |
| `--no-discussion`  | Do not include work item comments/discussion |
| `--no-children`    | Do not include child work item links |
| `--no-attachments` | Do not include attachment links |
| `--max-comments`   | Maximum number of discussion comments to fetch (0 = no limit) |

### Examples

```bash
toolbox ado-work-item https://dev.azure.com/org/project/_workitems/edit/1144734
toolbox ado-work-item https://org.visualstudio.com/project/_workitems/edit/1144734 --json
toolbox ado-work-item <WORK_ITEM_URL> --no-discussion
toolbox ado-work-item <WORK_ITEM_URL> --max-comments 50
```

## Supported URL Formats

- `https://dev.azure.com/{org}/{project}/_workitems/edit/{id}`
- `https://{org}.visualstudio.com/{project}/_workitems/edit/{id}`

## Authentication

1. **Environment variables**: `AZDO_PAT` or `ADO_PAT` (recommended scope: **Work Items > Read**)
2. **Azure CLI**: `az login`

## Configuration

Configuration is stored in `~/.toolbox/ado-work-item.json`.

Output is emitted in TOON by default, with configurable field inclusion to control token usage. Central sections (`description`, `discussion`, `children`, `attachments`) are included by default even when empty.

### Output Field Control

Each field can be set to one of:

| Mode       | Description |
| ---------- | ----------- |
| `always`   | Always include the field, even if empty |
| `notEmpty` | Only include when a value exists |
| `never`    | Never include the field |

See `examples/ado-work-item.json` for a complete example.
