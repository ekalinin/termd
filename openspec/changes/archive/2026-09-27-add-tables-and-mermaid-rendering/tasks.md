# Tasks

## 1. Project setup

- [x] 1.1 Initialize Go module `github.com/ekalinin/termd` with `cmd/termd` and packages `internal/render`, `internal/text`, `internal/table`, `internal/diagram`, `internal/style`, `internal/highlight`, `internal/termbg`; verify `go build ./...` succeeds
- [x] 1.2 Add dependencies `github.com/yuin/goldmark`, `github.com/rivo/uniseg`, `github.com/alecthomas/chroma/v2`, `golang.org/x/term` and `github.com/AlexanderGrooff/mermaid-ascii` pinned at `v0.0.0-20260908213847-5f00e3d9ac9f`; verify `go.mod` lists them and `go build ./...` succeeds
- [x] 1.3 Add a golden-file test helper that renders `testdata/*.md` at widths 40, 60 and 80 in plain and styled modes, compares with `*.golden` files and regenerates them with `-update`; verify a trivial fixture passes and `-update` rewrites its golden file

## 2. Display width and wrapping (`internal/text`)

- [x] 2.1 Implement display width by grapheme cluster; verify unit tests: `👨‍👩‍👧` = 2, `⚠️` = 2, `日本語` = 6, `кириллица` = 9, a letter with a combining mark = 1
- [x] 2.2 Implement word wrapping at whitespace that breaks only words longer than the width; verify tests that no line exceeds the width and that a 120-character word at width 80 is split into 80 + 40
- [x] 2.3 Implement the span model (text, style, optional link) with line width computed from span text only; verify a styled `bold` span line measures 4

## 3. Styles and hyperlinks (`internal/style`)

- [x] 3.1 Emit SGR sequences for bold, italic, strikethrough and inline code in styled mode and nothing in plain mode; verify a test that plain output contains no ESC byte
- [x] 3.2 Emit links as OSC 8 when hyperlinks are enabled and as `text (url)` (or the URL alone when text equals URL) when disabled; verify tests for the three scenarios of the markdown-rendering Hyperlinks requirement

## 4. Markdown rendering (`internal/render`)

- [x] 4.1 Parse with goldmark and the GFM extensions (Table, TaskList, Strikethrough, Linkify); verify a test that a fixture with a table, a task item, `~~text~~` and a bare URL yields the corresponding AST nodes
- [x] 4.2 Render paragraphs (wrapped) and headings (`#` markers kept, bold when styled, blank line before); verify golden tests at widths 40 and 80
- [x] 4.3 Render bold, italic, strikethrough and inline code styled or plain without markers; verify golden tests in both modes
- [x] 4.4 Render unordered, ordered, nested and task lists with continuation lines aligned to the item text; verify a golden test at width 40 with long nested items
- [x] 4.5 Render block quotes with the quote marker on every wrapped line; verify a golden test at width 40
- [x] 4.6 Render code blocks verbatim without wrapping and thematic breaks across the output width; verify golden tests including a code line wider than the width
- [x] 4.7 Render images as `[image: alt]` linked to the URL and raw HTML as source text; verify golden tests for both scenarios
- [x] 4.8 Mark code blocks, overflowing tables and diagrams as wide blocks that bypass wrapping and truncation; verify a test that a 120-column block at width 80 is output with every line intact
- [x] 4.9 Highlight fenced code blocks by info-string language with chroma, mapping tokens to styled spans with `github-dark` and `github` colors in 24-bit or 256-color depth; verify styled golden tests for `go`, `python` and `sh` blocks in both themes and depths, uncolored output for a block without info string and one with `foo`, and a test that stripping escapes from highlighted output yields the original source
- [x] 4.10 Implement background detection in `internal/termbg` (OSC 11 followed by DA1 on `/dev/tty` in raw mode, 100 ms timeout, luminance threshold 0.5), run only for `--theme=auto` with a code block to highlight; verify unit tests with scripted replies: white background -> light, black -> dark, DA1 only -> dark, no reply -> dark within 100 ms, and no query for `--theme=dark` or a document without highlightable code

## 5. Table rendering (`internal/table`)

