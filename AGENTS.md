# golangci-lint

Source of truth: `.golangci.yml`. Read it before changing Go code.

After any code generation or edits:

1. Do not modify files marked `DO NOT EDIT` / `Code generated` (`**/mocks/`, `**/mock/`, enumer). Lint already excludes them.
2. Bring handwritten code in line with the style enforced by `make lint`.
3. Imports follow gci: std → everything else → `github.com/bldsoft`.
4. Run `golangci-lint run ./...` (or `make lint`) and fix the findings. Do not silence linters with `//nolint` unless you were explicitly asked to.
