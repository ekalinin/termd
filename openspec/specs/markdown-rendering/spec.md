# markdown-rendering Specification

## Purpose

Defines how termd turns GFM markdown elements into terminal text: display width measurement, wrapping, block and inline formatting, blocks wider than the output and hyperlinks.

## Requirements

### Requirement: GFM parsing
termd SHALL parse documents as CommonMark with the GitHub Flavored Markdown extensions for tables, task lists, strikethrough, autolinks and footnotes.

#### Scenario: GFM extensions are recognized
- **WHEN** a document contains a pipe table, a `- [x]` task item, `~~text~~`, a bare `https://` URL and a footnote reference `[^1]` with the definition `[^1]: Note.`
- **THEN** they are rendered as a table, a checked task item, struck-through text, a link and a footnote respectively

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

When the document is read from a file and hyperlinks are enabled, termd SHALL resolve a relative destination of a link or an image against the directory of that file and write it into the hyperlink as an absolute `file:` URL. A relative destination is a destination without a scheme that has a path and does not start with `/`, for example `docs/guide.md`, `./a.md` or `../b.md`. The URL SHALL consist of `file://`, the host name of the machine, the absolute path of the target with the `.` and `..` segments resolved, and the fragment of the destination, if any; the query of the destination SHALL be dropped. Backslash escapes, character references and percent-encoded characters of the destination SHALL be decoded, and every character that a URL does not allow in a path, such as a space or a non-ASCII character, SHALL be percent-encoded, so that no character is encoded twice. When the host name of the machine cannot be determined, and on Windows, the host SHALL be empty.

termd SHALL write a destination into the hyperlink as written in the document when it has a scheme (`https:`, `mailto:` and any other), when it is fragment-only (`#section`), when it starts with `/` (including `//host/path`), when it is empty, and when it is not a valid URL reference. When the document is read from stdin, or when hyperlinks are disabled, termd SHALL NOT resolve any destination.

#### Scenario: Link in a terminal
- **WHEN** a document contains `[docs](https://example.com/very/long/path)` and hyperlinks are enabled
- **THEN** the visible output is `docs`, it opens `https://example.com/very/long/path` when clicked, and it occupies 4 columns

#### Scenario: Link in a pipe
- **WHEN** the same document is rendered with hyperlinks disabled
- **THEN** the output contains `docs (https://example.com/very/long/path)` with the full URL

#### Scenario: Autolink
- **WHEN** a document contains the bare URL `https://example.com` and hyperlinks are disabled
- **THEN** the output contains `https://example.com` once

#### Scenario: Relative link in a file
- **WHEN** the file `/home/u/proj/README.md` contains `[guide](docs/guide.md)`, it is rendered with hyperlinks enabled, and the host name of the machine is `box`
- **THEN** the visible output is `guide` and its hyperlink is `file://box/home/u/proj/docs/guide.md`

#### Scenario: Relative image in a file
- **WHEN** the same file contains `![arch](img/arch.png)`
- **THEN** the visible output is `[image: arch]` and its hyperlink is `file://box/home/u/proj/img/arch.png`

#### Scenario: Dot segments
- **WHEN** the same file contains `[a](./a.md)` and `[b](../other/b.md)`
- **THEN** their hyperlinks are `file://box/home/u/proj/a.md` and `file://box/home/u/other/b.md`

#### Scenario: Spaces and non-ASCII characters
- **WHEN** the same file contains `[n](<my notes.md>)`, `[n](my%20notes.md)` and `[f](файл.md)`
- **THEN** the first two hyperlinks are both `file://box/home/u/proj/my%20notes.md` and the third is `file://box/home/u/proj/%D1%84%D0%B0%D0%B9%D0%BB.md`

#### Scenario: Fragment and query
- **WHEN** the same file contains `[setup](guide.md#setup)` and `![logo](logo.png?raw=true)`
- **THEN** their hyperlinks are `file://box/home/u/proj/guide.md#setup` and `file://box/home/u/proj/logo.png`

#### Scenario: Destinations that are not resolved
- **WHEN** the same file contains links to `https://example.com`, `mailto:me@example.com`, `#usage`, `/docs/x.md` and `//example.com/x`
- **THEN** the hyperlink of each link is its destination exactly as written

#### Scenario: Empty destination
- **WHEN** the same file contains `[empty]()`
- **THEN** the output contains `empty` without a hyperlink

