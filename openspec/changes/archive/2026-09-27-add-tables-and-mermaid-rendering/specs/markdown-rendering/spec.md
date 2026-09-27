# Spec Delta

## Purpose

Defines how termd turns GFM markdown elements into terminal text: display width measurement, wrapping, block and inline formatting, blocks wider than the output and hyperlinks.

## ADDED Requirements

### Requirement: GFM parsing
termd SHALL parse documents as CommonMark with the GitHub Flavored Markdown extensions for tables, task lists, strikethrough and autolinks.

#### Scenario: GFM extensions are recognized
- **WHEN** a document contains a pipe table, a `- [x]` task item, `~~text~~` and a bare `https://` URL
- **THEN** they are rendered as a table, a checked task item, struck-through text and a link respectively

### Requirement: Display width measurement
termd SHALL measure text by its displayed width in terminal columns, counting grapheme clusters rather than code points or bytes: East Asian wide characters and emoji (including ZWJ sequences and emoji with a variation selector) SHALL count as 2 columns, combining marks as 0, and escape sequences (styles, hyperlinks) as 0. All wrapping, padding and alignment SHALL use this measurement.

#### Scenario: ZWJ emoji sequence
- **WHEN** a table cell contains the family emoji `👨‍👩‍👧`
- **THEN** it is measured as 2 columns and the column separators of that row line up with the other rows

#### Scenario: Emoji with variation selector
- **WHEN** a table cell contains `⚠️` (U+26A0 followed by U+FE0F)
- **THEN** it is measured as 2 columns

#### Scenario: CJK and Cyrillic text
- **WHEN** a paragraph contains `日本語` and `кириллица`
- **THEN** each CJK character counts as 2 columns and each Cyrillic letter as 1 column

#### Scenario: Styled text
- **WHEN** a table cell contains `**bold**` rendered with terminal styles
- **THEN** the cell is measured as 4 columns

### Requirement: Paragraph wrapping
termd SHALL wrap paragraph text at whitespace so that no line exceeds the output width. A single word longer than the available width SHALL be broken at the width limit.

#### Scenario: Long paragraph
- **WHEN** a paragraph is longer than the output width
- **THEN** it is split into lines at spaces and no line is wider than the output width

#### Scenario: Word longer than the width
- **WHEN** a paragraph contains a 120-character word and the output width is 80
- **THEN** the word is broken so that no line exceeds 80 columns

### Requirement: Headings
termd SHALL render headings on their own line, separated from the preceding content by a blank line, keeping the `#` markers so the level stays visible, and in bold when styling is enabled.

#### Scenario: Heading levels
- **WHEN** a document contains `# Title` and `### Section`
- **THEN** they are rendered as `# Title` and `### Section`, bold in a terminal

### Requirement: Inline formatting
termd SHALL render bold, italic, strikethrough and inline code with terminal styles when styling is enabled, and as plain text without markdown markers when styling is disabled.

#### Scenario: Styled inline formatting
- **WHEN** a paragraph contains `**bold**`, `*italic*`, `~~gone~~` and `` `code` `` and stdout is a terminal
- **THEN** each span is rendered with its own terminal style and without the markdown markers

#### Scenario: Plain inline formatting
- **WHEN** the same paragraph is rendered to a pipe
- **THEN** the output reads `bold italic gone code` with no markers and no escape sequences

### Requirement: Lists
termd SHALL render unordered items with a bullet, ordered items with their number, nested lists with additional indentation and task items with `[ ]` or `[x]`. Wrapped lines of an item SHALL be indented to the start of the item text.

#### Scenario: Nested list
- **WHEN** a list item contains a nested list
- **THEN** the nested items are indented further than their parent

#### Scenario: Wrapped list item
- **WHEN** a list item is longer than the output width
- **THEN** its continuation lines start at the same column as the item text, not at the bullet

#### Scenario: Task list
- **WHEN** a document contains `- [ ] todo` and `- [x] done`
- **THEN** they are rendered as `[ ] todo` and `[x] done` with their bullets

