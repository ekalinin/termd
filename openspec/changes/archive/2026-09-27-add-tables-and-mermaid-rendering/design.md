# Design

## Context

The project is empty apart from OpenSpec scaffolding. See proposal.md - Why for the motivation and specs/ for the required behavior.

Constraints observed on the target machine: Go 1.27 is installed (Node and Python are available, Rust is not), the terminal is iTerm2 3.5 with truecolor, `less` is version 581.2 and `PAGER` is not set. The primary environment is a local modern terminal; tmux and screen are used sometimes and SSH rarely.

A survey of the user's own markdown files (53 files with diagrams) shaped the diagram scope: mermaid `sequenceDiagram` 37 blocks, `flowchart`/`graph` 42, `stateDiagram-v2` 10, `classDiagram` 3, `gantt` and `erDiagram` 1 each, PlantUML 9 (7 of them sequence). In 29 sequence blocks without frontmatter, `note` appears in 16, `loop` in 13, activation in 5, `autonumber` in 4, `alt`/`else` in 3, `opt`, `par` and `box` in 1 each.

## Goals / Non-Goals

**Goals:**
- Alignment is correct by construction: every width, wrap and pad goes through one display-width function.
- Output is deterministic for a given input, width and mode, so every capability can be covered by golden-file tests.
- A single static binary with no runtime dependencies other than an optional `less`.

**Non-Goals:**
- Custom color themes and configuration files.
- Inline images and a built-in TUI.
- mermaid `stateDiagram` and `mindmap` (next change). The spike showed that rendering state through a flowchart translation merges opposite transitions (Loading <-> Error) into one line with swapped labels; that change has to decide between an upstream fix and a workaround.
- PlantUML rendering (change after next).

## Decisions

### Go as the implementation language
The best text renderer for the user's main diagram type is a Go library (see next decision), Go is already installed and produces a single binary.
Alternatives: TypeScript with beautiful-mermaid (needs a Node runtime, and the library lost the spike), Rust with merman-ascii (toolchain not installed, not evaluated).

### AlexanderGrooff/mermaid-ascii as the diagram renderer
A spike rendered the same synthetic samples (built from the constructs the user actually uses) with three libraries:

| | Grooff (Go) | pgavlin fork (Go) | beautiful-mermaid (TS) |
|---|---|---|---|
| sequence: note, loop, alt, activation, autonumber | all correct | truncates labels, lifeline shifts on activation | no autonumber or activation, a note after `end` lands inside `alt` |
| flowchart with subgraph | best | broken diamond | stray glyph at diamond |
| state | unsupported | silently wrong (drops transitions, invents one) | correct |
| ER | correct with cardinality markers | text list only | loses an entity and a relationship |
| mindmap | unsupported | tree, raw `root((...))` | unsupported |

A follow-up run confirmed Grooff also handles `opt`, `par`/`and`, `critical`, `break`, `rect`, `box`, `note left of`, self-messages and the `+`/`-` activation shorthand.

Grooff is MIT-licensed, actively developed (last commit 2026-09-08) and exposes public packages: `pkg/render.RenderDiagramWithStatus` (returns the output plus a `WidthStatus` with `Limit`, `Width` and `Met`), `pkg/sequence.Render(*SequenceDiagram, *diagram.Config)` which takes an exported model, and `pkg/diagram.Config` with `MaxWidth` and `GraphDirection`. It also strips frontmatter and prints its title.

termd uses the library API in-process, not the CLI as a subprocess. The module has no tagged releases, so it is pinned by pseudo-version `v0.0.0-20260908213847-5f00e3d9ac9f`. All calls go through one adapter package so the library can be forked or replaced without touching the rest of the renderer. The exported sequence model is the planned entry point for PlantUML sequence support later.

### termd detects the diagram type itself
Grooff's `DiagramFactory` treats any input that is not sequence or ER as a flowchart, so `stateDiagram-v2` or `mindmap` reach the flowchart parser and fail with a misleading error. termd reads the type keyword itself (skipping blank lines, `%%` comments and frontmatter) and passes only `sequenceDiagram`, `flowchart`, `graph` and `erDiagram` to the library. Everything else goes straight to the framed-source fallback with the reason "not supported".

Every library call runs under `recover()`: `pkg/graph/direction.go` panics on an unknown direction, and a panic in one diagram must not abort the document.

The frame label reads `<language> - <type> - <reason>`. Library errors start with a prefix such as `failed to parse sequence diagram: `, which repeats the type already in the label, so it is shortened to `parse error: ` (likewise `render error: `, `detect error: `). This keeps the label, and with it the frame, narrow; the rest of the message (line number and cause) is kept.

### Flowchart width fitting
`MaxWidth` only applies to flowcharts. The adapter renders a flowchart with `MaxWidth` set to the output width. If `WidthStatus.Met` is false and the direction is `LR` or `RL`, it rewrites the direction keyword in the header line to `TD` and renders again. If that still does not fit, the narrower of the two outputs is emitted as a wide block. Sequence and ER diagrams are emitted as rendered; if they are wider than the output they become wide blocks. The spike confirmed that Grooff compacts spacing under `--max-width` and reports when it cannot fit (a 50-column flowchart stayed 50 columns at a 40-column limit).

