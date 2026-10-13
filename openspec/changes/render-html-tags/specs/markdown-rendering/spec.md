# Spec Delta

## MODIFIED Requirements

### Requirement: Images and raw HTML
termd SHALL render an image as its alt text in the form `[image: alt]`, or as `[image]` when the alt text is empty or consists of whitespace only, linked to the image URL under the same rules as hyperlinks. termd SHALL render raw HTML blocks and inline HTML according to the HTML mode: in the mode `render`, the default, termd SHALL interpret their tags as described in "HTML tags"; in the mode `source`, termd SHALL render them as their source text.

#### Scenario: Image
- **WHEN** a document contains `![architecture](docs/arch.png)` and hyperlinks are disabled
- **THEN** the output contains `[image: architecture] (docs/arch.png)`

#### Scenario: Image without alt text
- **WHEN** a document contains `![](logo.png)` and `![ ](icon.png)` and hyperlinks are disabled
- **THEN** the output contains `[image] (logo.png)` and `[image] (icon.png)`

#### Scenario: Raw HTML
- **WHEN** a document contains `<details><summary>More</summary>text</details>` and styling is disabled
- **THEN** the output is `▾ More` followed by `  text`

#### Scenario: Raw HTML as source
- **WHEN** the same document is rendered with `--html=source`
- **THEN** the HTML is output as source text, `<details><summary>More</summary>text</details>`

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
- **WHEN** a document contains `![alt\033\007](x.png)`, `<span>\033[31m</span>` and `<span>&#27;[0m</span>`, and hyperlinks are disabled
- **THEN** the output contains `[image: alt␛␇] (x.png)`, `␛[31m` and `␛[0m`, and with `--html=source` it contains `<span>␛[31m</span>`

#### Scenario: HTML link destination with character references
- **WHEN** a document contains `<a href="https://example.com/&#27;]0;pwned&#7;">docs</a>` and hyperlinks are enabled without styling
- **THEN** the output is `\033]8;;https://example.com/␛]0;pwned␇\033\docs\033]8;;\033\`, with the visible text `docs`

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

## ADDED Requirements

### Requirement: HTML tags
When the HTML mode is `render`, termd SHALL interpret the tags of HTML blocks and of inline HTML instead of showing their source:

- `img` SHALL be rendered as an image with the value of `alt` as its alt text and the value of `src` as its destination.
- `a` with an `href` SHALL be rendered as a link around its content to the value of `href`; the content of `a` without `href` SHALL be kept as text.
- `p` and `div` SHALL start and end a paragraph; their attributes, such as `align`, SHALL be ignored.
- `br` SHALL be a hard line break.
- `h1` to `h6` SHALL be rendered as a heading of the same level.
- `details` SHALL always show its content: the content of its `summary` on its own line after the marker `▾` and a space, and the rest of its content below it, indented by 2 columns. A `details` without `summary` SHALL show `Details` after the marker. The marker SHALL have the marker style of the theme, and the summary SHALL get no style of its own.
- `kbd` and `code` SHALL be rendered in the inline code style, `b` and `strong` in bold, `i` and `em` in italic.
- Every other tag, such as `span`, `picture`, `source`, `sup` or `sub`, SHALL be removed and its content kept. The content of `script` and `style` and HTML comments SHALL be removed.

Images and links from HTML SHALL follow the rules of markdown images and links, including terminal hyperlinks, `text (url)` when hyperlinks are disabled and relative destinations resolved against the directory of the document file. Character references in text and in attribute values SHALL be decoded.

The text of HTML SHALL be laid out like the text of a paragraph: whitespace, including the line breaks of the source, collapses into one space, and lines wrap at the output width. The paragraphs, headings and `details` of an HTML block SHALL be separated by blank lines like markdown blocks, and an HTML block without text SHALL give no output and no blank line. Markdown inside an HTML block SHALL NOT be parsed, as in CommonMark. The summary line of a `details` SHALL be followed by its content without a blank line; the blocks of the content SHALL be separated by blank lines, and a nested `details` SHALL be indented by 2 more columns.

A start tag and its end tag can be in different parts of the document. A tag in a paragraph, a heading or a table cell SHALL apply to the content up to its end tag in the same paragraph, heading or table cell, and SHALL end with it when it is not closed there. A `details` SHALL contain the markdown blocks between the HTML block with its start tag and the HTML block with its end tag in the same container, such as the document, a list item or a block quote, and a `details` that is not closed SHALL end with its container. An end tag without a start tag SHALL be ignored. In a paragraph, a heading or a table cell, the tags `p`, `div`, `h1` to `h6`, `details` and `summary` SHALL break the line.

Text that cannot be read as a tag, such as an unclosed `<p align="center`, SHALL be shown as text, and HTML SHALL never make the run fail. When styling is disabled, the layout SHALL be the same, without styles.

#### Scenario: Centered logo
- **WHEN** a document contains the lines `<p align="center">`, `  <img src="logo.png" alt="termd" width="200">` and `</p>`, and hyperlinks are disabled
- **THEN** the output is `[image: termd] (logo.png)`

#### Scenario: Image without alt
- **WHEN** a document contains `<img src="logo.png">` and hyperlinks are disabled
- **THEN** the output is `[image] (logo.png)`

