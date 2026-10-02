# Tasks

## 1. Detection and title

- [x] 1.1 Add `TestAlerts` to `internal/render/render_test.go`, rendered in plain mode, with one case per scenario and design rule: a note (`│ Note`, `│ Useful info.`), a lower-case `[!warning]`, a hard line break after the marker, a marker-only quote (`│ Tip`), several blocks (`│ Important`, `│ First.`, `│`, `│ Second.`), a nested quote (`│ │ Quoted.`), an alert in a list item (`  │ Tip`), and the regular quotes: text after the marker, `[!FOO]`, an escaped `\[!NOTE]` and `[!tıp]` with a dotless i; add a case to `TestBlockQuoteMarker` that a note alert longer than width 30 keeps `│ ` on every line within 30 columns; verify `go test ./internal/render` fails
- [x] 1.2 In `internal/render/render.go` add the alert type list and the `alert` helper (raw text nodes up to the first line break, trimmed, compared with `strings.EqualFold`), and in `renderer.quote` replace the block of the first paragraph with the wrap of the title, a `text.Break` and the rest of the paragraph; verify the tests of 1.1 pass and `go test ./internal/render` passes with unchanged golden files

## 2. Colors

- [x] 2.1 Add `TestAlertStyles` to `internal/render/render_test.go`: in styled mode the five types have their markers in SGR `34`, `32`, `35`, `33` and `31` without the faint attribute on every line, and their titles bold in the same color (`\x1b[1;34mNote` for a note); a regular quote nested in an alert keeps its faint marker; in plain mode the output of the five types has no escape sequences; verify `go test ./internal/render` fails
- [x] 2.2 Style the alert marker with the color of the alert type and the title with bold and that color; verify `go test ./internal/render` passes

## 3. Golden fixture

- [x] 3.1 Add `testdata/alerts.md`: the five types, a lower-case marker, a marker-only quote, an alert with two paragraphs and a nested quote, an alert in a list item, and the regular quotes with text after the marker and with `[!FOO]`; run `make golden`; verify `git status testdata/golden` shows only the six new `alerts.*.golden` files, the plain files show the titles after `│ `, and the styled files have colored markers and bold colored titles

## 4. Documentation

- [x] 4.1 In `README.md` add an "Alerts" subsection to "How it works" (the five markers in any case, the titles, the colors, plain text in a pipe, other markers or text after the marker leave a regular quote, no icons); verify every statement matches `specs/markdown-rendering/spec.md`

## 5. Verification

- [x] 5.1 Run `make check` and `openspec validate add-github-alerts --strict`; verify both pass
- [x] 5.2 Run `make build`, then `./termd --no-pager testdata/alerts.md` through a pseudo-terminal (`script -q /dev/null ...`) and into a pipe; verify the terminal output has the colored markers and bold titles and the piped output is plain text
