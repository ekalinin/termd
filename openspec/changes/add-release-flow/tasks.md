# Tasks

## 1. `--version`

- [x] 1.1 In `cmd/termd/main.go` add an `env` field with the version, filled in `systemEnv` from `debug.ReadBuildInfo()` (`Main.Version`, or `(devel)` when the build info is missing or the version is empty), and a `version` bool flag handled right after the flags are parsed: print `termd <version>\n` to stdout and return 0; add tests to `cmd/termd/main_test.go` with a fake version `v1.2.3` for `--version` alone and for `--version README.md` (stdout is `termd v1.2.3\n`, status 0, neither stdin nor the file is read); verify `go test ./cmd/termd` passes
- [x] 1.2 Verify the version sources by hand: `go run ./cmd/termd --version` prints `termd (devel)`; `make build && ./termd --version` on a clean checkout prints a pseudo-version that contains the first 12 characters of `git rev-parse HEAD`; in a scratch clone with a local tag `v0.0.1-test` on HEAD, a binary built with `go build ./cmd/termd` prints `termd v0.0.1-test`

## 2. GoReleaser and Makefile

- [x] 2.1 Add `/dist/` to `.gitignore`; verify that after `mkdir dist && touch dist/x` the output of `git status --short` does not mention `dist` (remove `dist/` afterwards)
- [x] 2.2 Add `.goreleaser.yaml` as described in design.md (builds, archives, checksum, changelog, release, snapshot); verify `goreleaser check` passes
- [x] 2.3 Add the `release-check` and `release-snapshot` targets with `##` help comments, add them to `.PHONY`, and widen the `help` format from `%-12s` to `%-16s`; verify `make help` shows every description in one column, `make release-check` passes, and `make release-snapshot` produces in `dist/` the six archives named as in the "Release contents" requirement (`tar.gz` for linux and darwin, `zip` for windows), each with `termd` (`termd.exe` for windows), `README.md` and `LICENSE`, and a `checksums.txt` that `sha256sum -c` accepts; `git status` shows no changes afterwards
- [x] 2.4 Verify the version recorded by the GoReleaser build: in a scratch clone of the committed branch with a local tag `v0.0.1-test` on HEAD, run `goreleaser release --snapshot --clean`; `go version -m` on the darwin arm64 binary shows `mod github.com/ekalinin/termd v0.0.1-test` without `+dirty`, and that binary prints `termd v0.0.1-test` for `--version`

## 3. Workflows

- [x] 3.1 Add `workflow_call:` to the triggers of `.github/workflows/ci.yml`, with no other change; verify the file passes `actionlint` (for example `go run github.com/rhysd/actionlint/cmd/actionlint@latest`)
- [x] 3.2 Add `.github/workflows/release.yml` as described in design.md: name `Release`; `workflow_dispatch` with the required input `version`; permissions `contents: read`; concurrency group `release` with `cancel-in-progress: false`; jobs `validate` (checkout with `fetch-depth: 0`, `setup-go` with `go-version-file: go.mod`, the five checks with `::error::` messages), `ci` (`uses: ./.github/workflows/ci.yml`, `needs: validate`) and `release` (`needs: ci`, `contents: write`, checkout with `fetch-depth: 0`, `setup-go`, local `git tag`, `goreleaser/goreleaser-action` with `version: "~> v2"` and `args: release --clean`, then `gh release edit <version> --draft=false`); action versions as in `ci.yml` and `pages.yml`; verify the file passes `actionlint`
- [x] 3.3 Verify the version check of the `validate` job by running its script locally in a scratch repository, once for each scenario of the "Version check" and "Start" requirements: first release `v0.1.0` passes; `0.1.0` and `v0.1` fail on the format; an existing `v0.1.0` fails; `v0.1.5` after `v0.2.0` fails and names `v0.2.0`; `v0.2.0` after `v0.2.0-rc.1` passes; `v2.0.0` with the module path `github.com/ekalinin/termd` fails on the module path; a ref other than `refs/heads/main` fails and names the branch

## 4. README

- [x] 4.1 Update the README as described in design.md: the prebuilt binaries subsection in Install (Releases page, `checksums.txt`, Windows best-effort, the macOS quarantine note), `termd --version` in Usage, `make release-check`, `make release-snapshot` and `gh workflow run release.yml -f version=vX.Y.Z` in Development; verify `go test ./internal/site` passes (the README keeps a single `go install` line) and `make check` passes

## 5. Before the merge

- [x] 5.1 Verify that GoReleaser creates a draft without a tag: from a scratch clone of the committed branch with a local tag `v0.0.0-draft.1` on HEAD, run `GITHUB_TOKEN=$(gh auth token) goreleaser release --clean`; a draft `v0.0.0-draft.1` appears on the Releases page with the uploaded files and the commit of HEAD as its target, and `git ls-remote --tags origin` shows no tag; delete the draft with `gh release delete v0.0.0-draft.1 --yes` (if a tag appeared, delete it at once with `git push origin :refs/tags/v0.0.0-draft.1`, switch the `release` job to the fallback from design.md and repeat)
- [x] 5.2 Open the pull request with this change; verify the `check` job runs for it and passes (`gh pr checks`)

## 6. After the merge

- [ ] 6.1 The maintainer starts the first release with `gh workflow run release.yml -f version=<version>`; verify the run passes, the release has the six archives and `checksums.txt`, `git ls-remote --tags origin` shows the tag at the released commit, and `go install github.com/ekalinin/termd/cmd/termd@<version>` followed by `termd --version` prints `termd <version>`
- [ ] 6.2 Verify that failed starts publish nothing: start the release again with the same version, and once for another branch (`--ref <branch>`); both runs fail in `validate` with the matching error, and no new tag or release appears