#### Scenario: Row of badges
- **WHEN** a document contains the lines `<p align="center">`, `<a href="https://ci.example.com"><img src="https://ci.example.com/badge.svg" alt="CI"></a>`, `<a href="https://pkg.go.dev/x"><img src="https://pkg.go.dev/badge/x.svg"></a>` and `</p>`, and hyperlinks are enabled
- **THEN** the visible output is `[image: CI] [image]` on one line, and the images open `https://ci.example.com` and `https://pkg.go.dev/x` when clicked

#### Scenario: Linked image in a pipe
- **WHEN** the same document is rendered with hyperlinks disabled
- **THEN** the first image is shown as `[image: CI] (https://ci.example.com/badge.svg) (https://ci.example.com)`, as a markdown image inside a link

#### Scenario: Header of a README
- **WHEN** a document contains the lines `<p align="center"><img src="logo.png" alt="logo"></p>`, `<h1 align="center">termd</h1>` and `<p align="center">Markdown in the terminal</p>`, and styling and hyperlinks are disabled
- **THEN** the output is `[image: logo] (logo.png)`, a blank line, `# termd`, a blank line and `Markdown in the terminal`

#### Scenario: Line break
- **WHEN** a paragraph contains `one<br>two`
- **THEN** the output is `one` and `two` on two lines

#### Scenario: Line break in a table cell
- **WHEN** a table cell contains `one<br>two`
- **THEN** the cell shows `one` and `two` on two lines and is measured as 3 columns

#### Scenario: Details around markdown blocks
- **WHEN** a document contains `<details>`, `<summary>Install</summary>`, a blank line, `Run **this**:`, a blank line, a fenced code block with the line `go install x`, a blank line, `</details>`, a blank line and `After.`, and styling is disabled
- **THEN** the output is `▾ Install`, `  Run this:`, a blank line, `  go install x`, a blank line and `After.`

#### Scenario: Details without summary
- **WHEN** a document contains `<details>text</details>` and styling is disabled
- **THEN** the output is `▾ Details` followed by `  text`

#### Scenario: Nested details
- **WHEN** a document contains `<details><summary>A</summary>`, a blank line, `<details><summary>B</summary>`, a blank line, `text`, a blank line, `</details>`, a blank line and `</details>`, and styling is disabled
- **THEN** the output is `▾ A`, `  ▾ B` and `    text` on consecutive lines

#### Scenario: Long content of details
- **WHEN** the content of a `details` is a paragraph longer than the output width and it is rendered at width 40
- **THEN** every line of the paragraph starts with 2 spaces and no line is wider than 40 columns

#### Scenario: Details marker in a terminal
- **WHEN** a document contains `<details><summary>More</summary>text</details>` and styling is enabled
- **THEN** `▾` has the marker style of the theme and `More` has no style

#### Scenario: Inline formatting tags
- **WHEN** a paragraph contains `Press <kbd>Ctrl</kbd>+<kbd>C</kbd>, <b>bold</b> and <em>italic</em>` and styling is enabled
- **THEN** `Ctrl` and `C` have the inline code style, `bold` is bold and `italic` is italic, and with styling disabled the output is `Press Ctrl+C, bold and italic`

#### Scenario: Link in inline HTML
- **WHEN** a paragraph contains `See <a href="https://example.com">the *guide*</a>.` and hyperlinks are disabled
- **THEN** the output is `See the guide (https://example.com).`

#### Scenario: Relative image in a file
- **WHEN** the file `/home/u/proj/README.md` contains `<img src="img/arch.png" alt="arch">`, it is rendered with hyperlinks enabled, and the host name of the machine is `box`
- **THEN** the visible output is `[image: arch]` and its hyperlink is `file://box/home/u/proj/img/arch.png`

#### Scenario: Removed tags
- **WHEN** a document contains `<picture><source srcset="dark.png"><img src="light.png" alt="logo"></picture>` and the paragraph `x<sup>2</sup> <span class="note">text</span>`, and hyperlinks are disabled
- **THEN** the output contains `[image: logo] (light.png)` and `x2 text`

#### Scenario: Unknown tag in text
- **WHEN** a paragraph contains `Connect with ssh user@<host>.`
- **THEN** the output is `Connect with ssh user@.`

#### Scenario: Script, style and comments
- **WHEN** a document contains `A`, a blank line, `<!-- note -->`, a blank line, `<script>alert(1)</script>`, a blank line, `<style>p {}</style>`, a blank line and `B`
- **THEN** the output is `A`, a blank line and `B`

#### Scenario: Character references
- **WHEN** a document contains `<p>A &amp; B &lt;C&gt;</p>`
- **THEN** the output is `A & B <C>`

#### Scenario: Markdown in an HTML block
- **WHEN** a document contains `<div>**bold**</div>`
- **THEN** the output is `**bold**`

#### Scenario: Unclosed and stray tags
- **WHEN** a document contains the paragraph `<b>bold`, a blank line and the paragraph `next</b> text`, and styling is enabled
- **THEN** `bold` is bold, and `next text` is not bold

#### Scenario: Broken HTML
- **WHEN** a document contains the line `<p align="center`
- **THEN** the output is `<p align="center`
