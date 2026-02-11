# AGENTS.md

This repo is a Go monorepo for the sandbox server, MCP hub, and examples. Keep edits small and aligned with existing conventions. This file summarizes how to build, test, lint, and follow local style rules.

-------------------------------------------------------------------------------
Build, lint, test
-------------------------------------------------------------------------------

Primary make targets (repo root):
- Build binaries: `make build`
- Build Linux binaries: `make build-linux`
- Run sandbox server: `make run-server`
- Run MCP hub: `make run-mcp`
- Run all tests: `make test`
- Lint (golangci-lint): `make lint`
- Format: `make fmt`
- Go modules tidy: `make tidy`

Direct Go commands (equivalent to Makefile):
- Build server: `go build -o bin/sandbox-server ./cmd/sandbox-server`
- Build MCP hub: `go build -o bin/mcp-hub ./cmd/mcp-hub`
- Test all: `go test -v ./...`
- Lint: `golangci-lint run`
- Format: `go fmt ./...`

Run a single test:
- Single package: `go test ./internal/mcp/tools -run TestReadTool_Handler_SimpleRead -v`
- Single file: `go test ./internal/mcp/tools -run TestReadTool_Handler_SimpleRead -v ./internal/mcp/tools/read_test.go`
- Subtest (if present): `go test ./internal/mcp/tools -run TestReadTool_Handler_SimpleRead/CaseName -v`

Optional Docker workflows:
- Build image: `make docker-build` (multi-arch)
- Build image (local load): `make docker-build-load`
- Run image: `make docker-run`
- Compose up/down: `make docker-compose-up` / `make docker-compose-down`

Notes:
- Go version: `go 1.25` (see `go.mod`).
- There is no repo-wide golangci config file; defaults apply.
- There are no Cursor or Copilot rules in this repo.

-------------------------------------------------------------------------------
Code style and conventions (Go)
-------------------------------------------------------------------------------

Imports:
- Use standard Go import grouping: stdlib first, blank line, third-party, blank line, local modules.
- Keep import lists minimal and sorted by gofmt.
- Package paths use module root: `github.com/deep-agent/sandbox/...`.

Formatting:
- Use `gofmt`/`go fmt` for all Go code.
- Tabs are used for indentation (default gofmt output).
- Keep line length reasonable; prefer small helper functions over long blocks.

Types and structs:
- Request payloads are local structs with JSON tags and validation tags (`vd:"len($)>0"`).
- Response payloads use shared types from `pkg/model` and the `model.Response` wrapper.
- Use explicit struct fields rather than `map[string]interface{}` when possible.

Naming:
- Types: `CamelCase` (e.g., `BashHandler`, `GrepOptions`).
- Methods: verbs for actions (e.g., `ReadFile`, `Execute`, `GetContext`).
- JSON fields: lower_snake_case (e.g., `case_insensitive`, `output_mode`).
- Files: short, focused names matching their domain (`read.go`, `write.go`, `grep.go`).

API handlers and errors:
- Use `c.BindAndValidate(&req)` for request parsing/validation.
- On validation errors, respond with `http.StatusBadRequest` and a `model.Response` with `Code: 400`.
- On internal errors, respond with `http.StatusInternalServerError` and `Code: 500`.
- Success responses use `Code: 0` and optional `Message: "success"`.
- Prefer returning early on error; keep handlers linear.

Filesystem rules:
- `filesystem.Manager` owns workspace operations and path validation.
- Path validation uses `filepath.Abs`; be careful to keep workspace boundaries intact if re-enabling checks.
- For read/write ops, return wrapped errors with `fmt.Errorf("...: %w", err)`.

Bash execution:
- `bash.Executor` uses `bash -c` and captures stdout+stderr.
- Timeouts are applied via context; expose timeout inputs in request structs.
- Non-zero exit codes are captured; errors should retain context.

Testing:
- Tests use table-driven style where helpful, but current tests are direct/imperative.
- Use `os.MkdirTemp` and `defer os.RemoveAll` for filesystem tests.
- Use `t.Fatalf` for setup failures, `t.Errorf` for assertions.
- Keep tests in the same package unless an external API is required.

Logging:
- There is no explicit logging style guide; use existing patterns in handlers and services.
- Avoid noisy prints in production paths; rely on structured responses.

Dependencies and modules:
- Keep module graph tidy with `go mod tidy` after adding/removing deps.
- The examples folder contains separate Go modules; run commands in those module roots.

-------------------------------------------------------------------------------
Repository map (high level)
-------------------------------------------------------------------------------

- `cmd/sandbox-server`: main server entrypoint.
- `cmd/mcp-hub`: MCP hub entrypoint.
- `internal/api`: HTTP API handlers and router.
- `internal/api/middleware`: authentication middleware (JWT).
- `internal/services`: business logic (filesystem, bash, browser, terminal, web).
- `internal/mcp`: MCP tool registry and tool handlers.
- `model`: shared API response/request structs.
- `sdk/go`: Go SDK for sandbox API.
- `docs`: MCP tools documentation (tools.json, web_tools.json).
- `examples/*`: separate Go modules with example clients (cdp, filesystem, web).
- `docker/volumes/app.supervisor.d/`: user Supervisord configs (mounted to `/home/sandbox/app.supervisor.d`).
- `docker/volumes/userdata/`: user scripts/binaries (mounted to `/home/sandbox/userdata`).
- `docker/volumes/init.d/`: entrypoint init scripts (mounted to `/docker-entrypoint.d`).

-------------------------------------------------------------------------------
Environment variables (runtime)
-------------------------------------------------------------------------------

Key environment variables for the sandbox container:

- `WORKSPACE`: workspace directory (default `/home/sandbox/workspace`).
- `SUPERVISOR_CONF_DIR`: Supervisord config directory (default `/home/sandbox/app.supervisor.d`).
- `USERDATA_DIR`: user data directory (default `/home/sandbox/userdata`).
- `JWT_SECRET`: HMAC shared secret for JWT authentication (HS256/384/512). Leave empty to disable auth.
- `JWT_AUTH_REQUIRED`: set to `true` to reject all requests when `JWT_SECRET` is not configured.

-------------------------------------------------------------------------------
Agent guidance
-------------------------------------------------------------------------------

- Prefer minimal, focused changes that align with existing patterns.
- Follow existing JSON response shapes and error handling in handlers.
- Avoid changing workspace path validation semantics unless requested.
- Update tests when behavior changes; keep tests deterministic.
- If you need new tooling instructions, update this file and `Makefile` together.
