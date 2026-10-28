# Tasks

## 1. Markers and the list

- [x] 1.1 Add `TestTableLinks` to `internal/render/render_test.go`: a table of cases that renders each scenario of `specs/table-rendering/spec.md` with `plain(80)` and compares the whole output: the link in a pipe (`docs [1]`, a blank line, `[1] https://example.com/very/long/path/to/documentation/page`), several links (`Name [1]`, `a [2]`, `[image: logo] [3]`, `b [2]` and the list of three URLs), two tables that each start from `[1]` with their own list, the GitHub autolink (`ekalinin/termd#30 [1]`), the other bare URL and the link whose text is its URL (`example.com [1]`), the cells without a host (`me@example.com`, `a.md`, `empty`, no list) and the HTML link (`guide [1]`); a subtest renders the table of "Table with a long URL fits the output width" with `plain(40)` and checks that every table line is at most 40 columns wide, `docs` is on one line, the list line is in full and `renderer.block` marks the block wide; a subtest renders the same tables with `styled(80)` and checks after `stripEscapes` that there is no marker and no list; verify `go test ./internal/render -run TestTableLinks` fails on the plain cases only
- [x] 1.2 Add `linkRefs` with `marker(url string) int`, the field `refs` of `renderer`, `renderer.link` and `shortText` as in design.md, replace the four calls of `text.LinkSpans` (the `ast.Link` and `ast.AutoLink` cases of `renderer.inline`, `renderer.image`, `inlineHTML.pop`) with `renderer.link`, passing `bare` as `n.AutoLinkType == ast.AutoLinkURL` in the `ast.AutoLink` case; in `renderer.table` set `r.refs` around the cells when hyperlinks are disabled and append the blank line and the `[N] url` lines after the table, marking the block wide when a line is wider than `width`; verify the test of 1.1 passes
- [x] 1.3 Run `go test ./...`; verify that only `TestGolden` fails, and only on the plain files of `tables`, `regression` and `autolinks`

## 2. Golden files

- [x] 2.1 Run `make golden`; verify `git status testdata` shows only `tables.w{40,60,80}.plain.golden`, `regression.w{40,60,80}.plain.golden` and `autolinks.w{40,60,80}.plain.golden` as modified, that the `Ссылка` column of `tables` and `regression` shows `docs [1]` followed by the list, that the `PR` column of `autolinks` shows `ekalinin/termd#30 [1]` and `ekalinin/termd#28 [2]`, and that `go test ./...` passes

## 3. Documentation

- [x] 3.1 In the "Links" section of `README.md` add a paragraph: with hyperlinks disabled, a link or an image in a table cell shows its text and a marker such as `[1]`, the URLs are listed below the table as `[1] url`, a bare URL in a cell shows the GitHub text or the host, and links with the same URL share a marker; remove the plain mode limitation about long URLs in tables from "Limitations"; verify every statement matches `specs/table-rendering/spec.md`

## 4. Verification

- [x] 4.1 Run `make check` and `openspec validate list-table-links-in-plain-mode --strict`; verify both pass
- [x] 4.2 Run `make build`, then `./termd --no-pager --width 40 testdata/tables.md | cat` and `./termd --no-pager --width 40 testdata/autolinks.md | cat`; verify the tables with links are at most 40 columns wide, the URLs are listed below them in full and there are no escape sequences; run the same in a terminal without `| cat` and verify the cells show only the link text, open the full URL on click and have no markers
