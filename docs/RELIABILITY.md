# Reliability

The repository relies on a small, repeatable review path:

- `matr validate` for fast structural checks
- `matr test` for package tests
- `matr review` for the full local review path, including conventional-commit validation
- [`docs/REVIEW_LOOP.md`](REVIEW_LOOP.md) for manual review questions that tests and CI do not answer by themselves

Use `matr -timeout <duration> ...` when a task legitimately needs longer than the default 5-minute execution window.

When changing timeout behavior, use the manual smoketest in [`examples/timeout-smoketest/Matrfile.go`](../examples/timeout-smoketest/Matrfile.go) to verify both the default timeout path and an explicit override.

Reliability expectations:

- keep CI aligned with local commands
- add tests for behavior changes before implementation
- prefer small, explicit checks over broad undocumented expectations
- avoid adding hidden generation or caching behavior without documentation
- add a new validation rule when the same review issue appears repeatedly
