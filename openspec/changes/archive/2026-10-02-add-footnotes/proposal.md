# Proposal

## Why

GitHub renders footnotes, termd does not parse them. A reference `[^1]` stays as text, and when the text of the definition `[^1]: Note.` happens to be a valid link destination, the definition becomes a link reference definition: `Text[^1].` renders as `Text^1 (Note.).`, and with hyperlinks enabled `^1` is a clickable link to `Note.`.

## What Changes

- termd parses GFM footnotes.
- A footnote reference is rendered as `[1]`, numbered by the order of the first reference rather than by its label.
- The footnote definitions are rendered at the end of the document, after a horizontal line, as a numbered list in the order of their numbers. A definition with several paragraphs is laid out like a list item.
- A definition that is never referenced is not shown.
- The reference marker is not a hyperlink: a terminal hyperlink cannot jump to another place of the output.

Out of scope:

- Back-references from a definition to its reference.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `markdown-rendering`: GFM parsing includes footnotes, and a new requirement describes how references and definitions are rendered.

## Impact

- `internal/render`: `Parse` adds the goldmark footnote extension; the renderer handles the reference and the definition list nodes.
- New golden fixture with footnotes; a test for the current bug case `[^1]: Note.`.
- `README.md`: footnotes in "How it works".