### Markdown parsing with goldmark, own renderer
goldmark with its GFM extensions (Table, TaskList, Strikethrough, Linkify) parses the document; termd walks the AST with its own renderer.
Alternatives: glamour (the renderer glow uses) - rejected because its table layout and width handling are the defects this change fixes, and termd needs full control of both; gomarkdown/markdown - less active and less CommonMark-compliant than goldmark.

### Block rendering model
Each block renders to a list of lines plus its natural width. Lines are built from spans (plain text plus a style and an optional link), so the width of a line is the sum of its spans' plain-text widths and escape sequences never enter the measurement. The document is the concatenation of blocks with blank-line separation. Code blocks, overflowing tables and diagrams are marked as wide blocks and bypass wrapping. Styles are emitted by a small internal package that writes SGR and OSC 8 sequences; in plain mode it writes nothing. Adjacent spans with the same style and link share one escape sequence, which keeps the output compact without changing what is shown.

Styles beyond the ones the specs name use the basic ANSI palette, so they follow the terminal's own colors on any background: link text is underlined and blue (`4;34`), inline code is cyan (`36`), list bullets, quote markers and thematic breaks are faint (`2`).
Alternative: lipgloss - rejected to keep width measurement in one place and avoid a second layout model.

### Syntax highlighting with chroma
chroma v2 (v2.27.0) is used only to tokenize code. The lexer is looked up by the first word of the info string through chroma's lexer names and aliases; there is no content-based guessing, because guessing is unreliable and produces wrong colors. Each token becomes a span whose style carries the foreground color that the selected chroma style assigns to the token type: `github-dark` for the dark theme and `github` for the light theme. The spans go through the same style package as the rest of the document, so width measurement is unchanged. The style package writes `38;2;R;G;B` when `COLORTERM` is `truecolor` or `24bit` and the nearest 256-color index (`38;5;N`) otherwise.
Alternatives: tree-sitter (needs cgo and per-language grammars); the 16 ANSI colors of the terminal palette (adapts to any background without detection, but a poorer palette - truecolor with detection was chosen); a fixed dark truecolor theme (unreadable on a light background).

