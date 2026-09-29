# ci Specification

## Purpose

Defines the automatic checks for pull requests and pushes to `main`: when they run, that they are the same checks a maintainer runs locally, and what those checks cover.

## Requirements

### Requirement: Triggers
The checks SHALL run for every pull request to `main`, again for every new commit pushed to that pull request, and for every push to `main`. The result SHALL be shown as a status on the pull request and on the commit. A failing check SHALL NOT block the merge of a pull request.

#### Scenario: Pull request
- **WHEN** a pull request to `main` is opened or a new commit is pushed to it
- **THEN** the checks run, and the pull request shows whether they passed or failed

#### Scenario: Push to main
- **WHEN** a commit is pushed to `main`, including the merge of a pull request
- **THEN** the checks run for that commit, and the commit shows whether they passed or failed

#### Scenario: Failing check
- **WHEN** a check fails for a pull request
- **THEN** the pull request shows the failed status, and the maintainer can still merge it

### Requirement: Local parity
CI SHALL run the checks with the same single command a maintainer runs locally, `make check`, and with the Go version declared in `go.mod`. CI SHALL NOT run checks that `make check` does not run.

#### Scenario: Failure in CI
- **WHEN** the checks fail in CI for a commit
- **THEN** `make check` on that commit fails locally with the same check

#### Scenario: New check
- **WHEN** a check is added to `make check`
- **THEN** CI runs it without a change to the workflow

### Requirement: Check contents
`make check` SHALL fail when any of the following fails:
- the gofmt formatting of the Go code under `cmd/` and `internal/`;
- module tidiness: `go.mod` and `go.sum` SHALL be equal to what `go mod tidy` produces;
- `go vet` for the platform it runs on;
- `go vet` for Windows;
- the tests, including the golden file tests.

`make check` SHALL NOT change files in the working tree.

#### Scenario: Unformatted file
- **WHEN** a Go file under `cmd/` or `internal/` is not formatted with gofmt
- **THEN** `make check` fails and names the file

#### Scenario: Untidy module
- **WHEN** `go.mod` or `go.sum` differs from what `go mod tidy` produces
- **THEN** `make check` fails and shows the difference, and `go.mod` and `go.sum` stay unchanged

#### Scenario: Code built only for Windows
- **WHEN** a file that is built for Windows but not for macOS or Linux does not compile or has a `go vet` finding
- **THEN** `make check` fails on macOS and on Linux

#### Scenario: Output regression
- **WHEN** a change alters termd output and the golden files are not updated
- **THEN** `make check` fails
