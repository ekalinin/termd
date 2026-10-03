# termd

A terminal markdown viewer that renders tables and diagrams correctly.

Existing terminal viewers often break tables (truncated headers, words split in the middle, misaligned rows with emoji or CJK text, cut-off URLs) and show mermaid diagrams as raw source. termd lays out tables by the real display width of every character and draws mermaid diagrams as text.

## Features

- **Tables** - widths are measured by grapheme clusters, so emoji (including ZWJ sequences and variation selectors) and CJK text keep columns aligned. Words are never split, headers are never truncated, and free width goes to the columns with long text. Column alignment from the markdown source is kept. A table that does not fit even at its minimum widths is emitted in full and can be scrolled horizontally in the pager.
- **Mermaid diagrams as text** - `sequenceDiagram`, `flowchart`/`graph` and `erDiagram` are drawn with box-drawing characters. Any other diagram is shown as its source in a labelled frame instead of failing.
- **Frontmatter** - a YAML frontmatter block is shown as a table of its keys and values at the top of the document.
- **Syntax highlighting** - fenced code blocks with a known language are highlighted, with a dark or light theme chosen by the terminal background.
- **Color themes** - `dark` and `light` follow the color scheme of the terminal; ten named themes such as `dracula`, `nord` or `catppuccin-mocha` color headings, links, inline code, markers, alerts, tables and code with the palette of their scheme.
- **Clickable links** - links are OSC 8 terminal hyperlinks, so a table cell shows `docs` instead of a long URL.
- **Paging** - long or wide output opens in `less -RS`.
- **Pipe-friendly** - when stdout is not a terminal, the output is plain text without escape sequences.

## Install

With Go 1.27 or later:

```sh
go install github.com/ekalinin/termd/cmd/termd@latest
```

