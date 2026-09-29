# Spec Delta

## ADDED Requirements

### Requirement: Version flag
termd SHALL accept `--version`. With it, termd SHALL print `termd <version>` followed by a newline to stdout and exit with status 0 without reading a document. `<version>` SHALL be the version Go records for the main module when the binary is built:
- the tag for a binary from a release archive and for `go install github.com/ekalinin/termd/cmd/termd@<tag>`;
- a pseudo-version for a build from a git checkout whose commit has no tag, with a `+dirty` suffix when the working tree has uncommitted changes;
- `(devel)` when Go records no version.

#### Scenario: Release binary
- **WHEN** the user runs `termd --version` with the binary from the `v0.1.0` release
- **THEN** termd prints `termd v0.1.0` to stdout and exits with status 0

#### Scenario: Build from a checkout
- **WHEN** termd is built with `make build` from a clean checkout of a commit without a tag, and the user runs `termd --version`
- **THEN** termd prints `termd` followed by a pseudo-version that contains the first 12 characters of the commit hash

#### Scenario: No recorded version
- **WHEN** the user runs `go run ./cmd/termd --version`
- **THEN** termd prints `termd (devel)`

#### Scenario: Version with a file argument
- **WHEN** the user runs `termd --version README.md`
- **THEN** termd prints the version, does not read or render `README.md`, and exits with status 0
