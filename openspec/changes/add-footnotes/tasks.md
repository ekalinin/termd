# Tasks

## 1. Parsing

- [x] 1.1 Extend `TestParseGFM` in `internal/render/render_test.go` with `Text[^1].` and the definition `[^1]: Note.`; the AST must contain an `east.FootnoteLink` and an `east.FootnoteList`; verify `go test ./internal/render -run TestParseGFM` fails
- [x] 1.2 Add `extension.Footnote` to `Parse` in `internal/render/render.go`; verify `TestParseGFM` passes and `go test ./internal/render ./internal/site` passes with unchanged golden files

## 2. References

- [x] 2.1 Add `TestFootnoteReferences` to `internal/render/render_test.go`: in plain mode `Text[^1].` with `[^1]: Note.` starts with `Text[1].` and contains neither `^1` nor `(Note.)`, `A[^b] B[^a] C[^b]` with `[^a]: Alpha.` and `[^b]: Beta.` starts with `A[1] B[2] C[1]`, `Text[^x].` without a definition renders as `Text[^x].`; in styled mode with hyperlinks the bug case contains `[1]` and no OSC 8 sequence; verify `go test ./internal/render` fails
- [x] 2.2 Render `east.FootnoteLink` as the span `[N]` in the surrounding style and `east.FootnoteBacklink` as nothing in `renderer.inline`; verify the tests of 2.1 pass

## 3. Definition list

- [x] 3.1 Add `TestFootnotes` to `internal/render/render_test.go`, comparing the full plain output at width 40 for the scenarios of the spec: the bug case (`Text[1].`, blank, a 40-column line, blank, `1. Note.`), numbering by first reference (`1. Beta.` and `2. Alpha.` on consecutive lines), a definition between two paragraphs (rendered after `Last.`), an unreferenced definition (`Unused.` absent), no referenced definition (`Text.` only, no line), a definition with two paragraphs (`1. First paragraph.`, blank, `   Second paragraph.`) and a long definition whose wrapped lines start at column 3; verify `go test ./internal/render` fails
- [x] 3.2 Move the thematic break line into a helper and split `renderer.list` into `list` (markers of an `ast.List`) and `items(parent, markers, tight, width)` (the layout); verify `go test ./internal/render ./internal/site` passes with unchanged golden files and the tests of 3.1 still fail
- [x] 3.3 Add an `east.FootnoteList` case to `renderer.block`: the line, a blank line and `items` with the markers `N.` from `Footnote.Index`, tight when every definition has at most one block besides backlinks; verify the tests of 3.1 pass
- [x] 3.4 Add `TestFootnotesWithFrontmatter`: `---`, `title: Doc`, `---`, `Text[^1].`, `[^1]: Note.` starts with the frontmatter row `title │ Doc`, contains `Text[1].` and ends with `1. Note.`; verify it passes

## 4. Golden fixture

- [x] 4.1 Add `testdata/footnotes.md`: a numeric and a named label, a footnote referenced twice, references out of label order, a reference in bold text and in a table cell, a definition in the middle of the document, a long definition, a definition with two paragraphs and a code block, an unreferenced definition and an undefined reference; run `make golden`; verify `git status testdata/golden` shows only the six new `footnotes.*.golden` files, and in them the references are `[N]`, the definitions follow the line in number order, the unreferenced one is absent, wrapped and further lines are indented to the item text at width 40, and the styled files have no OSC 8 sequence around a reference

## 5. Documentation

- [x] 5.1 In `README.md` add a "Footnotes" subsection to "How it works" (references as `[1]` numbered by first reference, definitions at the end after a horizontal line as a numbered list, unreferenced definitions not shown, the reference is not a link, no back-references); verify every statement matches `specs/markdown-rendering/spec.md`

## 6. Verification

- [x] 6.1 Run `make check` and `openspec validate add-footnotes --strict`; verify both pass
- [x] 6.2 Run `make build`, then `./termd --no-pager --width 60 testdata/footnotes.md` and the same with `--hyperlinks=always`; verify the references show `[N]` and are not links, and the definitions are at the end after the line
