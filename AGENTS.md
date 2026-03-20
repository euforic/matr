# AGENTS

## What This Repo Is

`matr` is a Go CLI that discovers exported handler functions in a `Matrfile`, generates a runtime wrapper, and executes named tasks from the command line.

## Where Context Lives

- Product context: [`docs/PRODUCT_SENSE.md`](docs/PRODUCT_SENSE.md)
- Architecture and boundaries: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- Reliability expectations: [`docs/RELIABILITY.md`](docs/RELIABILITY.md)
- Review checklist and PR path: [`docs/REVIEW_LOOP.md`](docs/REVIEW_LOOP.md)
- Security posture: [`docs/SECURITY.md`](docs/SECURITY.md)
- Active execution plans: [`docs/exec-plans/active/README.md`](docs/exec-plans/active/README.md)
- Completed execution plans: [`docs/exec-plans/completed/README.md`](docs/exec-plans/completed/README.md)
- Generated reports: [`docs/generated/`](docs/generated/)

## Repo Map

- `main.go`: CLI entrypoint.
- `matr/`: runtime, task registration, generated wrapper support, shell helpers.
- `parser/`: parses `Matrfile` source into command metadata.
- `Matrfile.go`: dogfooded repo task surface for setup, validate, test, and review.
- `.github/workflows/`: CI entrypoints. Keep them aligned with local commands.

## Standard Commands

- `matr setup`
- `matr validate`
- `matr test`
- `matr review`

`review` requires a conventional commit subject. It checks `HEAD` by default, or a subject passed as arguments, for example:

`matr review "docs: add harness foundations"`

After `review` passes, run the checklist in [`docs/REVIEW_LOOP.md`](docs/REVIEW_LOOP.md) before merge for changes that affect behavior, boundaries, or repo workflow expectations.

## Boundaries

- Keep `parser/` focused on parsing and command metadata extraction.
- Keep `matr/` focused on task execution, runtime orchestration, and generated wrapper support.
- Avoid importing CLI/runtime concerns into `parser/`.
- Put harness-specific validation logic outside `parser/` and `matr/` unless it is core product behavior.

## Escalation

If a change affects CLI behavior, generated runtime behavior, public task semantics, or package boundaries, update the relevant doc in `docs/` in the same change.
