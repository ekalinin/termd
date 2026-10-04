# cli Specification

## Purpose

Defines how termd is invoked, where it reads markdown from, how it chooses the output width and the highlighting theme, and how it behaves when writing to a terminal versus a pipe, including paging.

## Requirements

### Requirement: Input source
termd SHALL read the markdown document from the path given as its single positional argument. When the path is a directory, termd SHALL read the README of that directory and render it as when the path of the README is given as the argument. When the argument is `-`, or when no argument is given and stdin is not a terminal, termd SHALL read the document from stdin.

The README of a directory SHALL be the first of `README.md`, `README.markdown` and `README`, in this order, that is an entry of the directory itself, with the names compared case-insensitively. When several entries differ from a name only in case, the first of them in byte order of their names SHALL be used, so `README.md` comes before `readme.md`. An entry SHALL count only when it is a regular file, directly or through symbolic links; any other entry, such as a directory named `README.md` or a broken symbolic link, SHALL be skipped. termd SHALL NOT look into subdirectories. A symbolic link to a directory given as the argument SHALL be treated as that directory. When the directory has no README, termd SHALL print `termd: no README in <dir>` to stderr, where `<dir>` is the argument as given, and exit with status 1. When the directory or its README cannot be read, termd SHALL print an error naming the directory or the README to stderr and exit with status 1.

#### Scenario: Render a file
- **WHEN** the user runs `termd README.md` and the file exists
- **THEN** termd renders the contents of `README.md` and exits with status 0

#### Scenario: Render the README of a directory
- **WHEN** the user runs `termd docs/` and `docs/` contains `README.md` and `guide.md`
- **THEN** termd renders the contents of `docs/README.md` and exits with status 0

#### Scenario: Several README candidates
- **WHEN** the user runs `termd docs/` and `docs/` contains `README`, `README.markdown` and `README.md`
- **THEN** termd renders the contents of `docs/README.md`

#### Scenario: README.markdown before README
- **WHEN** the user runs `termd docs/` and `docs/` contains `README` and `README.markdown`
- **THEN** termd renders the contents of `docs/README.markdown`

#### Scenario: Lowercase name
- **WHEN** the user runs `termd docs/` and the only README in `docs/` is `readme.md`
- **THEN** termd renders the contents of `docs/readme.md`

#### Scenario: Names that differ only in case
- **WHEN** the user runs `termd docs/` on a case-sensitive file system and `docs/` contains `readme.md` and `README.md`
- **THEN** termd renders the contents of `docs/README.md`

#### Scenario: README entry that is a directory
- **WHEN** the user runs `termd docs/` and `docs/` contains a directory `README.md` and a file `README`
- **THEN** termd renders the contents of `docs/README`

#### Scenario: Symbolic link to a directory
- **WHEN** `link` is a symbolic link to the directory `docs/`, which contains `README.md`, and the user runs `termd link`
- **THEN** termd renders the contents of `docs/README.md`

#### Scenario: README that is a symbolic link
- **WHEN** the user runs `termd docs/` and `docs/README.md` is a symbolic link to `../README.md`
- **THEN** termd renders the contents of `README.md`

#### Scenario: README only in a subdirectory
- **WHEN** the user runs `termd docs/` and `docs/` contains `guide/README.md` but no README of its own
- **THEN** termd prints `termd: no README in docs/` to stderr and exits with status 1

#### Scenario: Directory without a README
- **WHEN** the user runs `termd docs/` and `docs/` contains only `guide.md`
- **THEN** termd prints `termd: no README in docs/` to stderr, writes nothing to stdout and exits with status 1

#### Scenario: Unreadable README
- **WHEN** the user runs `termd docs/` and `docs/README.md` exists but cannot be read
- **THEN** termd prints an error naming `docs/README.md` to stderr and exits with status 1

#### Scenario: Same output as for the README path
- **WHEN** the user runs `termd --width 60 docs/` in a terminal and `docs/README.md` is the README
- **THEN** the output, and whether it is shown in the pager, is the same as for `termd --width 60 docs/README.md`

#### Scenario: Render piped input
- **WHEN** the user runs `cat README.md | termd`
- **THEN** termd renders the piped document and exits with status 0

#### Scenario: Explicit stdin
- **WHEN** the user runs `termd -` with a document piped to stdin
- **THEN** termd renders the piped document

#### Scenario: No input available
- **WHEN** the user runs `termd` with no argument and stdin is a terminal
- **THEN** termd prints usage information to stderr and exits with status 2

#### Scenario: Unreadable file
- **WHEN** the user runs `termd missing.md` and the file does not exist or cannot be read
- **THEN** termd prints an error naming the file to stderr and exits with status 1

