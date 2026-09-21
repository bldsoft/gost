# golangci-lint

Source of truth: `.golangci.yml`. Read it before changing Go code.

After any code generation or edits:

1. Do not modify files marked `DO NOT EDIT` / `Code generated` (`**/mocks/`, `**/mock/`, enumer). Lint already excludes them.
2. When writing code bring it in line with the style enforced by `make lint`.
3. Run `golangci-lint run ./...` (or `make lint`) and fix the findings. Do not silence linters with `//nolint` unless you were explicitly asked to.

## Comments and rules language

When generating code anywhere in this repository, write comments in English only, if comments are necessary.

If you encounter a code comment, markdown documentation, or script comment written in Russian, translate it into English while preserving the original meaning, then follow the English version.
