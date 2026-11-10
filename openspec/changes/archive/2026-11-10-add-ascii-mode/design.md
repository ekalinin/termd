# Design

## Context

termd draws its own characters in three packages:

- `internal/render`: the bullet `•` in `list`, the quote marker `│` in `quote` (also for alerts), the line `─` in `rule` (thematic breaks and the line before the footnotes) and the `details` marker `▾` in `html.go`. `frontmatterBlock` draws an invalid block with `diagram.Frame` and a valid one with `table.Table`.
- `internal/table`: the constants `sep` (` │ `) and `ruleSep` (`─┼─`) and the `─` of the header rule.
- `internal/diagram`: `Frame` (`┌─┐│└┘`), the mindmap branches in `mindNode.layout` (`├── `, `└── `, `│   `), the gantt bars in `ganttChart.bar` (`█▓░◆`), the state symbols in `symbols` (`●◉◇━━━`), the region line `┈` in `stateDiagram.section` and the look-alikes `‑` and `›` in `breakLinks`. Sequence diagrams, ER diagrams, flowcharts, the flowcharts of state diagrams and PlantUML sequence diagrams are drawn by mermaid-ascii through `safeRender` with `mdiagram.DefaultConfig()`, whose `UseAscii` is false.

`diagram.Render(lang, src, width)` is called 39 times in `internal/diagram/diagram_test.go`, and `diagram.Frame` once there; `render.Render` calls `diagram.Render` and `frontmatterBlock` calls `diagram.Frame`.

The requirements are in `specs/cli/spec.md`, `specs/markdown-rendering/spec.md`, `specs/table-rendering/spec.md` and `specs/diagram-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- Without `--ascii`, the output is byte for byte the same as before: every existing golden file stays unchanged.
- In ASCII mode the layout is the same as in Unicode mode wherever the characters have the same width, so the width rules and wide blocks need no changes.

**Non-Goals:**
- Changing the text of the document or of the diagram source, see the proposal.
- Detecting from the locale or the terminal that the mode is needed.

## Decisions

### One boolean, characters in a struct per package

The mode is a boolean `ASCII` on `render.Options` and on `table.Table`. Each package keeps the characters it draws in a small struct with a Unicode value and an ASCII value, and picks one at its entry point, so the drawing code reads a field instead of testing the mode:

- `render`: `marks{bullet, quote, rule, details}`, with `unicodeMarks` (`•`, `│`, `─`, `▾`) and `asciiMarks` (`*`, `|`, `-`, `v`), picked by `Options.marks()` like `Options.palette()` where they are drawn; tests that build a `renderer` by hand get them from its options without setting another field. `rule` takes the character of the line as a parameter.
- `table`: `Render` picks the separator ` │ ` or ` | `, the rule separator `─┼─` or `-+-` and the rule character `─` or `-` from `Table.ASCII`; `sepWidth` does not change.
- `diagram`: `glyphs`, with `unicodeGlyphs` and `asciiGlyphs`, holding the frame, the mindmap branches, the gantt bars, the state symbols, the region line and `ascii`, which sets `UseAscii` and the behavior of `breakLinks`.

Alternative: a package-level mode variable. Rejected: it is global state shared by parallel tests and by documents rendered concurrently.

Alternative: an `ASCII bool` parameter tested at every drawing site. Rejected: the characters of one element would be spread over `if`s in many functions, and the table of characters in the spec would have no single place in the code.

### `RenderASCII` and `FrameASCII` next to `Render` and `Frame`

`diagram.Render(lang, src, width)` and `diagram.Frame(label, src)` keep their signatures and draw with `unicodeGlyphs`. The new `RenderASCII(lang, src, width)` and `FrameASCII(label, src)` draw with `asciiGlyphs`. All four call unexported `render(lang, src, width, g)` and `frame(label, src, g)`. `render.Render` calls the ASCII variants when `Options.ASCII` is set, and `frontmatterBlock` gets the mode as a parameter.

Alternative: an `ascii bool` parameter on `Render` and `Frame`. Rejected for the size of the diff: every call in `diagram_test.go` would change without a change in behavior.

### The glyphs travel with the diagram

Inside `internal/diagram`, `g *glyphs` is passed down the call chains that draw: `renderFlowchart(src, width, g)`, which builds the config with `UseAscii: g.ascii`; `safeRender` keeps taking a config; `renderPlantUML`, `renderMindmap`, `drawTree` and `mindNode.layout`; `renderGantt`, `ganttChart.layout` and `ganttChart.bar`; `renderState`, `stateDiagram.draw`, `section`, `flowchart` and `nodeLines`. Parsing does not change.

### State diagrams in ASCII mode

The symbols become node texts for mermaid-ascii: `*`, `(*)`, `<>` and `===`. A test with mermaid-ascii `5a9df1560871` showed that none of them is read as flowchart syntax inside `["..."]`. `(*)` is 2 columns wider than `◉`, so the box of the end state is wider; that is a different diagram, not a different layout rule.

`breakLinks` gets the mode. In Unicode mode it replaces the characters as today. In ASCII mode it keeps the same characters, a `-` after a `-` and a `>` after a `-` or a `=`, and puts a space before each of them: `a --> b` becomes `a - - > b`, `x ==> y` becomes `x == > y`. With the same mermaid-ascii version, `-->`, `==>` and `---` inside a label break the node, while `- - >`, `== >` and `- - -` are kept as text.

Alternative: other ASCII look-alikes, such as `~` for `-`. Rejected: they change what the text says; a space only changes the spacing.

### mermaid-ascii in ASCII mode keeps the geometry

With `UseAscii`, mermaid-ascii draws the fixtures `flow-lr`, `flow-td`, `er`, `seq-activation`, `seq-blocks`, `seq-notes` and `title` with the same number of lines and the same width of every line as without it, also with `MaxWidth` 40. So `renderFlowchart` fits a flowchart the same way, and `Result.Wide` is the same.

### Tests and golden files

- `table`: a test for the scenario of the spec and a check that the separators are at the same columns as in Unicode mode.
- `diagram`: tests for every ASCII scenario of the spec through `RenderASCII` and `FrameASCII`; the flowchart and sequence tests compare the line count and the widths with `Render`.
- `render`: tests for every ASCII scenario of markdown-rendering, including styles.
- Golden files: `TestGolden` gets a third mode `ascii`, plain output in ASCII mode, at width 80 only, written to `testdata/golden/<name>.w80.ascii.golden`; the mapping of characters does not depend on the width, and the existing modes cover the widths. The existing golden files do not change.
- A check over every fixture: in ASCII mode, at widths 40, 60 and 80, plain and styled, the output without escape sequences contains no character of the Box Drawing, Block Elements, Geometric Shapes and Arrows blocks and no `•`, `‑` or `›`, unless the fixture source contains it.
- `cmd/termd`: the flag in a pipe, with styles and hyperlinks, with `LANG=C` and without the flag.

## Risks / Trade-offs

- [A later mermaid-ascii version draws differently in ASCII mode, or reads the ASCII symbols as syntax] → The ASCII golden files and the state tests fail on the update.
- [The done bars of a gantt chart (`-`) look like a line] → They stand in the bar columns next to task names, and the chart has no other lines; accepted.
- [ASCII mode does not make the output pure ASCII: the text of the document can be Unicode] → Stated in the spec and the README; the mode is about the characters termd adds.
- [A document that uses `*` or `|` in its text reads less clearly next to `*` bullets and `|` quote markers] → The same trade-off as glamour's `ascii` style.

## Migration Plan

None. Without the flag the output does not change. Rollback is a revert of the change.