### Requirement: Output width
termd SHALL lay out the document for an output width equal to the terminal's column count when stdout is a terminal, and 80 columns when stdout is not a terminal. The `--width N` flag SHALL override the detected width in both cases.

#### Scenario: Terminal width is used
- **WHEN** stdout is a terminal that is 100 columns wide and no `--width` flag is given
- **THEN** paragraphs are wrapped and tables are fitted to 100 columns

#### Scenario: Pipe width defaults to 80
- **WHEN** stdout is redirected to a file and no `--width` flag is given
- **THEN** the document is laid out for 80 columns

#### Scenario: Width flag overrides detection
- **WHEN** the user runs `termd --width 60 README.md` in a 100-column terminal
- **THEN** the document is laid out for 60 columns

#### Scenario: Invalid width
- **WHEN** the user passes `--width 0`, a negative number or a non-numeric value
- **THEN** termd prints a usage error to stderr and exits with status 2

### Requirement: Terminal and pipe output modes
When stdout is a terminal, termd SHALL emit terminal styling (bold, italic, strikethrough, colors). When stdout is not a terminal, termd SHALL emit plain text without any ANSI escape sequences, unless hyperlinks are forced with `--hyperlinks=always`.

#### Scenario: Styled output in a terminal
- **WHEN** termd writes to a terminal
- **THEN** emphasized text is rendered with terminal styles

#### Scenario: Plain output in a pipe
- **WHEN** termd output is piped to another program with default flags
- **THEN** the output contains no ANSI escape sequences

### Requirement: Paging
When stdout is a terminal and the rendered output has more lines than the terminal height or has a line wider than the terminal, termd SHALL display the output through `less` with raw control sequences enabled and long lines chopped instead of wrapped (`less -RS`), so wide blocks can be scrolled horizontally. termd SHALL print directly to stdout when the output fits the terminal height and width, when stdout is not a terminal, when `--no-pager` is given, or when `less` is not available.

#### Scenario: Long document is paged
- **WHEN** termd renders a document that is taller than the terminal and stdout is a terminal
- **THEN** the output is shown in `less -RS`

#### Scenario: Short document with a wide block is paged
- **WHEN** termd renders a document that fits the terminal height but contains a table or diagram wider than the terminal, and stdout is a terminal
- **THEN** the output is shown in `less -RS`, so the wide block is not wrapped by the terminal

#### Scenario: Short document is printed directly
- **WHEN** the rendered document fits within the terminal height and width
- **THEN** termd prints it directly without starting a pager

#### Scenario: Pager disabled
- **WHEN** the user runs `termd --no-pager` on a long document
- **THEN** termd prints the whole output directly to stdout

#### Scenario: Pager missing
- **WHEN** `less` is not found on the system and the document is taller than the terminal
- **THEN** termd prints the output directly and exits with status 0

### Requirement: Hyperlink mode flag
termd SHALL accept `--hyperlinks=auto|always|never`, defaulting to `auto`. In `auto` mode hyperlinks SHALL be emitted as terminal hyperlinks only when stdout is a terminal. Any other value SHALL be a usage error.

#### Scenario: Default mode in a terminal
- **WHEN** termd writes to a terminal without the flag
- **THEN** links are emitted as terminal hyperlinks

#### Scenario: Default mode in a pipe
- **WHEN** termd output is piped without the flag
- **THEN** links are emitted as plain text

#### Scenario: Invalid mode
- **WHEN** the user passes `--hyperlinks=sometimes`
- **THEN** termd prints a usage error to stderr and exits with status 2

### Requirement: Theme selection
termd SHALL accept `--theme` with the value `auto` or the name of a built-in theme: `dark`, `light`, `dracula`, `nord`, `onedark`, `monokai`, `solarized-dark`, `solarized-light`, `gruvbox`, `gruvbox-light`, `catppuccin-mocha` or `catppuccin-latte`. When `--theme` is not given, termd SHALL take the value from the `TERMD_THEME` environment variable when it is set and not empty, and SHALL use `auto` otherwise. A `--theme` given on the command line, including `--theme=auto`, SHALL take precedence over `TERMD_THEME`, and the variable SHALL then be ignored, even when its value is invalid.

An invalid `--theme` value SHALL be a usage error with exit status 2, and the message SHALL name `--theme` and list the valid values. An invalid `TERMD_THEME` value that termd would use SHALL be a usage error with exit status 2, also when the output goes to a pipe, and the message SHALL name `TERMD_THEME` and list the valid values. The `--help` text of `--theme` SHALL list `auto` and every theme name.

