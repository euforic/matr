# Reliability

The repository relies on a small, repeatable review path:

- `go run . validate` for fast structural checks
- `go run . test` for package tests
- `go run . review` for the full local review path, including conventional-commit validation
- [`docs/REVIEW_LOOP.md`](REVIEW_LOOP.md) for manual review questions that tests and CI do not answer by themselves

Reliability expectations:

- keep CI aligned with local commands
- add tests for behavior changes before implementation
- prefer small, explicit checks over broad undocumented expectations
- avoid adding hidden generation or caching behavior without documentation
- add a new validation rule when the same review issue appears repeatedly
