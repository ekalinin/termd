# Design

## Context

`render.Parse` builds the goldmark parser with `extension.GFM`, which bundles tables, strikethrough, autolinks and task lists but not footnotes. `render.Render` cuts off the frontmatter, parses the body and lays out the document blocks; `cmd/termd` and the site generator both go through it. Ordered lists are laid out by `renderer.list`, thematic breaks by the `ast.ThematicBreak` case of `renderer.block`.

goldmark v1.8.6 ships `extension.Footnote` with the nodes in `extension/ast`. A spike with the bug case and the edge cases of the spec showed how it builds the AST:

- A definition `[^label]:` is a `Footnote` block. All definitions, also those inside block quotes and list items, are moved into one `FootnoteList`, which the AST transformer appends as the last child of the document.
- A reference is a `FootnoteLink` inline. Its `Index` is assigned when a label is first referenced during inline parsing, so the numbers follow the first reference, not the label. Every reference to the same footnote has the same `Index`; `RefCount` and `RefIndex` only serve back-references.
- The transformer removes the definitions without a reference, sorts the rest by `Index` (so the numbers are 1, 2, 3 without gaps) and removes the list when nothing is left.
- It appends one `FootnoteBacklink` per reference to the last paragraph of a definition, or to the `Footnote` itself when its last child is not a paragraph or it has no children.
- A reference to a label without a definition stays text: `Missing[^x].` gives the text `Missing[^x].`.
- Labels are compared byte by byte, so `[^Note]` does not match `[^note]:`.

The requirements are in `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- The renderer only renders the footnote nodes; numbering, order and dropping unreferenced definitions come from goldmark.
- The definition list reuses the list layout and the thematic break line, so it looks like the rest of the document.
- Documents without footnotes render byte for byte as before; no current fixture contains `[^`.

**Non-Goals:**
- A heading such as "Footnotes" above the definitions; the proposal asks for a horizontal line only.
- A footnote example on the landing page. The site gets the feature through `render.Render`, its examples stay as they are.
- Changing the goldmark numbering in the edge cases listed under risks.

## Decisions

### Enable `extension.Footnote` in `Parse`

`Parse` becomes `goldmark.WithExtensions(extension.GFM, extension.Footnote)`. The footnote block parser runs before the paragraph parser and the link reference definition transformer, so `[^1]: Note.` is no longer a link reference definition; the spike confirmed `Text[^1].` then parses as text, a `FootnoteLink` and text.

Alternative: a parser of our own for references and definitions. Rejected: goldmark already does the numbering and the cleanup the spec asks for, and a second parser would have to agree with goldmark on where blocks end.

### Reference marker

A `FootnoteLink` renders as one span `[N]`, with `N` its `Index`, in the style of the surrounding text, so a reference inside bold text is bold. It does not go through `text.LinkSpans`, so it is never a hyperlink.

Alternative: the faint marker style used for list bullets. Rejected: the marker is part of the sentence the reader follows, and faint text is hard to read on some themes.

### Back-references

`FootnoteBacklink` renders nothing. An explicit case in `renderer.inline` returns no spans. A backlink appended to the `Footnote` itself is a block child without children, which `renderer.blocks` already skips like any other such node.

### Definition list

A `FootnoteList` case in `renderer.block` returns one block: the horizontal line, a blank line and the list of definitions. The line is the one of a thematic break, a row of `─` as wide as the output in the faint marker style; it moves into a small helper used by both cases. The blank line is what a thematic break followed by a list gets today, since document blocks are joined with a blank line. The list is always a child of the document, so it gets the full output width.

`renderer.list` is split in two: `list` computes the markers of an `ast.List` as before, and a new `items(parent, markers, tight, width)` lays out the children of any node as items, with the right-aligned markers, the indentation of wrapped lines and further blocks, and the blank lines of a loose list. `list` calls it with `n.IsTight`; the footnote case calls it with the markers `1.`, `2.`, ... taken from `Footnote.Index`. A definition is thus laid out exactly like an ordered list item, including definitions with ten or more entries, where the markers are right-aligned.

The markers are `1.` rather than `[1]`: the proposal describes a numbered list, and this is how termd renders numbered lists; the numbers are the same as in the references, so they are easy to match.

Alternatives:
- Building an `ast.List` with `ListItem` nodes from the definitions and passing it to `list`. Rejected: it moves nodes between parents in the AST, only to reuse the layout.
- A separate layout for the definitions. Rejected: it would duplicate the list layout that the spec refers to.

### Tight or loose definition list

The list is tight when every definition has at most one block, not counting backlinks, and loose otherwise: then a blank line separates the definitions and the blocks inside them. This is the CommonMark rule for a list written with the same content.

Alternatives:
- Always loose, as the HTML of GitHub, where every definition is a paragraph. Rejected: a blank line between short one-line notes doubles the height of the list.
- Always tight. Rejected: the paragraphs of a definition with several paragraphs would run together.
- Loose when the definitions are separated by blank lines in the source. Rejected: definitions often stand after the paragraphs that reference them, scattered through the document, so the blank lines around them say nothing about the list.

### Frontmatter

No change. `splitFrontmatter` runs before `Parse`, and the renderer reads the segments of the body, so the footnote nodes point into the right slice. A test covers a document with both.

## Risks / Trade-offs

- [goldmark numbers the references inside definitions as if all definitions stood where the first definition is: in `[^a]: x`, `P1[^b]`, `[^b]: See[^a]`, `P2[^a]` the footnote `a` gets 1 although `[^b]` is referenced first] → Accepted: it needs a reference inside a definition and a definition before the first reference. The references and the list still show the same numbers.
- [goldmark compares labels with their case, GitHub ignores it] → `[^Note]` with `[^note]:` stays text and the definition is not shown. Accepted as rare; it can be fixed upstream.
- [A document that uses `[^x]: url` as a link reference definition renders differently] → GitHub treats it as a footnote too.

## Migration Plan

None. Documents without footnotes render as before, so the existing golden files do not change. Rollback is a revert of the change.
