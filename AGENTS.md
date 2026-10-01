# golangci-lint

Source of truth: `.golangci.yml`. Read it before changing Go code.

After any code generation or edits:

1. Do not modify files marked `DO NOT EDIT` / `Code generated` (`**/mocks/`, `**/mock/`, enumer). Lint already excludes them.
2. When writing code bring it in line with the style enforced by `make lint`.
3. Run `golangci-lint run ./...` (or `make lint`) and fix the findings. Do not silence linters with `//nolint` unless you were explicitly asked to.

## Comments

A comment is allowed only for a contract or invariant that cannot be expressed in code and whose violation has a concrete failure mode.