Prebuilt binaries for Linux, macOS and Windows on amd64 and arm64 are attached to every [release](https://github.com/ekalinin/termd/releases). Unpack the archive for your platform and put `termd` on your `PATH`; `checksums.txt` holds the SHA-256 of every archive. On macOS, a binary downloaded with a browser is quarantined by Gatekeeper; `xattr -d com.apple.quarantine termd` removes the attribute. The Windows binaries are best-effort: termd does not enable virtual terminal processing in the console and was not tested on Windows.

From source:

```sh
git clone https://github.com/ekalinin/termd.git
cd termd
make build    # produces ./termd
```

`less` is optional. It is used for paging; version 566 or later is needed for clickable links inside the pager. Without `less`, termd prints directly.

## Usage

```sh
termd README.md             # render a file
termd docs/                 # render the README of a directory
cat README.md | termd       # render standard input
termd - < README.md         # same, explicitly
termd --width 60 doc.md     # lay out for 60 columns
termd doc.md > doc.txt      # plain text, 80 columns
termd main.go               # show a source file as code
termd --version             # print the version
```

| Flag | Values | Default | Description |
|---|---|---|---|
| `--width` | positive number | terminal width, or 80 when not a terminal | output width in columns |
| `--no-pager` | | off | print directly instead of paging through `less` |
| `--hyperlinks` | `auto`, `always`, `never` | `auto` | terminal hyperlinks; `auto` enables them only when stdout is a terminal |
| `--theme` | `auto` or a theme name | `TERMD_THEME`, or `auto` when it is not set | color theme, see [Themes](#themes); `auto` asks the terminal for its background color |
| `--version` | | | print the version and exit |

Exit status: `0` when the document was rendered (even if some diagrams were shown as source), `1` when the file cannot be read, `2` on a usage error.

## Examples

A table at `--width 60`: short columns keep their full width, only the description wraps.

```text
Option  │ Type   │ Default │ Description
────────┼────────┼─────────┼────────────────────────────────
--width │ int    │ 80      │ Maximum output width in
        │        │         │ terminal columns, after which
        │        │         │ text wraps
--theme │ string │ auto    │ Color theme: auto or the name
        │        │         │ of a built-in theme
```

A mermaid sequence diagram with `autonumber`, activation and a note:

```text
┌────────┐     ┌─────┐
│ Client │     │ API │
└────┬───┘     └──┬──┘
     │            │
     │ 1. GET /doc│
     ├───────────►│
     │            ┃
     │            ┃ ┌─────────────┐
     │            ┃ │ check token │
     │            ┃ └─────────────┘
     │            ┃
     │ 2. 200 OK  ┃
     │◄┈┈┈┈┈┈┈┈┈┈┈┨
     │            │
```

A diagram type that is not supported yet:

```text
┌─ mermaid - stateDiagram-v2 - not supported ┐
│ stateDiagram-v2                            │
│     [*] --> Idle                           │
└────────────────────────────────────────────┘
```

## How it works

### Output width and modes

In a terminal, termd lays the document out for the terminal width and emits styles, colors and hyperlinks. When stdout is not a terminal, it uses 80 columns and emits plain text. `--width` overrides the width in both cases, and `--hyperlinks=always` forces hyperlinks into a pipe.

### Control characters

The control characters of a document reach the terminal as visible characters, so a document cannot change the window title, write the clipboard or redraw the screen. Every C0 control character except tab and line feed, and DEL, is shown as its Unicode control picture, one column wide: ESC as `␛`, BEL as `␇`, NUL as `␀`, DEL as `␡`. C1 control characters and bytes that are not valid UTF-8 are shown as `�`. This applies to all text of the document, including code blocks, link destinations, frontmatter and diagrams, and to control characters written as character references (`&#27;`) or YAML escapes (`"\e"`). CRLF and a lone CR are line endings, as in CommonMark. The styles and hyperlinks termd emits itself do not change.

### Paging

When stdout is a terminal and the output is taller than the screen, or has a line wider than the screen, termd shows it in `less -RS`: `-R` passes styles and hyperlinks through, `-S` keeps wide tables and diagrams unwrapped so they scroll horizontally with the arrow keys. `$PAGER` is not used, because a pager without these two behaviors breaks wide blocks and links.

### Links

With hyperlinks enabled, a link shows only its text and opens its destination when clicked (Cmd+click in iTerm2). With hyperlinks disabled, a link is written as `text (url)`, and images as `[image: alt] (url)`. URLs are never truncated.

When the document is read from a file, relative destinations of links and images, such as `docs/guide.md` or `../img/arch.png`, are resolved against the directory of the file and become `file://` URLs with the host name of the machine, for example `file://myhost/home/me/project/docs/guide.md`, so a click opens the target with the default application of the system. The fragment of a destination (`guide.md#setup`) is kept, its query (`logo.png?raw=true`) is dropped. Absolute URLs, `#section` links and destinations that start with `/` are left as written; GitHub reads `/docs/x.md` as relative to the repository root, which termd does not know. A document read from stdin has no directory, so its links are left as written too. With hyperlinks disabled, `text (url)` always shows the destination as written in the document.

### Footnotes

A footnote reference is shown as `[1]`. Footnotes are numbered in the order of their first reference, not by their labels. The definitions are shown at the end of the document, after a horizontal line, as a numbered list in the order of their numbers; a definition with several paragraphs is laid out like a list item. A definition that is never referenced is not shown. The reference is not a hyperlink, because a terminal hyperlink cannot jump to another place of the output, and there are no back-references from a definition to its reference.

### Themes

A theme sets the colors of the whole document. `--theme` selects it; without the flag, the `TERMD_THEME` environment variable does, and without both termd uses `auto`. A `--theme` on the command line, `--theme=auto` included, wins over the variable, and the variable is then not checked. An unknown name in the value that is used is a usage error with exit status 2, also when the output goes to a pipe; the message names the flag or the variable and lists the valid values, which `termd --help` lists too.

`dark` and `light` use the attributes and the basic colors of the terminal, so they follow its color scheme: bold headings, cyan inline code, underlined blue links, faint list markers, horizontal lines and quote markers, alerts in their colors, bold table headers and frontmatter keys. They differ only in the colors of code blocks, chroma's `github-dark` and `github` styles. `auto` picks one of them, see "Code highlighting".

The other themes use 24-bit colors from the palette of their color scheme, and the chroma style of the same name for code blocks: `dracula`, `nord`, `onedark`, `monokai`, `solarized-dark`, `solarized-light`, `gruvbox`, `gruvbox-light`, `catppuccin-mocha` and `catppuccin-latte`. Headings, table headers and frontmatter keys are bold in the heading color of the theme, links are underlined in its link color, list markers, horizontal lines, quote markers and table borders take its marker color without being faint, and alerts take its blue, green, purple, yellow and red. These themes never query the terminal and do not adapt to its background: `solarized-light`, `gruvbox-light` and `catppuccin-latte` are made for a light background, the others for a dark one.

In every theme, the style of an inner element wins: inline code in a link keeps the inline code color, a link in a heading takes the link color and stays bold. No theme colors paragraph text, emphasis, footnote references, task checkboxes or diagrams, and no theme sets a background color. In a pipe, the output is the same with every theme.

### Code highlighting

A fenced code block is highlighted when the first word of its info string names a language known to [chroma](https://github.com/alecthomas/chroma) (names and common aliases such as `js`, `sh`, `yml`). Blocks without a language or with an unknown one are shown without colors; the text of a block is never changed.

With `--theme=auto`, termd asks the terminal for its background color (OSC 11) and waits at most 100 ms for the answer; a light background selects the `light` theme, anything else the `dark` one. Since the two differ only in code colors, the query is sent only when the document has a code block to highlight. A theme name never sends the query. Code colors and the colors of the named themes are 24-bit when `COLORTERM` is `truecolor` or `24bit`, and the nearest 256-color palette colors otherwise.

### Code files

When chroma recognizes the name of the input file as a language other than markdown or plain text, by its extension or by its whole name (`main.go`, `config.yaml`, `Makefile`, `Dockerfile`), termd shows the whole file as one code block instead of parsing it as markdown: every line as it is in the file, never wrapped, highlighted in a terminal, and scrolled horizontally in the pager when a line is wider than the screen. In a pipe, the output is the text of the file, with control characters and line endings handled as described in "Control characters".

Markdown files, plain text files such as `notes.txt`, files with a name chroma does not recognize (for example `README` without an extension) and standard input are rendered as markdown. The choice is made by the file name only, never by the content.

### Diagrams

Fenced blocks with the info string `mermaid`, `plantuml` or `puml` are treated as diagrams. Supported mermaid diagrams are drawn with [mermaid-ascii](https://github.com/AlexanderGrooff/mermaid-ascii):

- `sequenceDiagram`: participants and actors with aliases, messages, self-messages, notes (`left of`, `right of`, `over`), `loop`, `alt`/`else`, `opt`, `par`/`and`, `critical`, `break`, `rect`, activation (`activate`/`deactivate` and `+`/`-`), `autonumber`, `box`.
- `flowchart` and `graph`: directions, subgraphs, edge labels, node shapes including decisions.
- `erDiagram`: entities and relationships with cardinality markers and labels.

A frontmatter `title` is printed above the diagram. A flowchart that is too wide is first compacted; a left-to-right one that still does not fit is laid out top-to-bottom. A diagram that still does not fit is emitted in full and scrolls horizontally in the pager.

Everything else (other mermaid types, PlantUML, a diagram with a syntax error) is shown as source in a frame labelled with the language, the diagram type and the reason. One broken diagram never stops the rest of the document from rendering.

### Frontmatter

A document whose first line is `---`, which has a closing `---` or `...` line and holds a YAML mapping between them, starts with a table of the keys and their values, in source order and without a header row. A list is shown as its items separated by commas, a nested mapping as one `key: value` line per entry, a multi-line string with its line breaks, and deeper nesting as one-line YAML. The table is laid out like any other table.

A block that is not valid YAML is shown as source in a frame labelled `frontmatter - invalid YAML`, and the rest of the document is rendered as usual. An empty or comment-only block is omitted. Anything else that starts with `---`, for example a horizontal rule followed by text, is rendered as markdown.

### Alerts

A block quote whose first line is `[!NOTE]`, `[!TIP]`, `[!IMPORTANT]`, `[!WARNING]` or `[!CAUTION]`, alone on the line and in any case, is shown as a GitHub alert: the marker line is replaced by the title `Note`, `Tip`, `Important`, `Warning` or `Caution`, and the rest of the quote follows under it. In a terminal the title is bold, and the title and the quote marker take the color of the alert type: with the `dark` and `light` themes from the terminal palette, blue for a note, green for a tip, purple for important, yellow for a warning and red for a caution, and with the other themes from the palette of the theme. In plain text the title follows the usual quote marker. A quote with another marker, such as `[!FOO]`, or with text after the marker on the same line stays a regular quote. Alerts have no icons, because terminals disagree on the width of emoji.

## Limitations

- `stateDiagram`, `mindmap`, `classDiagram`, `gantt` and PlantUML are shown as source.
- Emoji inside diagram labels can misalign, because the diagram library measures text by code points.
- The `box` group frame in sequence diagrams overlaps message lines visually; the diagram stays correct.
- Terminals disagree on the width of emoji with a variation selector (for example `⚠️`). termd follows Unicode and counts 2 columns; in iTerm2 this matches the "Use Unicode version 9+ widths" setting.
- In plain mode a long URL is an unbreakable word and can push a table beyond the output width.
- GNU screen does not support hyperlinks; use `--hyperlinks=never` there.
- Background detection works on macOS, Linux and the BSDs; elsewhere the dark theme is used unless `--theme` or `TERMD_THEME` is given.
- TOML (`+++`) frontmatter is rendered as markdown.

## Roadmap

- Rendering of `stateDiagram` and `mindmap`.
- PlantUML sequence diagrams.
- A built-in pager as an option instead of `less`.
- Inline images and a built-in TUI.

## Development

```sh
make          # list targets
make check    # format check, go.mod tidiness check, go vet (host and Windows) and tests
make run      # render testdata/regression.md
make run FILE=README.md
make golden   # regenerate golden files after an intended output change
```

| Path | Contents |
|---|---|
| `cmd/termd/` | CLI: flags, input, width and terminal detection, pager |
| `internal/render/` | markdown AST walk and block layout |
| `internal/text/` | display width, word wrapping, styled spans |
| `internal/table/` | table layout |
| `internal/diagram/` | diagram detection, mermaid-ascii adapter, framed-source fallback |
| `internal/highlight/` | syntax highlighting |
| `internal/theme/` | color themes |
| `internal/style/` | SGR and OSC 8 escape sequences |
| `internal/termbg/` | terminal background color query |
| `internal/golden/` | golden-file test helper |
| `testdata/` | markdown fixtures and golden outputs |
| `openspec/` | specifications and change proposals |

### Golden files

A golden file is a committed file that holds the expected output of a test. The test renders its input, compares the result byte by byte with the golden file and fails on any difference, printing both versions (`--- want` and `--- got`). This fits termd well: its output is deterministic for a given input, width and mode, and a whole screen of text is easier to review as a file than as assertions in code.

The golden files and the tests that use them:

- `testdata/golden/<name>.w<width>.<mode>.golden` - every `testdata/<name>.md` rendered at widths 40, 60 and 80 in `plain` and `styled` modes (`TestGolden` in `internal/render`).
- `testdata/golden/<file>.styled.golden` - every file in `testdata/codefiles/` rendered as a code file in `styled` mode; its `plain` output must be the file itself (`TestCodeFileGolden` in `internal/render`).
- `testdata/golden/theme.<name>.golden` - `testdata/themes/sample.md` rendered at width 60 in `styled` mode with every theme (`TestThemeGolden` in `internal/render`).
- `internal/diagram/testdata/<name>.golden` - every diagram fixture in the same directory, rendered at width 200 (`TestGolden` in `internal/diagram`).
- `internal/highlight/testdata/<lang>.<theme>.<depth>.golden` - code samples in `go`, `python` and `sh` with the `dark` and `light` themes in `truecolor` and `256` colors (`TestHighlightGolden` in `internal/highlight`).

Styled golden files contain escape sequences: `cat` shows them as colors in a terminal, `cat -v` shows the codes themselves.

Adding a fixture:

1. Create `testdata/<name>.md` (or a diagram fixture in `internal/diagram/testdata/`).
2. Run `make golden` to create its golden files.
3. Read the new files and check that the output is what termd should produce.
4. Commit the fixture together with its golden files.

Changing the output on purpose:

1. Change the code and run `make test`; the affected golden tests fail and show the differences.
2. When every difference is intended, run `make golden`.
3. Review the result with `git diff` and commit the code and the golden files together.

When a golden test fails after a change that should not affect the output, it has found a regression: fix the code, do not update the golden files. Never run `make golden` just to make the tests pass without reading the diff.

Useful commands:

```sh
# one fixture, width and mode
go test ./internal/render -run 'TestGolden/tables/60/plain'
# one diagram
go test ./internal/diagram -run 'TestGolden/er$'
# update the golden files of one fixture only
go test ./internal/render -run 'TestGolden/tables' -update
```

The `-update` flag exists only in the packages with golden tests, so pass it per package, as `make golden` does; `go test ./... -update` fails with `flag provided but not defined: -update`.

### Releases

A release is published from `main` by the Release workflow. It runs `make check`, builds the archives with [GoReleaser](https://goreleaser.com) and creates the version tag only after everything was uploaded:

```sh
gh workflow run release.yml -f version=vX.Y.Z
```

The same build runs locally without publishing:

```sh
make release-check      # validate the GoReleaser config
make release-snapshot   # build the release archives into dist/
```

## Acknowledgements

termd is built on [goldmark](https://github.com/yuin/goldmark) (markdown parsing), [mermaid-ascii](https://github.com/AlexanderGrooff/mermaid-ascii) (diagrams), [chroma](https://github.com/alecthomas/chroma) (syntax highlighting) and [uniseg](https://github.com/rivo/uniseg) (display width).

## License

MIT, see [LICENSE](LICENSE).
