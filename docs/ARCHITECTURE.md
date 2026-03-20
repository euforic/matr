# Architecture

## Domain Map

- `main.go` starts the CLI and delegates to `matr.Run()`.
- `matr/` owns runtime concerns:
  - flag parsing
  - Matrfile discovery and caching
  - wrapper generation
  - task execution helpers
- `parser/` owns source parsing for Matrfiles and returns command metadata.
- `internal/harness/` owns repository-specific validation helpers that support local review workflows.

## Dependency Rules

- `main.go` may depend on `matr/`.
- `matr/` may depend on `parser/`.
- `parser/` must not depend on `matr/` or repository-specific harness code.
- `internal/harness/` must stay independent from CLI runtime internals unless a validation rule becomes core product behavior.

## Cross-Cutting Concerns

- Shell execution belongs in `matr/` or repo task definitions, not in `parser/`.
- Validation for repository workflows belongs in `Matrfile.go` plus small helpers under `internal/harness/`.
- Behavior-changing work should update docs when it changes command expectations, boundaries, or reliability assumptions.

## Command Surface

The repository standardizes on these commands:

- `matr setup`
- `matr validate`
- `matr test`
- `matr review`

CI should call the same validation and test paths instead of re-defining separate logic.