### Background color detection
With `--theme=auto`, styling enabled and at least one code block to highlight, termd opens `/dev/tty` (so it works when the document comes from stdin), switches it to raw mode with `golang.org/x/term`, writes the OSC 11 background query (`ESC ]11;? ESC \`) followed by the DA1 query (`ESC [c`), and reads until the DA1 reply arrives or 100 ms pass. An OSC 11 reply (`rgb:RRRR/GGGG/BBBB`) before the DA1 reply gives the background; relative luminance of 0.5 or more selects the light theme. A DA1 reply alone, or a timeout, selects the dark theme. Every terminal answers DA1, so terminals that ignore OSC 11 do not cost the full timeout and late replies do not leak into the shell. The query runs before any output and before the pager starts.

The terminal is opened with `unix.Open` and read with `unix.Select` from `golang.org/x/sys/unix`, which becomes a direct dependency: on macOS `poll` and kqueue do not work on terminal devices, so Go's deadline-based reads cannot bound the wait, while `select` can. Detection is built for darwin, linux and the BSDs; on other platforms it reports no answer and the dark theme is used.
Alternatives: the `COLORFGBG` environment variable (set only by some terminals and not updated when the profile changes); always dark (the user chose detection).

### Display width with rivo/uniseg
`uniseg` measures by grapheme cluster and handles ZWJ sequences, variation selector 16 and East Asian width, which is what the markdown-rendering spec requires. Wrapping splits at whitespace and measures each word with it.
Alternative: mattn/go-runewidth - rune-based, which is the likely source of the ZWJ misalignment seen in glow.

### Table layout algorithm
For each column: `min` = widest word (header included), `natural` = widest cell on one line. `avail` = output width minus separators (3 columns per separator, ` │ `).
1. If the sum of `natural` fits `avail`, use `natural` for every column.
2. Else if the sum of `min` fits: start every column at `min`; visit columns in ascending order of `natural - min` and give each its full `natural` width while the remaining budget allows; split whatever budget remains among the columns that still wrap, proportionally to `natural - current`.
3. Else: every column gets `min` and the table is a wide block.

Cells wrap at whitespace only and are padded according to the column alignment on every line. Separators are `│` between columns and a `─┼─` rule under the header, without an outer border.
For the parameters example in the spec, step 2 gives `--theme`, `string`, `auto` and `По умолчанию` their natural width and leaves the rest to `Описание`, which is the only column that wraps.
Alternatives: purely proportional distribution (what produces `strin`/`g` in glow), truncation with an ellipsis, transposing overflowing tables into cards - the first two violate the spec, the third is not needed while horizontal scrolling in the pager covers overflow.

### Hyperlinks as OSC 8
Links are written as `ESC ]8;;URL ESC \` + visible text + `ESC ]8;; ESC \`. iTerm2 opens them with Cmd+click, and `less` passes OSC 8 through with `-R` since version 566. In plain mode links become `text (url)`.

### Paging with less -RS
termd renders the whole document into a buffer first. If stdout is a terminal, the buffer has more lines than the terminal height or a line wider than the terminal, `--no-pager` is not set and `less` is in `PATH`, termd starts `less -RS` and writes the buffer to its stdin; otherwise it writes to stdout. Without the width check a short document with a wide table or diagram would be printed directly and the terminal would wrap its lines, breaking the frames. Line width is measured on the rendered text with escape sequences removed, counting a tab as advancing to the next multiple of 8 columns. `-R` passes styles and OSC 8, `-S` keeps wide blocks unwrapped so they scroll horizontally. `$PAGER` is not consulted because a pager without these two behaviors breaks wide blocks and links.

When `LESSCHARSET` is not set, termd starts `less` with `LESSCHARSET=utf-8`: the output is always UTF-8, and under a non-UTF-8 locale `less` would otherwise show box-drawing characters as `<E2><94>...`. While `less` runs, termd ignores Ctrl-C: the key belongs to `less`, and if termd exited on it the shell prompt would return while `less` still owned the terminal.

### CLI with the standard flag package
Flags: `--width`, `--no-pager`, `--hyperlinks`, `--theme`. The standard `flag` package accepts both `-flag` and `--flag` and is enough for four flags.
Alternative: cobra - an unnecessary dependency for this surface.

### Layout and module path
Module `github.com/ekalinin/termd`:

```
cmd/termd/          CLI, flag parsing, width and TTY detection, pager
internal/render/    goldmark AST walk, block rendering, wide blocks
internal/text/      display width, word wrapping
internal/table/     table layout
internal/diagram/   type detection, Grooff adapter, framed-source fallback
internal/style/     SGR and OSC 8 output, 24-bit and 256-color, plain mode
internal/highlight/ chroma adapter: info string -> lexer, tokens -> spans
internal/termbg/    background color query (OSC 11 + DA1)
internal/golden/    golden-file assertions with -update, imported by tests only
testdata/           markdown fixtures and golden outputs
```

### Testing with golden files
Each fixture in `testdata/` is rendered at fixed widths (for example 40, 60, 80) in plain and styled modes and compared with committed golden files; an `-update` flag regenerates them. Styled golden files use an explicit theme and color depth, so they never depend on the terminal running the tests; background detection is unit-tested against scripted terminal replies. Fixtures include the glow regression cases from the proposal (the parameters table, emoji and CJK cells, a long link in a cell) and one diagram per supported construct from the sequence constructs requirement.

The comparison and the `-update` flag live in `internal/golden`, shared by the render, diagram and highlight tests; only test code imports it, so the flag never reaches the binary. The flag is registered only in test binaries that import the package, so `-update` is passed per package (for example `go test ./internal/render -update`). To check that a panic in the diagram library leaves the rest of the document intact, `internal/diagram/export_test.go` exposes the library entry point as a test-only hook, which an external test package replaces with a panicking function and then renders a whole document.

## Risks / Trade-offs

- [Grooff has a single maintainer and no releases] → pinned pseudo-version, all use behind `internal/diagram`, fork if it stalls.
- [Grooff measures labels with go-runewidth, so emoji inside diagram labels can misalign] → accepted for v1; fix upstream if it shows up in real documents.
- [The `box` group frame overlaps message lines visually (seen in the spike)] → accepted, the diagram stays correct; report upstream.
- [Terminals disagree on the width of emoji with variation selector 16; iTerm2 depends on its "Unicode version 9+ widths" setting] → termd follows Unicode (width 2); golden tests pin this, and the setting is documented.
- [Rewriting `LR` to `TD` changes the author's orientation] → only applied when the horizontal layout does not fit.
- [In plain mode long URLs are unbreakable words and can push tables into overflow] → accepted; plain mode is mainly for pipes where horizontal space is not limited by a screen.
- [GNU screen does not support OSC 8] → `--hyperlinks=never`.
- [Grooff's go.mod lists web-server dependencies] → only imported packages are linked; check the binary size after the first build.
- [A chroma lexer can tokenize some code wrongly] → only colors are affected, the text is always the original source.
- [tmux, screen or some SSH setups may not answer OSC 11 or may report their own background] → the dark theme is the fallback and `--theme` overrides detection.
- [The 256-color fallback only approximates the theme colors] → accepted; it applies only to terminals without truecolor.
- [The background query adds latency] → it runs only with `--theme=auto` and a code block to highlight, and the DA1 reply bounds the wait on terminals that ignore OSC 11.

## Migration Plan

Not applicable: this is the first release of a new tool. It is installed with `go install github.com/ekalinin/termd/cmd/termd@latest`.

## Open Questions

- Whether to add a `TERMD_PAGER` override later for users who prefer another pager.
- A built-in termd pager, available as an option instead of `less`, in one of the next changes.
