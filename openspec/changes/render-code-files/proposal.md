# Proposal

## Why

`termd main.go` or `termd config.yaml` parses the file as markdown: lines are joined into paragraphs and `# comment` lines become headings. glow shows such files as highlighted code. termd already highlights code and scrolls wide code horizontally in the pager.

## What Changes

- When the name of the input file is recognized by the highlighter as a language other than markdown or plain text, by its extension or by its whole name (for example `main.go`, `config.yaml`, `Makefile`, `Dockerfile`), termd renders the whole file as one code block: verbatim, not wrapped, highlighted when styling is enabled, and as a wide block when a line is wider than the output.
- Markdown files, `.txt` files, files with an unrecognized name (including `README` without an extension) and stdin are rendered as markdown, as today. glow treats every extension that is not a markdown one as code, including `.txt`; termd does not.
- When styling is disabled, the output is the text of the file.

Out of scope:

- A flag that selects the language or forces markdown.
- Detecting the language from the content; termd never guesses from content, as for code blocks.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `cli`: a new requirement for input files whose name names a programming language.

## Impact

- `internal/highlight`: a lookup by file name next to the lookup by info string.
- `internal/render`: an entry point that renders a whole file as a code block. Wrapping the file in a markdown fence, as glow does, breaks on files that contain a fence line themselves.
- `cmd/termd/main.go`: chooses between markdown and code by the input file name.
- Tests for a recognized extension, a recognized whole name, `.txt`, `README` and stdin; `README.md`, Usage and "How it works".