With a theme name, termd SHALL use that theme without querying the terminal. With `auto`, when styling is enabled and the document contains a code block to highlight, termd SHALL query the terminal for its background color and use the `light` theme for a light background and the `dark` theme otherwise; if the terminal does not answer within 100 ms, termd SHALL use the `dark` theme. Since `dark` and `light` differ only in the colors of code blocks, a document without a code block to highlight SHALL be rendered with `auto` without a query. Colors SHALL be emitted as 24-bit colors when the `COLORTERM` environment variable is `truecolor` or `24bit`, and as the nearest colors of the 256-color palette otherwise; this applies to code colors and to the colors of the named themes, while basic palette colors are emitted as they are.

#### Scenario: Light background detected
- **WHEN** the terminal reports a white background and a `go` block is rendered with default flags
- **THEN** the block is highlighted with the light theme

#### Scenario: Dark background detected
- **WHEN** the terminal reports a black background and a `go` block is rendered with default flags
- **THEN** the block is highlighted with the dark theme

#### Scenario: Terminal does not answer
- **WHEN** the terminal does not answer the background query
- **THEN** termd uses the dark theme after waiting at most 100 ms

#### Scenario: Explicit theme
- **WHEN** the user runs `termd --theme=light README.md`
- **THEN** code blocks use the light theme and no background query is sent to the terminal

#### Scenario: Named theme
- **WHEN** the user runs `termd --theme=dracula README.md` in a terminal and the document has a `go` block
- **THEN** the document and the block use the `dracula` theme and no background query is sent to the terminal

#### Scenario: Nothing to highlight
- **WHEN** the document contains no code block in a recognized language
- **THEN** termd does not query the terminal

#### Scenario: Terminal without truecolor
- **WHEN** `COLORTERM` is not set and a `go` block is rendered to a terminal
- **THEN** the block is colored using only 256-color palette sequences

#### Scenario: Invalid theme
- **WHEN** the user passes `--theme=blue`
- **THEN** termd prints a usage error that names `--theme` and lists `auto` and every theme name to stderr and exits with status 2

#### Scenario: Theme from the environment
- **WHEN** `TERMD_THEME` is `nord` and the user runs `termd README.md` in a terminal
- **THEN** the document uses the `nord` theme and no background query is sent to the terminal

#### Scenario: Flag over the environment
- **WHEN** `TERMD_THEME` is `nord` and the user runs `termd --theme=light README.md`
- **THEN** the document uses the `light` theme

#### Scenario: Explicit auto over the environment
- **WHEN** `TERMD_THEME` is `nord` and the user runs `termd --theme=auto README.md` in a terminal and the document has a `go` block
- **THEN** termd queries the terminal for its background color and uses the `dark` or the `light` theme

#### Scenario: Empty variable
- **WHEN** `TERMD_THEME` is set to an empty value and the user runs `termd README.md`
- **THEN** termd uses `auto`

#### Scenario: Invalid variable
- **WHEN** `TERMD_THEME` is `drakula` and the user runs `termd README.md > out.txt`
- **THEN** termd prints a usage error that names `TERMD_THEME` and lists `auto` and every theme name to stderr, writes nothing to `out.txt` and exits with status 2

#### Scenario: Invalid variable with a flag
- **WHEN** `TERMD_THEME` is `drakula` and the user runs `termd --theme=dark README.md`
- **THEN** termd renders the document with the `dark` theme and exits with status 0

#### Scenario: Theme names in the help
- **WHEN** the user runs `termd --help`
- **THEN** the description of `--theme` lists `auto` and every theme name

### Requirement: Diagram fallbacks do not fail the run
termd SHALL exit with status 0 when the document was rendered, even if some diagrams were shown as source because they are unsupported or invalid.

#### Scenario: Document with an unsupported diagram
- **WHEN** termd renders a document containing a `stateDiagram-v2` block
- **THEN** the diagram is shown as framed source and termd exits with status 0

### Requirement: Version flag
termd SHALL accept `--version`. With it, termd SHALL print `termd <version>` followed by a newline to stdout and exit with status 0 without reading a document. `<version>` SHALL be the version Go records for the main module when the binary is built:
- the tag for a binary from a release archive and for `go install github.com/ekalinin/termd/cmd/termd@<tag>`;
- a pseudo-version for a build from a git checkout whose commit has no tag, with a `+dirty` suffix when the working tree has uncommitted changes;
- `(devel)` when Go records no version.

#### Scenario: Release binary
- **WHEN** the user runs `termd --version` with the binary from the `v0.1.0` release
- **THEN** termd prints `termd v0.1.0` to stdout and exits with status 0

#### Scenario: Build from a checkout
- **WHEN** termd is built with `make build` from a clean checkout of a commit without a tag, and the user runs `termd --version`
- **THEN** termd prints `termd` followed by a pseudo-version that contains the first 12 characters of the commit hash

