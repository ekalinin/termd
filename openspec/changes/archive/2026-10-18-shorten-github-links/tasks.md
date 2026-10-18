# Tasks

## 1. Shortening

- [x] 1.1 Add `TestGitHubLinks` to `internal/render/render_test.go`: a table of cases that renders a paragraph with `Options{Width: 80, Style: style.Options{Hyperlinks: true}}` and compares the whole output with `"\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\\n"`, for the scenarios of the spec: the pull request (`ekalinin/termd#19`, 17 columns by `text.Width`), `<https://github.com/o/r/issues/3>` (`o/r#3`), the discussion (`o/r#7`), the three comment fragments (`o/r#12 (comment)`), the review (`o/r#12 (review)`), both commit forms (`o/r@0123456`, `o/r@abcdef1`), the repository with a dot (`vercel/next.js#12`) and `http://GitHub.com/o/r/issues/2` (`o/r#2`, hyperlink as written); every URL of "Autolinks that are not shortened" has its URL as its text; `[PR](https://github.com/o/r/pull/1)` and `[https://github.com/o/r/pull/1](https://github.com/o/r/pull/1)` keep their text; with `plain(80)` the bare URL of the pull request is rendered as `https://github.com/ekalinin/termd/pull/19\n`; verify `go test ./internal/render -run TestGitHubLinks` fails on the shortened cases only
- [x] 1.2 Add `internal/render/github.go` with the patterns and `githubText` of design.md, and in the `ast.AutoLink` case of `renderer.inline` replace the label with the result of `githubText` when hyperlinks are enabled, the type is `ast.AutoLinkURL` and the URL matches; verify the test of 1.1 passes and `go test ./...` passes with `git status testdata` clean

## 2. Tables

- [x] 2.1 Add a `table` subtest to `TestGitHubLinks`: the table with the header `PR | Note` and the row `https://github.com/o/r/pull/12 | x`, rendered with hyperlinks, gives after `stripEscapes` the same output as the table with the row `o/r#12 | x` rendered with `plain(80)`, so the first column is 6 columns wide; rendered with `plain(80)`, the cell shows the full URL; verify it passes

## 3. Golden fixture

- [x] 3.1 Add `testdata/autolinks.md` with GitHub autolinks of every form of the spec in a paragraph, a list and a table with a `PR` column, one URL in `<` and `>`, the URLs that are not shortened and a link with its own text; run `make golden`; verify `git status testdata/golden` shows only the six new `autolinks.*.golden` files, the styled files show the short texts with the full URLs in their hyperlinks, and the plain files show every URL in full and have no escape sequences

## 4. Documentation

- [x] 4.1 In the "Links" section of `README.md` add a paragraph: with hyperlinks enabled, a bare URL or a URL in `<` and `>` of a GitHub issue, pull request, discussion or commit is shown as `owner/repo#N`, `owner/repo#N (comment)`, `owner/repo#N (review)` or `owner/repo@` and 7 characters of the SHA, and the hyperlink opens the full URL; links with their own text, other hosts and other GitHub pages do not change; with hyperlinks disabled the URL is shown in full; verify every statement matches `specs/markdown-rendering/spec.md`

## 5. Verification

- [x] 5.1 Run `make check` and `openspec validate shorten-github-links --strict`; verify both pass
- [x] 5.2 Run `make build`, then `./termd --no-pager --width 60 testdata/autolinks.md` in a terminal and with `| cat`; verify that in the terminal the short texts open their full URLs on click and the `PR` column is as wide as `ekalinin/termd#30`, its widest short text, and that the piped output shows every URL in full without escape sequences
- [x] 5.3 Render the README of `github.com/klauspost/compress@v1.20.0` from the module cache with `--hyperlinks=always` and with `--hyperlinks=never`; verify that with hyperlinks its 186 bare links are shown as `klauspost/compress#N` and its links with their own text keep their text, and that without hyperlinks the output is the same as before the change
