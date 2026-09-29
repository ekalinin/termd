# Proposal

## Why

Nothing runs the checks for a pull request or for a push to `main`: the only workflow builds and deploys the site, and its design left the test suite to the PR testing flow. A regression, unformatted code or an outdated `go.mod` is caught only when the maintainer remembers to run `make check` locally, and the code that is built only on other platforms (`internal/termbg/tty_other.go`) is never compiled on macOS or Linux.

## What Changes

- New GitHub Actions workflow that runs the checks on every pull request to `main` and on every push to `main`. The result is shown as a status on the pull request and on the commit.
- The workflow runs `make check`, the same command a maintainer runs locally.
- `make check` gets two more checks, each as its own Makefile target:
  - module tidiness: fails when `go.mod` or `go.sum` differs from what `go mod tidy` produces (`go mod tidy -diff`);
  - `go vet` for the Windows build (`GOOS=windows go vet ./...`).
- The Development section of the README lists what `make check` runs.

Out of scope:

- Making the check required for merging (branch protection or a ruleset). A failing check does not block the merge.
- Making the Pages deployment wait for the checks. `pages.yml` does not change.
- Running the checks on several operating systems or Go versions, the race detector, additional linters.
- The release flow.

## Capabilities

### New Capabilities
- `ci`: when the checks run, that CI runs the same command as a maintainer, and what that command checks.

### Modified Capabilities

None - the behavior of termd and of the site does not change.

## Impact

- New workflow under `.github/workflows/`.
- `Makefile`: two new targets, `check` depends on them.
- README: the Development section.
- No changes to the Go code, no new Go dependencies.
- `pages.yml` and the `project-site` spec do not change.