#### Scenario: No recorded version
- **WHEN** the user runs `go run ./cmd/termd --version`
- **THEN** termd prints `termd (devel)`

#### Scenario: Version with a file argument
- **WHEN** the user runs `termd --version README.md`
- **THEN** termd prints the version, does not read or render `README.md`, and exits with status 0

### Requirement: Code files
When the input is a file whose name the highlighter recognizes, by its extension or by its whole name, as a language other than markdown or plain text (for example `main.go`, `config.yaml`, `Makefile`, `Dockerfile`), termd SHALL NOT parse the file as markdown. termd SHALL render the whole file as one code block: every line verbatim with its whitespace, never wrapped, and output in full when it is wider than the output width. The file SHALL NOT be rendered as a diagram. When styling is enabled, termd SHALL color the tokens in the language of the file name without changing the text, whitespace or line breaks, and SHALL select the theme as for a code block in a recognized language. When styling is disabled, the output SHALL be the text of the file. Control characters and line endings SHALL be handled as in the text of a markdown document, in both modes. When the file does not end with a line break, termd SHALL output a line break after its last line. An empty file SHALL produce no output. Read errors and exit statuses SHALL be the same as for a markdown file.

termd SHALL render as markdown, regardless of the content: a file whose name the highlighter recognizes as markdown (for example `doc.md`) or as plain text (for example `notes.txt`), a file whose name it does not recognize (for example `README` without an extension), and a document read from stdin.

#### Scenario: Recognized extension
- **WHEN** the user runs `termd main.go > out.txt` and `main.go` holds the lines `package main`, an empty line, `// main does nothing.` and `func main() {}`, followed by a line break
- **THEN** `out.txt` is byte for byte the text of `main.go`, and termd exits with status 0

#### Scenario: Comment lines are not headings
- **WHEN** the user runs `termd config.yaml` with the output to a pipe and the file holds the lines `---`, `# Server settings` and `port: 8080`
- **THEN** the output is these three lines as they are in the file, with no horizontal line, no frontmatter table and no blank line between them

#### Scenario: Recognized whole name
- **WHEN** the user runs `termd Makefile` with the output to a pipe and the file holds the line `build:` followed by a line that is a tab and `go build ./...`
- **THEN** the output is these two lines as they are in the file, the second one starting with the tab

#### Scenario: Highlighted in a terminal
- **WHEN** the user runs `termd main.go` in a terminal
- **THEN** keywords, strings and comments are shown in different colors, and removing the escape sequences from the output yields the text of `main.go`

#### Scenario: Theme of a code file
- **WHEN** the user runs `termd main.go` in a terminal with the default `--theme=auto`
- **THEN** termd queries the terminal for its background color once and highlights the file with the matching theme

#### Scenario: Fence line in a code file
- **WHEN** `main.go` contains a block comment with a line of three backticks followed by a line `# Title`
- **THEN** both lines and every line after them are output as they are in the file

#### Scenario: Long line
- **WHEN** `main.go` has a line 120 columns wide and the output width is 80
- **THEN** the line is output in full, without an ellipsis or an inserted line break, and when stdout is an 80-column terminal the output is shown in `less -RS`

#### Scenario: No line break at the end of the file
- **WHEN** `main.go` ends with `}` without a line break and the output goes to a pipe
- **THEN** the output is the text of the file followed by one line break

#### Scenario: Empty file
- **WHEN** the user runs `termd empty.go` and the file is empty
- **THEN** termd outputs nothing and exits with status 0

#### Scenario: Control characters in a code file
- **WHEN** the user runs `termd main.go > out.txt`, `main.go` uses CRLF line endings and holds the comment `// ` followed by ESC, `]0;x` and BEL
- **THEN** `out.txt` holds the comment as `// ␛]0;x␇`, has LF line endings and contains no control character other than line feed

#### Scenario: Markdown file
- **WHEN** the user runs `termd doc.md` with the output to a pipe and the file holds the lines `first line` and `second line`
- **THEN** it is rendered as markdown: the output is the paragraph `first line second line`

#### Scenario: Plain text file
- **WHEN** the user runs `termd notes.txt` with the output to a pipe and the file holds the lines `first line` and `second line`
- **THEN** it is rendered as markdown: the output is the paragraph `first line second line`

#### Scenario: Unrecognized name
- **WHEN** the user runs `termd README` with the output to a pipe and the file holds the lines `first line` and `second line`
- **THEN** it is rendered as markdown: the output is the paragraph `first line second line`

#### Scenario: Standard input
- **WHEN** the user runs `cat config.yaml | termd` or `termd - < config.yaml` and the file holds the lines `# Server settings` and `port: 8080`
- **THEN** the document is rendered as markdown: the heading `# Server settings`, a blank line and the paragraph `port: 8080`
