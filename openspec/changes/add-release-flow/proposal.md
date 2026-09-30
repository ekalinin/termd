# Proposal

## Why

termd has no releases: the repository has no tags, `go install ...@latest` installs the head of `main`, and the only way to install termd is with a Go toolchain. termd also has no `--version`, so a user cannot tell which version is installed.

## What Changes

- New GitHub Actions workflow that publishes a release when a maintainer starts it manually with a version (`workflow_dispatch`, for example `gh workflow run release.yml -f version=v0.1.0`). It runs for `main` only.
- Before the build the workflow checks the version: format `vX.Y.Z` or `vX.Y.Z-<pre-release>`, the tag does not exist yet, the version is greater than the latest tag, and the major version is below 2 while the module path has no `/vN` suffix.
- The release runs the same checks as CI (`make check`) before anything is built. The version tag becomes public only after the checks and the build passed and every file was uploaded, so a failed run publishes nothing and can be restarted with the same version.
- A release contains a GitHub Release with notes grouped by the conventional commit type, archives with prebuilt binaries for linux, darwin and windows on amd64 and arm64, and a checksums file. The archives also hold `README.md` and `LICENSE`. GoReleaser builds and publishes them.
- `termd --version` prints the version and exits. The version comes from the Go build info, so it is the tag both for `go install ...@vX.Y.Z` and for the release binaries, without `-ldflags`.
- `ci.yml` becomes callable from other workflows (`workflow_call`), so the release reuses the `check` job.
- Makefile: `release-check` validates the GoReleaser config, `release-snapshot` runs the full release build into `dist/` without publishing.
- `.gitignore`: `/dist/`.
- README: prebuilt binaries in the Install section with Windows marked as best-effort, `--version` in the Usage section.

Out of scope:

- Homebrew tap and other package managers.
- Code signing and macOS notarization.
- Windows-specific code (enabling virtual terminal processing in the console) and running the tests on Windows.
- Changes to the landing page: it keeps the single `go install` command.
- Major version 2 and the `/v2` module path.

## Capabilities

### New Capabilities
- `release`: how a release is started, what is checked before anything is published, when the version tag becomes public, and what a release contains.

### Modified Capabilities
- `cli`: new `--version` flag.

## Impact

- New `.github/workflows/release.yml` and `.goreleaser.yaml`.
- `.github/workflows/ci.yml`: the `workflow_call` trigger. The `ci` spec does not change.
- `cmd/termd/main.go` and its tests: the `--version` flag.
- `Makefile`: two new targets. `.gitignore`: `/dist/`.
- README: the Install and Usage sections.
- No new Go dependencies. GoReleaser is needed locally only for `release-check` and `release-snapshot`.
- After the first release, `go install ...@latest` installs the latest release instead of the head of `main`.
- `pages.yml` and the `project-site` spec do not change.