### Requirement: Block quotes
termd SHALL prefix every line of a block quote, including wrapped lines, with a quote marker, and wrap the quoted content to the output width minus the marker.

#### Scenario: Wrapped quote
- **WHEN** a block quote is longer than the output width
- **THEN** every output line of the quote starts with the quote marker and no line exceeds the output width

### Requirement: Code blocks
termd SHALL render fenced and indented code blocks verbatim, preserving whitespace and without wrapping lines. Fenced blocks whose info string names a diagram language are handled by diagram rendering instead.

#### Scenario: Whitespace is preserved
- **WHEN** a code block contains indented lines and runs of spaces
- **THEN** the lines are output with identical whitespace

#### Scenario: Long code line
- **WHEN** a code block line is wider than the output width
- **THEN** the line is output in full as part of a wide block

### Requirement: Syntax highlighting
When styling is enabled, termd SHALL color the tokens of a fenced code block whose info string's first word names a language termd recognizes, including common aliases (for example `js`, `sh`, `yml`). Blocks without an info string or with an unrecognized language SHALL be rendered without colors. Highlighting SHALL NOT change the text, whitespace or line breaks of the block. When styling is disabled, code blocks SHALL contain no escape sequences.

#### Scenario: Known language
- **WHEN** a fenced block with the info string `go` is rendered to a terminal
- **THEN** keywords, strings and comments are shown in different colors, and removing the escape sequences from the output yields the original source

#### Scenario: Unknown or missing language
- **WHEN** a code block has no info string or the info string `foo`
- **THEN** it is rendered without colors

#### Scenario: Plain mode
- **WHEN** a fenced block with the info string `go` is rendered to a pipe
- **THEN** the output contains no escape sequences

### Requirement: Thematic breaks
termd SHALL render a thematic break (`---`, `***`, `___`) as a horizontal line spanning the output width.

#### Scenario: Horizontal rule
- **WHEN** a document contains `---` between two paragraphs
- **THEN** a horizontal line as wide as the output width is rendered between them

### Requirement: Wide blocks
Blocks that cannot be wrapped without losing structure (code blocks, tables that do not fit at their minimum widths, diagrams wider than the output) SHALL be emitted at their natural width. termd SHALL NOT truncate these lines or insert line breaks into them.

#### Scenario: Wide block is not truncated
- **WHEN** a diagram is 120 columns wide and the output width is 80
- **THEN** each of its lines is output in full, 120 columns wide, without an ellipsis or an inserted line break

### Requirement: Hyperlinks
When hyperlinks are enabled, termd SHALL render each link as its visible text wrapped in an OSC 8 terminal hyperlink to the link destination, and SHALL measure only the visible text. When hyperlinks are disabled, termd SHALL render a link as `text (url)`, or as just the URL when the text equals the URL. termd SHALL NOT truncate URLs.

#### Scenario: Link in a terminal
- **WHEN** a document contains `[docs](https://example.com/very/long/path)` and hyperlinks are enabled
- **THEN** the visible output is `docs`, it opens `https://example.com/very/long/path` when clicked, and it occupies 4 columns

#### Scenario: Link in a pipe
- **WHEN** the same document is rendered with hyperlinks disabled
- **THEN** the output contains `docs (https://example.com/very/long/path)` with the full URL

#### Scenario: Autolink
- **WHEN** a document contains the bare URL `https://example.com` and hyperlinks are disabled
- **THEN** the output contains `https://example.com` once

### Requirement: Images and raw HTML
termd SHALL render an image as its alt text in the form `[image: alt]`, linked to the image URL under the same rules as hyperlinks, and SHALL render raw HTML blocks and inline HTML as their source text.

#### Scenario: Image
- **WHEN** a document contains `![architecture](docs/arch.png)` and hyperlinks are disabled
- **THEN** the output contains `[image: architecture] (docs/arch.png)`

#### Scenario: Raw HTML
- **WHEN** a document contains `<details><summary>More</summary>text</details>`
- **THEN** the HTML is output as source text