#### Scenario: Relative file argument
- **WHEN** the user runs `termd --hyperlinks=always docs/README.md` in the directory `/home/u/proj` and the file contains `[a](a.md)`
- **THEN** the hyperlink of `a` is `file://box/home/u/proj/docs/a.md`

#### Scenario: Document from stdin
- **WHEN** the user runs `cat README.md | termd --hyperlinks=always` and the document contains `[guide](docs/guide.md)`
- **THEN** the hyperlink of `guide` is `docs/guide.md`

#### Scenario: Hyperlinks disabled
- **WHEN** the file `/home/u/proj/README.md` contains `[guide](docs/guide.md)` and it is rendered with hyperlinks disabled
- **THEN** the output contains `guide (docs/guide.md)`

#### Scenario: Unknown host name
- **WHEN** the host name of the machine cannot be determined and the file `/home/u/proj/README.md` contains `[guide](docs/guide.md)`
- **THEN** the hyperlink of `guide` is `file:///home/u/proj/docs/guide.md`

#### Scenario: Windows
- **WHEN** termd runs on Windows and the file `C:\proj\README.md` contains `[guide](docs/guide.md)`
- **THEN** the hyperlink of `guide` is `file:///C:/proj/docs/guide.md`

### Requirement: Images and raw HTML
termd SHALL render an image as its alt text in the form `[image: alt]`, linked to the image URL under the same rules as hyperlinks, and SHALL render raw HTML blocks and inline HTML as their source text.

#### Scenario: Image
- **WHEN** a document contains `![architecture](docs/arch.png)` and hyperlinks are disabled
- **THEN** the output contains `[image: architecture] (docs/arch.png)`

#### Scenario: Raw HTML
- **WHEN** a document contains `<details><summary>More</summary>text</details>`
- **THEN** the HTML is output as source text

### Requirement: YAML frontmatter
When the first line of a document is `---`, a later line is `---` or `...`, and the lines between them form a YAML mapping, termd SHALL NOT render these lines as markdown. termd SHALL render the mapping as the first block of the document: a table with two columns, the key and its value, with one row per key in source order, without a header row and without a horizontal rule. Keys SHALL be bold when styling is enabled. The table SHALL follow the table layout rules: column separators at the same columns on every line, words never split, the value column wrapping first, and a table that does not fit at its minimum widths emitted as a wide block. A UTF-8 byte order mark before the first line and CRLF line endings SHALL NOT prevent the detection.

termd SHALL show a scalar value as its text without quotes, a list of scalars as its items separated by `, `, a mapping of scalars as one `key: value` line per entry, and a multi-line string with its line breaks. Any other value SHALL be shown as YAML in flow style on one line. A key without a value SHALL have an empty value cell.

When the lines between the delimiters are not valid YAML, termd SHALL show them as source in a frame labelled `frontmatter - invalid YAML`, render the rest of the document normally and exit with status 0. When the lines between the delimiters are empty or contain only YAML comments, termd SHALL omit them from the output. When they are valid YAML but not a mapping, when there is no closing line, or when the `---` line is not the first line of the document, termd SHALL render the document as CommonMark.

#### Scenario: Frontmatter table
- **WHEN** a document starts with `---`, `title: Doc`, `tags: [a, b]`, `---` followed by `# Hello`
- **THEN** the output starts with a row `title` and `Doc` and a row `tags` and `a, b`, separated by the column separator, then a blank line and `# Hello`, and it contains no horizontal line and no `## title` heading

#### Scenario: Nested mapping
- **WHEN** the frontmatter contains `author:` with the indented entries `name: Eugene` and `url: https://example.com`
- **THEN** the value cell of `author` holds `name: Eugene` and `url: https://example.com` on two lines

#### Scenario: Multi-line string
- **WHEN** the frontmatter contains `description: |` followed by the indented lines `line one` and `line two`
- **THEN** the value cell of `description` holds `line one` and `line two` on two lines

#### Scenario: Deeper nesting
- **WHEN** the frontmatter contains `authors:` with the list items `- name: A` and `- name: B`
- **THEN** the value cell of `authors` holds `[{name: A}, {name: B}]`

#### Scenario: Long value
- **WHEN** the frontmatter contains a `description` of 30 words and the output width is 40
- **THEN** the description wraps within the value column, `description` stays on one line and no line of the table is wider than 40 columns

#### Scenario: Plain mode
- **WHEN** a document with frontmatter is rendered to a pipe
- **THEN** the table contains no escape sequences

