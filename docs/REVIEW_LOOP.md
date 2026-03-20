# Review Loop

Use this loop for any meaningful change, especially changes that touch CLI behavior, generated runtime behavior, package boundaries, or repository workflow expectations.

## Required Checks

- `matr validate` passes locally
- `matr test` passes when behavior changed or when `validate` is not enough
- CI reruns the same core validation path
- the diff matches the stated task
- changed behavior has matching tests or targeted eval coverage
- changed interfaces, commands, or invariants have matching docs updates

## Checklist

1. Does the diff match the user request or active plan?
2. Does [`AGENTS.md`](../AGENTS.md) still point to the right context?
3. Does the change preserve the boundaries in [`docs/ARCHITECTURE.md`](ARCHITECTURE.md)?
4. Did `matr review` run cleanly?
5. Did the change introduce stale docs, missing docs, or broken links?
6. Is there a repeated failure pattern that should become a check in `validate` instead of another manual review comment?

If any answer is no, fix the source of the mismatch instead of adding a one-off exception.

## Lightweight Evals

Use a small eval prompt when tests alone do not prove the change is correct.

- "Show which doc changed because of this new command or invariant."
- "Explain why this import path does not violate the layer model."
- "Walk the happy path for the changed command and identify where validation happens."
- "List the user-visible behavior change and the test or doc that covers it."

Keep evals concrete and tied to a specific risk.

## PR Path

1. Run `matr review` locally.
2. Compare the diff to the task or plan.
3. Run the checklist above.
4. Let CI rerun the same core validation path.
5. Fix review findings before merge.
