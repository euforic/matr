# Harness Engineering Assessment

Date: 2026-03-19
Repo: `.`
Overall score: `3/21`

## Scorecard

| Dimension | Score | Evidence |
| --- | --- | --- |
| Repo map | 0 | No `AGENTS.md` or equivalent repo-map file found via `rg --files -g 'AGENTS.md'`. |
| Searchable docs | 1 | [`README.md`](../../README.md) documents install and usage, but no architecture, product constraints, or active-work docs exist under `docs/`. |
| Standard commands | 1 | CI runs `go build -v ./...` and `go test -v ./...` in [`.github/workflows/go.yml`](../../.github/workflows/go.yml), and [`Matrfile.go`](../../Matrfile.go) has a `Test` target, but there is no standard setup, validate, or review entrypoint. |
| Architecture constraints | 0 | No architecture document, boundary policy, or check-enforced layering rules were found. |
| Boundary validation | 0 | [`parser/parser.go`](../../parser/parser.go) validates the Matrfile build tag, but there are no repo-level checks enforcing domain or external-data boundaries. |
| Review loops | 1 | Pull requests trigger build and test in [`.github/workflows/go.yml`](../../.github/workflows/go.yml), but there is no lint, coverage, policy check, conventional-commit enforcement, or explicit review checklist/eval loop. |
| Entropy control | 0 | No freshness checks, maintenance jobs, generated status docs, or quality tracking files were found. |

## Top Gaps

### 1. Missing repo map and deeper navigation docs

Why it matters: Agents have no concise entrypoint for repository layout, key packages, or safe change paths, so every task starts with rediscovery.

Concrete next step: Add `AGENTS.md` with a short repo map pointing to package-level responsibilities, the standard command surface, and any future architecture docs.

### 2. No standardized validation and review surface

Why it matters: The repo has working CI, but the expected local path for setup, validation, testing, and review is implicit and incomplete.

Concrete next step: Add a small standard command surface such as `make setup`, `make test`, `make validate`, and document those commands in `README.md` and `AGENTS.md`.

This standard path should also define a conventional-commits requirement and enforce it in the review path, for example with commit-message linting in CI or a documented pre-push check.

### 3. No documented or enforced architecture boundaries

Why it matters: The current structure is simple, but there is nothing that tells an agent what must stay isolated between CLI entrypoint, task runtime, and parser behavior, or how that separation will be checked as the codebase grows.

Concrete next step: Write a brief architecture note covering package responsibilities and add at least one enforceable check tied to that note, such as package-level test expectations or a validation script that fails on prohibited dependencies.

## Recommended Next Skill

Run `harness-engineering-scaffolding-foundations`.

Reason: This repo is missing the basics first. A concise repo map, searchable architecture notes, and a standard command surface will raise agent reliability faster than adding heavier enforcement or operations layers right now.
