# Spec Delta

## RENAMED Requirements

- FROM: `### Requirement: Highlighting theme selection`
- TO: `### Requirement: Theme selection`

## MODIFIED Requirements

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