#### Scenario: Invalid YAML
- **WHEN** a document starts with `---`, `title: [unclosed`, `---` followed by `# Hello`
- **THEN** `title: [unclosed` is shown in a frame labelled `frontmatter - invalid YAML`, `# Hello` is rendered after it, and termd exits with status 0

#### Scenario: Empty frontmatter
- **WHEN** a document starts with `---`, `---` followed by `# Hello`
- **THEN** the output starts with `# Hello`

#### Scenario: Not a mapping
- **WHEN** a document starts with `---`, `Some text`, `---`
- **THEN** it is rendered as CommonMark: a horizontal line followed by the heading `## Some text`

#### Scenario: No closing line
- **WHEN** a document starts with `---`, `title: Doc` and has no other `---` or `...` line
- **THEN** it is rendered as CommonMark: a horizontal line followed by the paragraph `title: Doc`

#### Scenario: Byte order mark and CRLF
- **WHEN** a document starts with a UTF-8 byte order mark and uses CRLF line endings, and its first lines are `---`, `title: Doc`, `---`
- **THEN** the output starts with the frontmatter table

### Requirement: GitHub alerts
When a block quote starts with a paragraph whose first line is `[!NOTE]`, `[!TIP]`, `[!IMPORTANT]`, `[!WARNING]` or `[!CAUTION]`, alone on the line and compared case-insensitively, termd SHALL render the block quote as an alert: the marker line SHALL be replaced by the title `Note`, `Tip`, `Important`, `Warning` or `Caution`, and the rest of the quote SHALL follow under the title, laid out by the block quote rules. A block quote that holds only the marker line SHALL be rendered as the title alone. Block quotes nested in other block quotes or in list items SHALL be detected as alerts the same way.

When styling is enabled, the title and the quote marker on every line of the alert SHALL be shown in the color of the alert type, taken from the basic palette of the terminal: blue for note, green for tip, purple (magenta) for important, yellow for warning and red for caution. The title SHALL also be bold. When styling is disabled, the title SHALL be plain text after the usual quote marker. Alerts SHALL NOT have icons.

A block quote whose first line holds another marker, for example `[!FOO]`, or text after the marker SHALL be rendered as a regular block quote, with the marker as text.

#### Scenario: Note alert
- **WHEN** a document contains `> [!NOTE]` and `> Useful info.` and is rendered to a pipe
- **THEN** the output is `│ Note` followed by `│ Useful info.`, with no escape sequences

#### Scenario: Alert colors
- **WHEN** a note, a tip, an important, a warning and a caution alert are rendered to a terminal
- **THEN** their quote markers are blue, green, magenta, yellow and red (SGR `34`, `32`, `35`, `33` and `31`) and not faint, and each title is bold in the color of its alert, for example `Note` in SGR `1;34`

#### Scenario: Case-insensitive marker
- **WHEN** a document contains `> [!warning]` and `> Careful.`
- **THEN** the output is `│ Warning` followed by `│ Careful.`

#### Scenario: Marker only
- **WHEN** a document contains only `> [!TIP]`
- **THEN** the output is the single line `│ Tip`

#### Scenario: Several blocks
- **WHEN** a document contains `> [!IMPORTANT]`, `> First.`, `>` and `> Second.`
- **THEN** the output is `│ Important`, `│ First.`, `│` and `│ Second.`

#### Scenario: Long alert
- **WHEN** a note alert holds a paragraph longer than the output width of 30
- **THEN** every output line of the alert starts with the quote marker and no line is wider than 30 columns

#### Scenario: Nested quote inside an alert
- **WHEN** a document contains `> [!NOTE]`, `> Text.`, `>` and `> > Quoted.`
- **THEN** the output is `│ Note`, `│ Text.`, `│` and `│ │ Quoted.`, and in a terminal the inner quote marker is faint as in a regular quote

#### Scenario: Alert in a list item
- **WHEN** a list item `- item` continues with the indented lines `> [!TIP]` and `> Hint.`
- **THEN** the output is `• item`, `  │ Tip` and `  │ Hint.`

#### Scenario: Text after the marker
- **WHEN** a document contains `> [!NOTE] Useful info.`
- **THEN** it is rendered as a regular block quote: `│ [!NOTE] Useful info.`

#### Scenario: Unknown marker
- **WHEN** a document contains `> [!FOO]` and `> Text.`
- **THEN** it is rendered as a regular block quote: `│ [!FOO] Text.`