- [x] 5.1 Compute per-column minimum (widest word, header included) and natural widths with the display-width function; verify unit tests including `По умолчанию` giving a minimum of 9
- [x] 5.2 Implement the three-step width allocation from design.md; verify unit tests for a table that fits naturally, the parameters table at width 60 (only `Описание` wraps) and a 12-column overflow table rendered at minimum widths
- [x] 5.3 Wrap cells at whitespace and apply left, center and right alignment to every line of a cell; verify tests for right-aligned `1`, `42`, `1000` and a centered column
- [x] 5.4 Draw tables with `│` separators and a `─┼─` header rule, filling missing cells and ignoring extra ones; verify a test that separators sit at identical display columns on every line, including rows with emoji and CJK text
- [x] 5.5 Render inline formatting, links and escaped pipes inside cells and wire tables into the renderer; verify golden tests for the glow regression fixture (parameters table, emoji and CJK table, long link in a cell) at widths 40, 60 and 80

## 6. Diagram rendering (`internal/diagram`)

- [x] 6.1 Detect diagram blocks by info string (`mermaid`, `plantuml`, `puml`, case-insensitive) and the mermaid type from the first line that is not blank, a `%%` comment or frontmatter; verify unit tests including frontmatter followed by `sequenceDiagram` and a `go` block staying a code block
- [x] 6.2 Implement the framed-source fallback with a label naming language, detected type and reason; verify golden tests for `stateDiagram-v2` (not supported), a `plantuml` block and a `sequenceDiagram` with a syntax error
- [x] 6.3 Implement the Grooff adapter for `sequenceDiagram`, `flowchart`, `graph` and `erDiagram`, running every call under `recover()`; verify a test that an injected panic produces framed source and the next block still renders
- [x] 6.4 Implement flowchart width fitting: render with `MaxWidth`, rewrite `LR`/`RL` to `TD` when `WidthStatus.Met` is false, emit the narrower result as a wide block; verify tests at widths where each branch is taken
- [x] 6.5 Add golden fixtures for every sequence construct in the spec (notes left/right/over, `loop`, `alt`/`else`, `opt`, `par`/`and`, `critical`, `break`, `rect`, `activate`/`deactivate` and `+`/`-`, `autonumber`, `box`, Cyrillic labels), flowcharts (subgraph, edge labels, diamond, `TD` and `LR`), ER cardinality and a frontmatter title; verify the golden outputs by eye once and `go test ./internal/diagram/...` passes
- [x] 6.6 Wire diagram rendering into the renderer so diagram blocks replace code blocks; verify a golden test of a document with a supported diagram, an unsupported diagram and a paragraph after them

## 7. CLI (`cmd/termd`)

- [x] 7.1 Parse `--width` (positive integer), `--no-pager`, `--hyperlinks=auto|always|never` and `--theme=auto|dark|light` with the standard `flag` package, exiting with status 2 on invalid values; verify tests for `--width 0`, `--width abc`, `--hyperlinks=sometimes` and `--theme=blue`
- [x] 7.2 Read input from the file argument, from `-`, or from stdin when it is not a terminal; print usage and exit 2 when there is no input, exit 1 with the file name on read errors; verify tests for each scenario of the Input source requirement
- [x] 7.3 Detect the output width (terminal columns, 80 for non-terminals, `--width` override) and select styled or plain mode by whether stdout is a terminal; verify a test that piped output contains no ESC byte and uses width 80
- [x] 7.4 Implement paging: buffer the output and start `less -RS` only when stdout is a terminal, the output is taller than the terminal, `--no-pager` is absent and `less` is in `PATH`; verify unit tests of the decision with terminal state, height and pager lookup injected
- [x] 7.5 Exit with status 0 when diagrams fell back to framed source; verify a test running the CLI on a document with a `stateDiagram-v2` block
- [x] 7.6 Also page output that fits the terminal height but has a line wider than the terminal (width measured without escape sequences, tabs to multiples of 8); verify unit tests of the decision for a short document with a wide table and with a tab-indented code line, and a CLI test that such a document goes to the pager

## 8. Integration verification

- [x] 8.1 Run `go vet ./...` and `go test ./...`; verify both pass
- [x] 8.2 Build the binary and render a fixture with the glow regression tables, a long sequence diagram and links in iTerm2; verify separators line up on screen, Cmd+click opens the full URL, `less -RS` scrolls the wide diagram horizontally, code blocks are highlighted and readable, and switching the iTerm2 profile to a light background selects the light theme
- [x] 8.3 Render the same fixture with glow 2.1.2 and termd at width 60; verify that none of the five defects from proposal.md (truncated header, split word, emoji misalignment, truncated URL, raw mermaid source) appear in termd output, and record the binary size to check the dependency risk from design.md
