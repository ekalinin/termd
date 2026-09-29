# Tasks

## 1. Makefile

- [x] 1.1 Add a `tidy-check` target that runs `go mod tidy -diff`, with a `##` help comment, and add it to `.PHONY`; verify `make tidy-check` passes on the current tree, and that after temporarily deleting one `// indirect` line from `go.mod` it fails, prints the difference and leaves `go.mod` as edited (restore the line afterwards)
- [x] 1.2 Add a `vet-windows` target that runs `GOOS=windows go vet ./...`, with a `##` help comment, and add it to `.PHONY`; verify `make vet-windows` passes, and that a temporary `go vet` finding in `internal/termbg/tty_other.go` (for example a `fmt.Printf` with a mismatched verb) makes it fail on macOS while `make vet` passes (remove the finding afterwards)
- [x] 1.3 Change `check` to `fmt-check tidy-check vet vet-windows test` and the `help` format from `%-10s` to `%-12s`; verify `make help` shows every description in one column, `make check` passes, and `git status` shows no changes made by `make check`

## 2. README

- [x] 2.1 Update the `make check` line in the Development section to list the format check, the module tidiness check, `go vet` for the host and for Windows, and the tests; verify the list matches the prerequisites of `check` in the Makefile

## 3. Workflow

- [x] 3.1 Add `.github/workflows/ci.yml`: name `CI`; triggers `pull_request` and `push` to `main`; permissions `contents: read`; a concurrency group of the workflow name plus the pull request number or, for a push, the run id, with `cancel-in-progress` true; one job `check` on `ubuntu-latest` with `actions/checkout@v7`, `actions/setup-go@v7` with `go-version-file: go.mod`, and `make check`; verify the file parses with `actionlint` (for example `go run github.com/rhysd/actionlint/cmd/actionlint@latest`)
- [x] 3.2 Open the pull request with this change; verify the `check` job runs for it and passes (`gh pr checks`)
- [x] 3.3 After the merge, verify the push to `main` ran the `check` job for the merge commit and it passed, next to the Pages deployment (`gh run list --branch main`)
