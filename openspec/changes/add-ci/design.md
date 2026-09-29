# Design

## Context

See proposal.md - Why for the motivation, and `specs/ci/spec.md` for the requirements.

State of the repository that shapes the approach:

- `.github/workflows/pages.yml` is the only workflow. It runs on push to `main`, uses `actions/checkout@v7` and `actions/setup-go@v7` with `go-version-file: go.mod`, and does not run the tests.
- `Makefile` has `check: fmt-check vet test`. `fmt-check` runs `gofmt -l cmd internal` and does not change files. `help` prints the target names with `%-10s`.
- The tests do not need a terminal, `less` or environment variables: the CLI tests use a fake environment, the `termbg` tests a scripted fake terminal, and `internal/site` builds the whole site in `TestBuild`. The full run with `-race` and no cache takes about 5 seconds.
- `internal/termbg/tty_unix.go` is built for darwin, linux and the BSDs, `tty_other.go` for every other platform. `GOOS=windows go vet ./...` passes today.
- `go mod tidy -diff` prints nothing today. The non-test code starts no goroutines.
- The repository is public, `main` has no branch protection.

## Goals / Non-Goals

**Goals:**
- One place defines the checks: the `check` target. The workflow only prepares Go and calls it.
- A maintainer reproduces any CI failure with `make check`.

**Non-Goals:**
- Caching or speeding up beyond what `setup-go` does by default.
- Reusing the job from other workflows. The release flow can add `workflow_call` when it needs it.

## Decisions

### A separate workflow, independent of Pages
New `.github/workflows/ci.yml` with one job `check` on `ubuntu-latest`: checkout, `setup-go` with `go-version-file: go.mod`, `make check`. Triggers: `pull_request` to `main` and `push` to `main`. Permissions: `contents: read`. The action versions are the same as in `pages.yml`.
Alternatives: a test job in `pages.yml` with `needs:` before the deploy, or `pages.yml` triggered by `workflow_run` after CI. Both keep a red `main` off the site, but change `pages.yml` and the `project-site` requirement "Build and publication". A failing commit reaches `main` only by merging a failing pull request, so the gain is small.

### The new checks are Makefile targets inside `check`
Two targets, in the style of `fmt-check`:
- `tidy-check`: `go mod tidy -diff`. It prints the difference and exits non-zero without writing `go.mod` or `go.sum`, unlike the existing `tidy` target.
- `vet-windows`: `GOOS=windows go vet ./...`. It type-checks the files built only for other platforms.

`check` becomes `fmt-check tidy-check vet vet-windows test`: the fast checks first, the tests last. The `help` format widens from `%-10s` to `%-12s`, because `vet-windows` has 11 characters.
Alternatives: run the two commands as separate workflow steps (CI and `make check` would then differ, against "Local parity"); a shorter name such as `vet-win` that fits `%-10s` (less clear).

### Concurrency cancels only pull request runs
The concurrency group is the workflow name plus the pull request number, or plus the run id for a push; `cancel-in-progress` is true. A new commit to a pull request cancels the run for the previous one, while every push to `main` has a group of its own and keeps its own result, as "Triggers" requires.
Alternatives:
- A group per workflow and ref, with `cancel-in-progress` true only for `pull_request` events. GitHub keeps at most one pending run in a group and cancels it when a newer run is queued, even without `cancel-in-progress`, so with three quick pushes to `main` the second commit would be left with a cancelled status instead of pass or fail.
- The same group per ref with `queue: max` for pushes, which keeps up to 100 pending runs. Pushes to `main` would run one after another; not chosen because it is unclear whether `queue` accepts an expression and whether `actionlint` knows the option.
- Cancel for every event with a group per ref (the same problem as the first alternative, also for a running check).

### No required status check
The check is not added to branch protection or a ruleset; see proposal.md - Out of scope. The job name `check` becomes the status check name if this changes later.

## Risks / Trade-offs

- [A pull request run tests the merge with `main` as it was at the time of the run; when `main` moves on, the result can be stale] → the run on the push to `main` checks the merged commit; branches are not required to be up to date.
- [The site is deployed even when the checks fail on `main`] → accepted, see "A separate workflow, independent of Pages".
- [The checks run only on Linux; code built only for macOS is not checked in CI] → darwin and linux share `tty_unix.go`, and the maintainer runs `make check` on macOS.
- [The Windows code is vetted but its tests never run] → accepted; the README already says that background detection works only on macOS, Linux and the BSDs.
- [`go mod tidy -diff` considers all platforms and the tests of dependencies, so with an empty module cache it downloads modules that `go build` does not need, and fails without network] → CI has network access and the `setup-go` cache; locally the cache is already filled.

## Migration Plan

1. Merge the change. The pull request that adds `ci.yml` already runs it, so the first run is visible before the merge.
2. The push to `main` from the merge runs the checks next to the Pages deployment.

Rollback: remove `ci.yml`. The Makefile targets can stay; they change no behavior of termd.
