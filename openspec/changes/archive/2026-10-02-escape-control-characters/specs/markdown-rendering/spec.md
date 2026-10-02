# Spec Delta

## ADDED Requirements

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
