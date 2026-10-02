# Proposal

## Why

termd writes the text of a document to the terminal byte for byte, including control characters. A document with `hello \033]0;pwned\007` changes the title of the terminal window, and other sequences can write the clipboard (OSC 52) or redraw the screen. In a pipe the same bytes reach the output, although the `cli` spec promises plain text without escape sequences. The `\r` of a CRLF document also reaches the output inside code blocks.

## What Changes

- Every C0 control character (U+0000 to U+001F) except tab and line feed, and DEL (U+007F), in the text of a document is shown as its Unicode control picture: ESC as `␛` (U+241B), BEL as `␇` (U+2407), NUL as `␀` (U+2400), DEL as `␡` (U+2421). A control picture takes one column.
- C1 control characters (U+0080 to U+009F) have no control pictures and are shown as `�` (U+FFFD), as are bytes that are not valid UTF-8.
- CRLF and a lone CR are line endings, as in CommonMark, and are treated as line feeds: a CRLF document renders like an LF one, including code blocks.
- This applies to all text that comes from the document: paragraphs, headings, code blocks with and without highlighting, tables, link text and destinations, image alt text, raw HTML, frontmatter and diagrams, including framed diagram source. A link destination can no longer end or extend the OSC 8 sequence termd writes around it.
- The escape sequences termd emits itself (styles and OSC 8 hyperlinks) do not change.

Out of scope:

- Other invisible characters that are not control characters, such as bidirectional overrides (U+202A to U+202E) and zero-width characters.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `markdown-rendering`: a new requirement for control characters in the text of a document.

## Impact

- `internal/render`: the source is cleaned once before frontmatter detection and parsing, so the layout measures the control pictures and every block, including highlighted code and diagrams, gets the cleaned text. `text.Line.Render`, where span text and link destinations are written, is the alternative place; the design picks one.
- New golden fixture with control characters, CRLF and a lone CR.
- `README.md`: the behavior in "How it works".
