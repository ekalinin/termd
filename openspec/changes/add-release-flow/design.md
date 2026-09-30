# Design

## Context

See proposal.md - Why for the motivation, and `specs/release/spec.md` and `specs/cli/spec.md` for the requirements.

State of the repository that shapes the approach:

- `.github/workflows/ci.yml` runs `make check` in one job `check` on `pull_request` and `push` to `main`. Its concurrency group is `${{ github.workflow }}-${{ github.event.pull_request.number || github.run_id }}` with `cancel-in-progress: true`. The add-ci design left `workflow_call` for the release flow.
- The repository has no tags. Pull requests are merged with merge commits, and the commits follow conventional commits.
- No cgo: every dependency is pure Go. `internal/termbg/tty_unix.go` is built for darwin, linux and the BSDs, `tty_other.go` everywhere else. There is no Windows-specific code; the console mode is never changed.
- `cmd/termd/main.go` parses the flags with its own `flag.FlagSet` in `run`, and takes everything from the outside world through the `env` struct, so the tests fake it.
- `go.mod` declares Go 1.27.1. Since Go 1.24, `go build` records the main module version from git. Checked with Go 1.27.1:
  - a shallow clone with a tag on HEAD records the tag (`v0.0.1-probe`);
  - an untracked `dist/` directory in the same clone records `v0.1.0+dirty`;
  - a commit without a tag records a pseudo-version (`v0.0.0-20260929194532-218cf45f37c5`);
  - `go run` records `(devel)`.
- `internal/site/site_test.go` (`TestInstallCommandMatchesReadme`) fails unless the README has exactly one line that starts with `go install `, equal to the command on the landing page.
- The maintainer's `anygrade` project has the same workflow layout (`ci.yml`, `pages.yml`, `release.yml`) and releases with GoReleaser; it starts the release with a tag push. GoReleaser 2 is installed locally.

## Goals / Non-Goals

**Goals:**
- A failed run leaves nothing public: no tag, no published release.
- One source of the version: the tag. No version constant in the code and no `-ldflags` for it.
- The release build can be reproduced locally with one command, like `make check` for CI.

**Non-Goals:**
- Reproducible (bit-for-bit) archives.
- A dry-run mode of the workflow itself; the local snapshot build covers it.
- Custom release notes text beyond the grouped commit list.

## Decisions

### Manual start with the version as input
`release.yml` is triggered by `workflow_dispatch` with one required string input `version`. The maintainer runs `gh workflow run release.yml -f version=vX.Y.Z` (or uses the Actions UI).
Alternatives:
- Tag push (as in `anygrade`): the tag is public before the checks run. Once a version was fetched through proxy.golang.org, its content is recorded in sum.golang.org, so a broken tag cannot be fixed, only replaced by a new version.
- release-please: a bot keeps a release pull request. More moving parts, and a pull request opened with `GITHUB_TOKEN` does not trigger `ci.yml`, so it needs a personal token or a GitHub App.

### Three jobs: validate, ci, release
```
 validate --> ci (uses ci.yml) --> release
```
- `validate`: checks the branch and the version (see the next decision). Fails fast, before any build.
- `ci`: `uses: ./.github/workflows/ci.yml`, `needs: validate`. `ci.yml` gets `workflow_call:` next to its triggers; nothing else in it changes. The release thus runs exactly `make check`, as the `ci` spec requires for CI.
- `release`: `needs: ci`, the only job with `contents: write`. The workflow default is `contents: read`.

All jobs check out `github.sha`, which for `workflow_dispatch` is the head of the chosen branch at the time of the start, so the checks and the build use the same commit.

`release.yml` has its own concurrency group `release` with `cancel-in-progress: false`: a second release waits for the first one, and its version check then sees the first tag. Inside a called workflow `github.workflow` is the caller's name, so the group of `ci.yml` becomes `Release-<run_id>` and does not collide with `release`.
Alternatives: one job that runs `make check` itself (duplicates the steps of `ci.yml`); `ci` in parallel with `validate` (saves seconds, but an invalid version also starts a full check run).

### Version check in shell with git version sort
The `validate` job checks out with `fetch-depth: 0` and fails with `::error::` naming the failed check when:
1. `github.ref` is not `refs/heads/main`;
2. the version does not match `^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$` (no build metadata: the Go module system does not use it);
3. `refs/tags/<version>` exists;
4. the version is not the greatest after it is added to the existing version tags. The order is `git -c versionsort.suffix=- tag --sort=-v:refname`, restricted to tags that match the regex; the error names the greatest existing tag;
5. the major version is 2 or more and the module path (`go list -m`) does not end with `/v<major>`.

`versionsort.suffix=-` makes git sort `v0.2.0-rc.1` before `v0.2.0` and `rc.2` before `rc.10`; plain `sort -V` puts `v0.2.0` before `v0.2.0-rc.1`.
Alternatives: `golang.org/x/mod/semver` in a small Go program (exact semver order, but a new dependency for one comparison); `npx semver` (a download from npm in every release).

### GoReleaser builds a draft; publishing it is the last step
The `release` job checks out with `fetch-depth: 0`, sets up Go from `go.mod`, creates the tag locally (`git tag <version>`, never pushed), and runs `goreleaser/goreleaser-action` with `release --clean`. A last step publishes the draft with `gh release edit <version> --draft=false`.

