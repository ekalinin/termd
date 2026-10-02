# Spec Delta

## ADDED Requirements

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
