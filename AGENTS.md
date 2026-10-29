# golangci-lint

Source of truth: `.golangci.yml`. Read it before changing Go code.

After any code generation or edits:

1. Do not modify files marked `DO NOT EDIT` / `Code generated` (`**/mocks/`, `**/mock/`, enumer). Lint already excludes them.
2. When writing code bring it in line with the style enforced by `make lint`.
3. Run `golangci-lint run ./...` (or `make lint`) and fix the findings. Do not silence linters with `//nolint` unless you were explicitly asked to.

## Comments

A comment is allowed only for a contract or invariant that cannot be expressed in code and whose violation has a concrete failure mode.

# Building a service on gost

Read `docs/agent-guide.md` first, then `docs/quickstart.md`. The invariants and checklists there
apply to any service that imports this module; `examples/scaffold/AGENTS.md` is the file a new
service repository starts from. `cmd/scaffold` copies it.

# Example service

`examples/notes` is a complete service inside this module. It is built, linted and tested by CI
(`go build ./examples/...`, `golangci-lint run ./...`, `gotestsum ... ./...`, and a Docker build).
When you change a public API, update the example and the relevant `docs/*.md` page in the same
change. Mocks for it are configured in `.mockery.yml` and regenerated with `go tool mockery`.

# Docs conventions

- Code shown in `docs/` must exist in `examples/notes` or in a gost `_test.go`; mark the origin
  with `<!-- source: path -->`.
- `examples/scaffold/golangci.yml` must stay byte identical to `.golangci.yml`; CI checks it.
- Known quirks of gost are documented in `docs/best-practices.md` and `docs/server.md`; do not
  change library behaviour as part of a documentation change.