GitHub creates the tag of a draft release only when the release is published. Until the last step nothing is public: a failure in the build or the upload leaves at most a draft, visible only to maintainers. `replace_existing_draft: true` lets a new run with the same version replace that draft. `target_commitish: "{{ .FullCommit }}"` fixes the commit of the tag; without it, GitHub would create the tag at the head of `main` at the time of publishing.

The local tag is needed for two reasons: GoReleaser takes the version and the previous tag for the changelog from git, and `go build` records the tag as the module version only when it is on HEAD.
Alternatives:
- `goreleaser release --skip=publish` and `gh release create --draft --target` with the files from `dist/` (no dependency on how GoReleaser creates releases, but the upload, notes and pre-release flag are then done by hand). This is the fallback if GoReleaser refuses a tag that exists only locally, see Risks.
- A hand-written matrix of `go build` plus `gh release create`: no new tool, but archives, checksums and grouped notes would be reimplemented; the maintainer's other projects already use GoReleaser.

### `.goreleaser.yaml`
Following `anygrade`:
- `builds`: `main: ./cmd/termd`, `binary: termd`, `CGO_ENABLED=0`, `-trimpath`, `goos: [linux, darwin, windows]`, `goarch: [amd64, arm64]`. `ldflags: [-s -w]` is set explicitly: the GoReleaser default adds `-X main.version=...` and similar, which termd does not use, because the version comes from the build info.
- `archives`: `tar.gz`, `zip` for windows, `name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"`, `files: [README.md, LICENSE]`.
- `checksum`: `checksums.txt` (SHA-256 is the GoReleaser default).
- `changelog`: `use: git`, `sort: asc`, groups for `feat`, `fix`, `perf`, `refactor`, `docs` and the rest; excluded: `chore`, `Merge pull request`, `Merge branch`.
- `release`: `draft: true`, `replace_existing_draft: true`, `target_commitish: "{{ .FullCommit }}"`, `prerelease: auto`.
- `snapshot`: `version_template: "{{ incpatch .Version }}-next"`.

`/dist/` goes to `.gitignore`: GoReleaser refuses a dirty tree, and an untracked `dist/` makes `go build` record `+dirty`.

### `--version` from the build info
`run` gets a `version` bool flag. Right after the flags are parsed and before any other validation or input, `--version` prints `termd <version>\n` to stdout and returns 0. The version comes through a new `env` field filled in `systemEnv` from `debug.ReadBuildInfo()`: `Main.Version`, or `(devel)` when the build info is missing or the version is empty. The tests set the field directly.
Alternatives: an `internal/version` package with `-ldflags -X` and a build info fallback (as in `anygrade`; since Go 1.24 the build info already has the tag, so the `-ldflags` part repeats it); a `version` subcommand (termd has no subcommands).

### Makefile
- `release-check`: `goreleaser check`.
- `release-snapshot`: `goreleaser release --snapshot --clean`, into `dist/`.

Both with `##` help comments and in `.PHONY`. The `help` format widens from `%-12s` to `%-16s`, because `release-snapshot` has 16 characters. `check` does not change: GoReleaser is not required to work on termd.

### README
- Install: a subsection about the archives on the Releases page: unpack, put `termd` on `PATH`, `checksums.txt`. Windows binaries are best-effort: termd does not enable virtual terminal processing and was not tested on Windows. On macOS, a binary downloaded with a browser is quarantined by Gatekeeper; `xattr -d com.apple.quarantine termd` removes the attribute. The subsection adds no second `go install` line (`TestInstallCommandMatchesReadme`).
- Usage: `termd --version`.
- Development: `make release-check`, `make release-snapshot`, and the release command `gh workflow run release.yml -f version=vX.Y.Z`.

## Risks / Trade-offs

- [GoReleaser refuses to release a tag that exists only in the runner, or creates the tag at once] → before the first release, run GoReleaser with this config locally from a scratch clone with a local tag against the real repository: a draft does not create a tag and is visible only to maintainers; check that no tag appeared, then delete the draft. If it fails, use the fallback from "GoReleaser builds a draft".
- [Published versions are permanent: the Go module proxy keeps them even after the tag and the release are deleted] → the tag is created last, after the checks and the upload.
- [The binaries are not signed; macOS quarantines a binary downloaded with a browser] → the README note; signing and notarization are out of scope.
- [Windows binaries are not tested; consoles without virtual terminal processing show raw escape sequences] → best-effort note in the README; Windows code is out of scope.
- [The tag is created with `GITHUB_TOKEN`, so it does not trigger other workflows] → nothing depends on it; the landing page does not show the version.
- [After the first release `go install ...@latest` installs the latest release, not the head of `main`] → accepted; the README and the landing page keep `@latest`.
- [The notes of a release after its pre-release list only the commits since the pre-release, because GoReleaser takes the nearest previous tag] → accepted.
- [git version sort and semver differ for unusual pre-release identifiers] → accepted; `-rc.N`, `-beta.N` and similar are ordered correctly.

## Migration Plan

1. Merge the change. `workflow_dispatch` works only for a workflow on the default branch, so the release can be started only after the merge.
2. Verify the draft behavior of GoReleaser (see Risks).
3. The maintainer starts the first release with the version of their choice. Check the release: six archives and `checksums.txt`, the tag points to the released commit, `go install github.com/ekalinin/termd/cmd/termd@<version>` followed by `termd --version` prints the version.

Rollback: remove `release.yml`, `.goreleaser.yaml` and the `workflow_call` trigger. Versions that were already published stay in the Go module proxy. `--version` can stay.
