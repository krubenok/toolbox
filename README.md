# toolbox

`toolbox` is a growing collection of small CLI utilities aimed at improving developer workflows (especially for AI-assisted development).


> [!NOTE]  
> This repository and its tooling are entirely maintained and written by AI. The goal is not to produce "vibe-coded slop," but rather to demonstrate how genuinely useful tools can be created with AI "context engineering". Contributions and feedback are welcome!

## Install

### Homebrew

// TODO
```bash
brew install krubenok/tap/toolbox
```

### Go

```bash
go install github.com/krubenok/toolbox/cmd/toolbox@latest
```

## Usage

```bash
toolbox version
```

### ado-pr-comments

Fetch and display pull request comments from Azure DevOps.

```bash
toolbox ado-pr-comments <PR_URL>
toolbox ado-pr-comments <PR_URL> --status active
toolbox ado-pr-comments <PR_URL> --json
```

More details: `docs/ado-pr-comments.md`.

## Development

```bash
make build
make test
make lint
```

Notes:
- `make install` uses `go install` and respects `GOBIN`/`GOPATH`.
- `make lint` requires `golangci-lint`.

## Adding a new tool

Pattern used in this repo:
- Tool logic lives in `internal/tools/<toolname>/`.
- CLI wiring lives in `internal/cli/` and registers a Cobra subcommand.
- `cmd/toolbox/main.go` stays as the single entrypoint.
