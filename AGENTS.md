# Repository Guidelines

## Project Structure & Module Organization

- `cmd/toolbox/`: single CLI entrypoint (`main.go`).
- `internal/cli/`: Cobra command wiring (flags, output, subcommands).
- `internal/tools/<tool>/`: tool implementations (business logic, API clients, formatters).
- `internal/auth/`, `internal/config/`: shared helpers.
- `docs/`: user-facing docs for each tool; `examples/`: sample inputs/outputs.
- `bin/`: local build output (generated; gitignored).

## Build, Test, and Development Commands

This repo targets Go `1.23` (see `go.mod` and CI).

- `make build`: build `bin/toolbox`.
- `make run ARGS="ado-pr-comments <PR_URL>"`: run locally with args.
- `make test`: run unit tests (`go test -v ./...`).
- `make lint`: run `golangci-lint` (version pinned in `.golangci.version`).
- `make fmt`: format (`go fmt ./...`).
- `make tidy`: tidy modules (`go mod tidy`).
- `make install`: install via `go install` (respects `GOBIN`/`GOPATH`).
- `make release-dry`: GoReleaser snapshot build (no publish).

## Coding Style & Naming Conventions

- Format with `make fmt` before pushing; keep changes gofmt-clean.
- Prefer small, focused packages and explicit error handling.
- New tools follow the existing pattern:
  - implementation: `internal/tools/<toolname>/`
  - CLI command: `internal/cli/<command>.go` registered in `internal/cli/root.go`
  - keep `cmd/toolbox/main.go` as the only entrypoint
- Use Go conventions: `MixedCaps` for exported identifiers, `lowerCamel` for locals.

## Testing Guidelines

- Use Go’s standard `testing` package; keep tests close to code as `*_test.go`.
- Prefer table-driven tests for parsers/formatters and add cases for edge inputs.
- Run `make test` locally; CI runs `make build`, `make test`, and lint on PRs.

## Commit & Pull Request Guidelines

- Commits are typically short, imperative subjects; optional prefixes like `fix:`, `feat:`, `docs:`, `test:`, `chore:` are welcome.
- PRs should include: what/why summary, any new flags or behavior changes, and doc updates in `docs/` when user-facing output changes.
- Ensure `make test` and `make lint` are clean before requesting review.

## Security & Configuration Tips

- Never commit secrets. Azure DevOps tools support `AZDO_PAT`/`ADO_PAT` for auth.
- Per-user configuration lives under `~/.toolbox/` (e.g., `~/.toolbox/ado-pr-comments.json`).

## Agent-Specific Notes

- Prefer AST-aware searches with `ast-grep`; use `rg` for plain-text searches.
- Keep patches minimal and aligned with existing structure and conventions.
