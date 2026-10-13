# Tasks

## 1. Images without alt text

- [x] 1.1 Add `TestImageWithoutAlt` to `internal/render/render_test.go`: in plain mode `![](logo.png)` gives `[image] (logo.png)`, `![ ](icon.png)` gives `[image] (icon.png)` and `![architecture](docs/arch.png)` still gives `[image: architecture] (docs/arch.png)`; verify `go test ./internal/render -run TestImageWithoutAlt` fails
- [x] 1.2 Move the `[image: alt]` label of `ast.Image` in `internal/render/render.go` into a helper that gives `[image]` for an alt text that is empty or only whitespace; verify the test of 1.1 passes and `go test ./internal/render ./internal/site` passes with unchanged golden files

## 2. Tokenizer

- [x] 2.1 Add `golang.org/x/net` to `go.mod` and `internal/render/html.go` with the tokenizing of an HTML fragment as in design.md: tokens of `html.NewTokenizer` with text and attribute values passed through `controlPicture`, and the rest that the tokenizer dropped returned as text; add `TestHTMLTokens`: `<a href="x&amp;y">A &lt;b&gt; &#27;</a>` gives a start tag `a` with `href` `x&y`, the text `A <b> ␛` and an end tag `a`; `ok <b>x</b> <broken <p` ends with the text `<broken <p`; `a < b` is the text `a < b`; verify `go test ./internal/render -run TestHTMLTokens` passes and `make tidy-check` passes

## 3. Inline HTML

- [x] 3.1 Add `HTMLSource bool` to `render.Options` and `TestInlineHTML` to `internal/render/render_test.go`, comparing the full plain output for the inline scenarios of the spec: line break (`one`, `two`), link in inline HTML (`See the guide (https://example.com).`), inline formatting tags (`Press Ctrl+C, bold and italic`), removed tags (`x2 text`), unknown tag in text (`Connect with ssh user@.`), unclosed and stray tags (`bold`, blank, `next text`), a line break in a table cell (the cell shows `one` and `two` on two lines and the column is 3 columns wide), and with `HTMLSource` the paragraph `Press <kbd>Ctrl</kbd>.` unchanged; verify `go test ./internal/render -run TestInlineHTML` fails
- [x] 3.2 Implement the open tags of inline HTML in `renderer.inlines` and `renderer.alert` as in design.md, keeping the current `RawHTML` path when `HTMLSource` is set; verify the test of 3.1 passes
- [x] 3.3 Add `TestInlineHTMLStyles`: in styled mode `Ctrl` of `<kbd>Ctrl</kbd>` has the inline code style of the palette, `<b>bold</b>` is bold, `<em>italic</em>` is italic, in `<b>bold`, blank, `next</b> text` only `bold` is bold; with hyperlinks `<a href="https://example.com">docs</a>` is an OSC 8 link to `https://example.com`, and `<a href="https://example.com/&#27;]0;pwned&#7;">docs</a>` without styling gives exactly `\033]8;;https://example.com/␛]0;pwned␇\033\docs\033]8;;\033\`; verify it passes

## 4. HTML blocks

- [x] 4.1 Add `TestHTMLBlocks` to `internal/render/render_test.go`, comparing the full plain output for the block scenarios of the spec: centered logo, image without alt, linked image in a pipe, header of a README (`[image: logo] (logo.png)`, blank, `# termd`, blank, `Markdown in the terminal`), script, style and comments (`A`, blank, `B`), character references, markdown in an HTML block, broken HTML and `<span>\033[31m</span>` (`␛[31m`); with hyperlinks enabled the row of badges is `[image: CI] [image]` on one line with the links of the `href` values, and an `<img src="img/arch.png" alt="arch">` of a document with `Dir` set links to the `file:` URL of `img/arch.png`, as in `TestRelativeLinks`; with `HTMLSource` every block is its source, as before; verify `go test ./internal/render -run TestHTMLBlocks` fails
- [x] 4.2 Render an `ast.HTMLBlock` as parts as in design.md: paragraphs and headings wrapped at the width, joined with blank lines, a block without parts gives no block; verify the test of 4.1 passes

## 5. Details

- [x] 5.1 Add `TestDetails` to `internal/render/render_test.go`, comparing the full plain output at width 40 for the `details` scenarios of the spec: details around markdown blocks (`▾ Install`, `  Run this:`, blank, `  go install x`, blank, `After.`), details in one block (`▾ More`, `  text`), details without summary (`▾ Details`), nested details (`▾ A`, `  ▾ B`, `    text`), long content (every line starts with 2 spaces and no line is wider than 40), a `details` that is not closed inside a block quote ends with the quote, and with `HTMLSource` the blocks are their source; in styled mode `▾` has the marker style of the palette and `More` has no style; verify `go test ./internal/render -run TestDetails` fails
- [x] 5.2 Group `details` in `renderer.blocks` and lay out a `details` as in design.md, without grouping when `HTMLSource` is set; verify the tests of 5.1 and `go test ./internal/render` pass, except the golden files of `code`, `inline` and `control`

## 6. Flag

- [x] 6.1 Add tests to `cmd/termd/main_test.go`: without the flag a document with `Press <kbd>Ctrl</kbd>.` renders to a pipe as `Press Ctrl.`, with `--html=source` as `Press <kbd>Ctrl</kbd>.`, and `--html=strip` is a usage error that names `--html` and lists `render` and `source`, with exit status 2, as in `TestInvalidFlags`; verify `go test ./cmd/termd` fails
- [x] 6.2 Add `--html` to `cmd/termd/main.go` with the default `render`, the validation and its usage text, and pass `HTMLSource` to `render.Options`; verify the tests of 6.1 pass

## 7. Golden fixture

- [x] 7.1 Add `testdata/html.md` with the header patterns of real READMEs: a centered logo with `alt`, a row of badges with and without `alt` inside `<a>`, `<h1 align="center">`, `<br>`, a `details` around a paragraph and a code block, a nested `details`, `<kbd>`, `<a href>` around markdown emphasis, `<picture>` with `<source>`, an HTML comment, a table cell with `<br>`, `<sup>` and a placeholder such as `<host>`; run `make golden`; verify `git status testdata/golden` shows the six new `html.*.golden` files and changed `code.*`, `inline.*` and `control.*` golden files only, and in them no tag of the table is shown as source, images without `alt` read `[image]`, the content of every `details` starts 2 columns right of its marker at every width, and the plain files have no escape sequences

## 8. Documentation

- [x] 8.1 In `README.md` add `--html` to the flags table, a "HTML" subsection to "How it works" (the tags and their output, `[image]` without alt text, the layout of `details`, removed tags including placeholders such as `<host>`, `--html=source`), `[image]` to the "Links" section, and the tags without layout (`hr`, `pre`, `ul`/`li`, `table`) to "Limitations"; verify every statement matches `specs/markdown-rendering/spec.md` and `specs/cli/spec.md`

## 9. Verification

- [x] 9.1 Run `make check` and `openspec validate render-html-tags --strict`; verify both pass
- [x] 9.2 Run `make build`, then `./termd --no-pager --width 60 testdata/html.md` in a terminal, with `| cat` and with `--html=source`; verify the tags are interpreted in the first two, the plain output has no escape sequences, and `--html=source` shows the HTML as source
- [x] 9.3 Render the READMEs of `github.com/gliderlabs/ssh` and `go.mongodb.org/mongo-driver/v2` from the module cache with and without `--html=source`; verify the first screen shows text and images instead of markup and the sponsors of `gliderlabs/ssh` read as a row of `[image]`
