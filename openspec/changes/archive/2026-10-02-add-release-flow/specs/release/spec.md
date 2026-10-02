# Spec Delta

## Purpose

Defines how a termd release is published: how a maintainer starts it, what is checked before anything is published, when the version tag becomes public, what a release contains, and how the release build is reproduced locally.

## ADDED Requirements

### Requirement: Start
A release SHALL be started only manually by a maintainer, with the version as input, and SHALL be built from the commit at the head of `main` at the time of the start. A run started for any other branch SHALL fail without publishing anything.

#### Scenario: Start from main
- **WHEN** a maintainer starts the release for `main` with the version `v0.1.0`
- **THEN** the release is built from the commit at the head of `main`

#### Scenario: Start from another branch
- **WHEN** a maintainer starts the release for a branch other than `main`
- **THEN** the run fails with an error that names the branch, and nothing is published

### Requirement: Version check
Before anything is built, the release SHALL fail without publishing anything when the version:
- is not a semantic version with a leading `v`, either `vX.Y.Z` or `vX.Y.Z-<pre-release>`;
- already exists as a tag;
- is not greater than the greatest existing version tag;
- has a major version of 2 or more while the module path in `go.mod` has no `/vN` suffix for that major version.

The error SHALL name the failed check. When the repository has no version tags, any version that passes the other checks SHALL be accepted.

#### Scenario: First release
- **WHEN** the repository has no version tags and a maintainer starts the release with `v0.1.0`
- **THEN** the version check passes

#### Scenario: Invalid format
- **WHEN** a maintainer starts the release with `0.1.0` or `v0.1`
- **THEN** the run fails with an error about the version format, and nothing is published

#### Scenario: Existing tag
- **WHEN** the tag `v0.1.0` exists and a maintainer starts the release with `v0.1.0`
- **THEN** the run fails with an error that the tag exists, and nothing is published

#### Scenario: Version not greater than the latest
- **WHEN** the greatest version tag is `v0.2.0` and a maintainer starts the release with `v0.1.5`
- **THEN** the run fails with an error that names `v0.2.0`, and nothing is published

#### Scenario: Pre-release before its release
- **WHEN** the greatest version tag is `v0.2.0-rc.1` and a maintainer starts the release with `v0.2.0`
- **THEN** the version check passes

#### Scenario: Major version without module path suffix
- **WHEN** the module path in `go.mod` is `github.com/ekalinin/termd` and a maintainer starts the release with `v2.0.0`
- **THEN** the run fails with an error about the module path, and nothing is published

### Requirement: Checks before the build
The release SHALL run the same checks as CI runs for pull requests and pushes to `main` (`make check`) on the commit it releases. When the checks fail, the release SHALL NOT build or publish anything.

#### Scenario: Failing check
- **WHEN** `make check` fails on the head of `main` and a maintainer starts the release
- **THEN** the run fails, and no tag and no release are published

### Requirement: Tag publication
The version tag SHALL appear in the repository only after the checks and the build passed and every file of the release was uploaded, and SHALL point to the released commit. When a run fails at any step, the repository SHALL have no tag and no published release for that version, and a new run with the same version SHALL be able to publish it.

#### Scenario: Successful release
- **WHEN** a release run for `v0.1.0` finishes successfully
- **THEN** the tag `v0.1.0` points to the released commit, and the release `v0.1.0` is published with all its files

#### Scenario: Failure after the checks
- **WHEN** a release run for `v0.1.0` fails during the build or the upload
- **THEN** the repository has no tag `v0.1.0` and no published release `v0.1.0`, and a new run with `v0.1.0` can publish it

### Requirement: Release contents
A published release SHALL contain:
- release notes with the changes since the previous version tag, grouped by conventional commit type, without `chore` commits and merge commits;
- one archive for each of linux, darwin and windows on amd64 and on arm64, named `termd_<version>_<os>_<arch>`, where `<version>` is the version without the leading `v`; `.tar.gz` for linux and darwin, `.zip` for windows;
- in each archive, the `termd` binary (`termd.exe` for windows), `README.md` and `LICENSE`;
- a checksums file with the SHA-256 of every archive.

A version with a pre-release part SHALL be published as a pre-release.

#### Scenario: Archives
- **WHEN** the release `v0.1.0` is published
- **THEN** it has the archives `termd_0.1.0_linux_amd64.tar.gz`, `termd_0.1.0_linux_arm64.tar.gz`, `termd_0.1.0_darwin_amd64.tar.gz`, `termd_0.1.0_darwin_arm64.tar.gz`, `termd_0.1.0_windows_amd64.zip` and `termd_0.1.0_windows_arm64.zip`, and a checksums file

#### Scenario: Checksums
- **WHEN** a user downloads an archive and the checksums file of a release
- **THEN** the SHA-256 of the archive matches its line in the checksums file

#### Scenario: Pre-release
- **WHEN** a release run for `v0.2.0-rc.1` finishes successfully
- **THEN** the release is marked as a pre-release

### Requirement: Released version
A binary from a release archive and a binary installed with `go install github.com/ekalinin/termd/cmd/termd@<version>` SHALL report the released version with `termd --version`.

#### Scenario: Binary from an archive
- **WHEN** a user unpacks the linux amd64 archive of `v0.1.0` and runs `termd --version`
- **THEN** termd prints `termd v0.1.0`

#### Scenario: Installed with go install
- **WHEN** a user runs `go install github.com/ekalinin/termd/cmd/termd@v0.1.0` and then `termd --version`
- **THEN** termd prints `termd v0.1.0`

### Requirement: Local dry run
A maintainer SHALL be able to validate the release configuration with one command, and to run the full release build locally with another command, without publishing anything, into a directory that git ignores.

#### Scenario: Snapshot build
- **WHEN** a maintainer runs the local release build
- **THEN** the same archives and checksums file as in a release are produced in the ignored directory, nothing is published, and `git status` shows no changes
