# cli Specification

## Purpose

Defines how termd is invoked, where it reads markdown from, how it chooses the output width and the highlighting theme, and how it behaves when writing to a terminal versus a pipe, including paging.

## Requirements

### Requirement: Input source
termd SHALL read the markdown document from the file path given as its single positional argument. When the argument is `-`, or when no argument is given and stdin is not a terminal, termd SHALL read the document from stdin.

#### Scenario: Render a file
- **WHEN** the user runs `termd README.md` and the file exists
- **THEN** termd renders the contents of `README.md` and exits with status 0

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

### Requirement: Highlighting theme selection
termd SHALL accept `--theme=auto|dark|light`, defaulting to `auto`, and SHALL treat any other value as a usage error with exit status 2. With `dark` or `light`, termd SHALL use the corresponding theme without querying the terminal. With `auto`, when styling is enabled and the document contains a code block to highlight, termd SHALL query the terminal for its background color and use the light theme for a light background and the dark theme otherwise; if the terminal does not answer within 100 ms, termd SHALL use the dark theme. Colors SHALL be emitted as 24-bit colors when the `COLORTERM` environment variable is `truecolor` or `24bit`, and as the nearest colors of the 256-color palette otherwise.

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

#### Scenario: Nothing to highlight
- **WHEN** the document contains no code block in a recognized language
- **THEN** termd does not query the terminal

#### Scenario: Terminal without truecolor
- **WHEN** `COLORTERM` is not set and a `go` block is rendered to a terminal
- **THEN** the block is colored using only 256-color palette sequences

#### Scenario: Invalid theme
- **WHEN** the user passes `--theme=blue`
- **THEN** termd prints a usage error to stderr and exits with status 2

### Requirement: Diagram fallbacks do not fail the run
termd SHALL exit with status 0 when the document was rendered, even if some diagrams were shown as source because they are unsupported or invalid.

#### Scenario: Document with an unsupported diagram
- **WHEN** termd renders a document containing a `stateDiagram-v2` block
- **THEN** the diagram is shown as framed source and termd exits with status 0