### Requirement: Control characters in document text
termd SHALL show every C0 control character (U+0000 to U+001F) except tab and line feed, and DEL (U+007F), that occurs in the text of a document as its Unicode control picture: the character U+2400 plus the code point for U+0000 to U+001F, for example ESC as `␛` (U+241B), BEL as `␇` (U+2407) and NUL as `␀` (U+2400), and DEL as `␡` (U+2421). A control picture SHALL take one column. termd SHALL show C1 control characters (U+0080 to U+009F) and bytes that are not valid UTF-8 as `�` (U+FFFD).

This SHALL apply to all text that comes from the document: paragraphs, headings, code blocks with and without highlighting, tables, link text and destinations, image alt text, raw HTML, frontmatter (the key-value table and the invalid YAML frame) and diagrams, including framed diagram source. It SHALL also apply to control characters that the document writes as character references, such as `&#27;`, or as escapes in double-quoted YAML strings of the frontmatter, such as `"\e"`. A link destination SHALL NOT be able to end or extend the OSC 8 sequence termd writes around a link.

termd SHALL treat CRLF and a lone CR in the source as line endings, the same as a line feed, as CommonMark does: a CRLF document SHALL render like the same document with LF line endings, including code blocks.

The escape sequences termd emits itself for styles and hyperlinks SHALL NOT change, and a document without control characters other than tab and line feed SHALL render as before.

In the scenarios, `\033` stands for ESC, `\007` for BEL, `\000` for NUL, `\177` for DEL and `\r` for CR in the document source.

#### Scenario: Escape sequences in a paragraph
- **WHEN** a paragraph contains `hello \033[31mRED\033[0m and \033]0;pwned\007 title` and the document is rendered to a pipe
- **THEN** the output is `hello ␛[31mRED␛[0m and ␛]0;pwned␇ title` and contains no ESC or BEL character

#### Scenario: Styled text keeps termd's own escape sequences
- **WHEN** a paragraph contains `**bold \033[31m**` and styling is enabled
- **THEN** the output is the bold sequence `\033[1m`, the text `bold ␛[31m` and the reset `\033[0m`

#### Scenario: NUL, DEL, C1 controls and invalid UTF-8
- **WHEN** a paragraph contains `a\000b\177c`, then the character U+009B, then `d`, then the byte 0xFF, then `e`
- **THEN** the output is `a␀b␡c�d�e`

#### Scenario: Character references
- **WHEN** a paragraph contains `&#27;[31m red &#7; &#13; &#x9b;`
- **THEN** the output is `␛[31m red ␇ ␍ �`

#### Scenario: Control picture width in a table
- **WHEN** a table has the header `Key | Value` and the row `a\033b | x`
- **THEN** the cell reads `a␛b`, it is measured as 3 columns, and the column separator is at the same column on every line of the table

#### Scenario: Code block with and without highlighting
- **WHEN** a fenced block with the info string `go` contains `fmt.Println("\033[31m")` and a fenced block without an info string contains `echo \007`, and styling is enabled
- **THEN** removing termd's style sequences from the output yields `fmt.Println("␛[31m")` and `echo ␇`

