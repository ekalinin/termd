# Tasks

## 1. Tables

- [x] 1.1 Add `TestASCII` to `internal/table/table_test.go`: the header `A`, `B` and the row `1`, `2` with `ASCII: true` give the lines `A | B`, `--+--` and `1 | 2`; a table with wrapped cells and a `BorderStyle` has its separators at the same columns as without `ASCII` and the same style on `|`, `-` and `+`; verify `go test ./internal/table` fails
- [x] 1.2 Add `ASCII bool` to `table.Table` and pick the separator, the rule separator and the rule character from it in `Render` and `row`; verify the tests of 1.1 pass and the other tests of `internal/table` still pass

## 2. Diagrams

- [x] 2.1 Add tests to `internal/diagram/diagram_test.go` for the "ASCII mode" scenarios of `specs/diagram-rendering/spec.md`: the flowchart `flowchart TD`, `A --> B` and the sequence diagram `sequenceDiagram`, `Alice->>Bob: Hi` through `RenderASCII` are all ASCII and have the line count and line widths of `Render`; the choice, fork and join diagram equals the ASCII flowchart with the nodes `*`, `<>`, `===`, `===` and `(*)`; the concurrent regions are separated by a blank line, a line of `.` and a blank line; `state "a --> b" as T` shows `a - - > b`; the mindmap `Plan` tree gives `|-- `, `|   ` and a backtick and `-- `; the gantt task states give `-`, `=`, `#` and a single `*`; `FrameASCII` of the `classDiagram` block gives the 4 lines of the spec; `A[Café] --> B` keeps `Café`; the PlantUML sequence, mindmap and state fixtures through `RenderASCII` contain no non-ASCII character that their source does not contain; verify `go test ./internal/diagram` fails
- [x] 2.2 Add `glyphs` with `unicodeGlyphs` and `asciiGlyphs` to `internal/diagram`, move `Render` and `Frame` onto unexported `render` and `frame` that take the glyphs, and add `RenderASCII` and `FrameASCII`; set `UseAscii` from the glyphs in every mermaid-ascii config (`renderFlowchart`, `sequenceDiagram` and `erDiagram` in `render`, and `renderPlantUML`); verify the frame, flowchart and sequence tests of 2.1 pass and `go test ./internal/diagram` passes for every test that existed before
- [x] 2.3 Pass the glyphs through `renderMindmap`, `drawTree` and `mindNode.layout`, through `renderGantt`, `ganttChart.layout` and `ganttChart.bar`, and through `renderState`, `stateDiagram.draw`, `section`, `flowchart` and `nodeLines`, including PlantUML mindmaps and state diagrams; make `breakLinks` put a space before the characters in ASCII mode; verify every test of 2.1 passes and `git status internal/diagram/testdata` shows no changes

## 3. Rendering

- [x] 3.1 Add `TestASCIIMode` to `internal/render/render_test.go` for the scenarios of `specs/markdown-rendering/spec.md` and `specs/table-rendering/spec.md`: lists and quotes, the alert, the thematic break and footnotes at width 20, `details`, the styles of `|` and `*` with styling enabled, the invalid YAML frontmatter frame, the frontmatter table `title | Doc`, the table `A | B`, `--+--`, `1 | 2`, a mermaid flowchart drawn as `diagram.RenderASCII` draws it, and the text `Café :tada: &copy;` with ESC kept as `Café 🎉 ©` and `␛`; verify `go test ./internal/render` fails
- [x] 3.2 Add `ASCII bool` to `render.Options`, add `marks` with `unicodeMarks` and `asciiMarks` picked by `Options.marks()`, and use them for the bullet in `list`, the quote marker in `quote`, the line in `rule` and its callers, and the `details` marker in `html.go`; set `table.Table.ASCII` in `table` and in `frontmatterBlock`, which gets the mode as a parameter, and call `diagram.RenderASCII` and `diagram.FrameASCII` in ASCII mode; verify the tests of 3.1 pass, `go test ./internal/render ./internal/site` passes and `git status testdata/golden` shows no changes

## 4. Golden files

- [x] 4.1 Add the mode `ascii` to `TestGolden` in `internal/render/golden_test.go`: every fixture rendered with `goldenOptions(80, false)` and `ASCII` set, compared with `testdata/golden/<name>.w80.ascii.golden`; run `go test ./internal/render -run TestGolden -update`, then verify `git status testdata/golden` shows only new `*.w80.ascii.golden` files, and read the ASCII golden files of `lists`, `quotes`, `alerts`, `tables`, `footnotes`, `frontmatter`, `html`, `diagrams`, `mindmap`, `gantt`, `state` and `plantuml` to check the characters of the spec
- [x] 4.2 Add `TestASCIIFixtures` to `internal/render/golden_test.go`: every fixture in ASCII mode at widths 40, 60 and 80, plain and styled, with escape sequences removed, contains no character of the blocks Box Drawing (U+2500 to U+257F), Block Elements (U+2580 to U+259F), Geometric Shapes (U+25A0 to U+25FF) and Arrows (U+2190 to U+21FF) and no `•`, `‑` or `›`, unless the source of the fixture contains it; verify it passes, and that it fails when one entry of `asciiMarks` or `asciiGlyphs` is set back to its Unicode character

## 5. Command line

- [x] 5.1 Add `TestASCIIFlag` to `cmd/termd/main_test.go`: `- item` and `> quoted` from stdin with `--ascii` give `* item` and `| quoted`; with `--ascii --hyperlinks=always` in a styled terminal, `> [docs](https://example.com)` gives the marker `|` with the marker style and an OSC 8 link to `https://example.com`; with `LANG=C` in the environment and without the flag, `- item` gives `• item`; verify `go test ./cmd/termd` fails
- [x] 5.2 In `run` add the flag `ascii` with the usage text "draw tables, lines, markers and diagrams with ASCII characters only" and pass it as `Options.ASCII`; verify `go test ./cmd/termd` passes and `go run ./cmd/termd -h` lists `-ascii`

## 6. Documentation

- [x] 6.1 In `README.md` add an `--ascii` row to the flags table and a section "ASCII mode" under "How it works" with the table of characters from the proposal, the space before `-` and `>` in state texts, that styles, themes and hyperlinks do not change, that the text of the document is not changed, and that termd does not turn the mode on by itself; verify every statement matches the specs of the change

## 7. Verification

- [x] 7.1 Run `make check` and `openspec validate add-ascii-mode --strict`; verify both pass and `git status testdata/golden` shows only the new `*.w80.ascii.golden` files
- [x] 7.2 Run `make build`, then `./termd --no-pager --ascii testdata/diagrams.md`, `testdata/state.md`, `testdata/gantt.md`, `testdata/mindmap.md`, `testdata/tables.md` and `README.md` through a pseudo-terminal and in a pipe; verify the drawn characters are ASCII, the styled output keeps its colors and hyperlinks, and the piped output has no escape sequences
