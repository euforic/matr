# Reliability

The repository relies on a small, repeatable review path:

- `go run . validate` for fast structural checks
- `go run . test` for package tests
- `go run . review` for the full local review path, including conventional-commit validation

Reliability expectations:

- keep CI aligned with local commands
- add tests for behavior changes before implementation
- prefer small, explicit checks over broad undocumented expectations
- avoid adding hidden generation or caching behavior without documentation