#### Scenario: Link destination with escape sequences
- **WHEN** a document contains `[docs](https://example.com/\033]0;pwned\007)` and hyperlinks are enabled without styling
- **THEN** the output is `\033]8;;https://example.com/␛]0;pwned␇\033\docs\033]8;;\033\`, with the visible text `docs`

#### Scenario: Link destination in a pipe
- **WHEN** the same document is rendered with hyperlinks disabled
- **THEN** the output is `docs (https://example.com/␛]0;pwned␇)`

#### Scenario: Image alt text and raw HTML
- **WHEN** a document contains `![alt\033\007](x.png)` and `<span>\033[31m</span>`, and hyperlinks are disabled
- **THEN** the output contains `[image: alt␛␇] (x.png)` and `<span>␛[31m</span>`

#### Scenario: Frontmatter values
- **WHEN** the frontmatter contains `title: "\e]0;pwned\a"` with YAML escapes and `raw: a\033b`
- **THEN** the value cell of `title` reads `␛]0;pwned␇` and the value cell of `raw` reads `a␛b`

#### Scenario: Invalid YAML frontmatter
- **WHEN** a document starts with `---`, `title: [\033[31m`, `---`
- **THEN** the frame labelled `frontmatter - invalid YAML` shows `title: [␛[31m`

#### Scenario: Diagrams
- **WHEN** a mermaid flowchart has a node labelled `\033[31mStart`, and an unsupported `stateDiagram-v2` block contains `\033[31m`
- **THEN** the flowchart box shows `␛[31mStart` with its border lines aligned, and the framed source of the unsupported diagram shows `␛[31m`

#### Scenario: CRLF line endings
- **WHEN** a document with CRLF line endings contains a paragraph of two lines and a fenced code block with the lines `code a` and `code b`
- **THEN** the output is the same as for the document with LF line endings, and the code lines are `code a` and `code b` without a CR

#### Scenario: Lone CR
- **WHEN** a paragraph contains `one\rtwo` and a fenced code block contains `three\rfour`
- **THEN** the paragraph is rendered as `one two` and the code block as the two lines `three` and `four`

### Requirement: Footnotes
termd SHALL render a footnote reference as the number of its footnote in square brackets, for example `[1]`. Footnotes SHALL be numbered from 1 in the order of their first reference, regardless of their labels, and every reference to the same footnote SHALL show the same number. The reference SHALL NOT be a hyperlink, also when hyperlinks are enabled. A reference to a label without a definition SHALL be rendered as its source text.

termd SHALL render the definitions of the referenced footnotes at the end of the document, wherever they stand in the source: a horizontal line spanning the output width, the same as a thematic break, followed by a numbered list of the definitions in the order of their numbers. A definition SHALL be laid out like an item of an ordered list: wrapped lines and further blocks of the definition are indented to the start of the item text. When every definition consists of one block, the definitions SHALL follow each other without blank lines; when a definition has several blocks, the definitions and the blocks inside them SHALL be separated by blank lines, as in a loose list.

A definition that is never referenced SHALL NOT be shown, and when no definition is referenced, termd SHALL NOT render the horizontal line. A footnote definition SHALL NOT be treated as a link reference definition. termd SHALL NOT render back-references from a definition to its references.

#### Scenario: Footnote reference and definition
- **WHEN** the document `Text[^1].`, a blank line, `[^1]: Note.` is rendered with hyperlinks disabled at width 40
- **THEN** the output is `Text[1].`, a blank line, a horizontal line 40 columns wide, a blank line and `1. Note.`, and it contains neither `^1` nor `(Note.)`

#### Scenario: Reference is not a hyperlink
- **WHEN** the same document is rendered with hyperlinks enabled
- **THEN** the output contains `[1]` and no hyperlink escape sequence

#### Scenario: Numbering by first reference
- **WHEN** a document contains the paragraph `A[^b] B[^a] C[^b]` followed by the definitions `[^a]: Alpha.` and `[^b]: Beta.`
- **THEN** the paragraph is rendered as `A[1] B[2] C[1]` and the definitions as `1. Beta.` and `2. Alpha.` on consecutive lines

#### Scenario: Definitions in the middle of the document
- **WHEN** a document contains the paragraph `First[^1].`, the definition `[^1]: Note.` and then the paragraph `Last.`
- **THEN** `Last.` is rendered before the horizontal line and `1. Note.` after it

#### Scenario: Unreferenced definition
- **WHEN** a document contains `Text[^1].` and the definitions `[^1]: Used.` and `[^2]: Unused.`
- **THEN** the definitions are rendered as `1. Used.` only, and the output does not contain `Unused.`

#### Scenario: No referenced definition
- **WHEN** a document contains the paragraph `Text.` and the definition `[^1]: Unused.`
- **THEN** the output is `Text.` without a horizontal line

#### Scenario: Definition with several paragraphs
- **WHEN** a document contains `Text[^1].` and the definition `[^1]: First paragraph.` followed by a blank line and the indented line `    Second paragraph.`
- **THEN** the definitions are rendered as `1. First paragraph.`, a blank line and `   Second paragraph.`, indented to the start of the item text

#### Scenario: Undefined reference
- **WHEN** a document contains `Text[^x].` and no definition of `x`
- **THEN** the output is `Text[^x].` without a horizontal line

#### Scenario: Footnotes with frontmatter
- **WHEN** a document starts with the frontmatter `---`, `title: Doc`, `---` and continues with `Text[^1].` and the definition `[^1]: Note.`
- **THEN** the output starts with the frontmatter table, contains `Text[1].` and ends with `1. Note.`
