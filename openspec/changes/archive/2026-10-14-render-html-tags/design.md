# Design

## Context

`renderer.block` shows an `ast.HTMLBlock` with `verbatim`, line by line without wrapping, and `renderer.inline` shows an `ast.RawHTML` as its source text in the surrounding style. An `ast.Image` becomes the span `[image: alt]` passed through `text.LinkSpans`, as a link. `text.Wrap` already collapses whitespace, turns a `\n` inside a span into a space and breaks the line at a `text.Break` span; table cells are laid out with `text.Wrap`, and `text.NaturalWidth` takes the widest line between breaks.

A spike with goldmark v1.8.6 and `golang.org/x/net/html` v0.61.0 showed how goldmark gives HTML to the renderer:

- Each part of a README header that is separated by blank lines is one `HTMLBlock` (type 6 for `<p>`, `<div>`, `<h1>`, `<details>`): `<p align="center">`, an `<img>` line and `</p>` without blank lines between them are one block. Markdown inside an HTML block is not parsed.
- `<details>`, `<summary>Install</summary>`, a blank line, a paragraph, a code block, a blank line and `</details>` give the siblings `HTMLBlock`, `Paragraph`, `FencedCodeBlock`, `HTMLBlock`. `<div align="center">` around a markdown heading gives the same pattern.
- A line that holds only a complete tag, such as `<img src="a.png">`, starts an `HTMLBlock` of type 7, which takes the following lines up to a blank line, so `<img src="a.png">` followed by `text after` is one HTML block.
- Inline HTML is one `RawHTML` node per tag or comment, between `Text` nodes: `Press <kbd>Ctrl</kbd>` is `Text`, `RawHTML <kbd>`, `Text Ctrl`, `RawHTML </kbd>`. The same holds in table cells (`one<br>two`) and around markdown inlines (`<a href="x.md">the *guide*</a>` holds an `Emphasis`).
- `<b>` in one paragraph and `</b>` in the next give two paragraphs with one `RawHTML` each. A placeholder such as `<host>` in text is a `RawHTML`.
- The tokenizer of `x/net/html` decodes character references in text and in attribute values, reads the content of `script` and `style` as raw text, and returns text with a `<` that does not start a tag (`a < b`) as text. A tag that is not closed at the end of the input, such as `<p align="center`, is dropped without a token. The bytes it consumed are the sum of `len(z.Raw())` over the tokens, so the dropped rest can be found.

The requirements are in `specs/markdown-rendering/spec.md` and `specs/cli/spec.md`.

## Goals / Non-Goals

**Goals:**
- One table of tags for inline HTML and HTML blocks, so a tag means the same in both.
- With `--html=source`, the output for HTML is byte for byte the output of today.
- Documents without HTML and without images that lack alt text render as before. The golden files that change are those of `testdata/code.md` (`<details>`), `testdata/inline.md` (`<kbd>`) and `testdata/control.md` (`<span>`, `<div>`), and the new `testdata/html.md`.

**Non-Goals:**
- A DOM of the document. Tags are matched only as far as the spec requires: inside one paragraph, heading or table cell, inside one HTML block, and `details` across the blocks of one container.
- `p`, `div` or inline formatting tags that span markdown blocks. Only `details` changes the layout of the blocks between its tags; the other tags are boundaries or end with their block.
- An HTML example on the landing page. The site gets the feature through `render.Render`, and its examples contain no HTML.

## Decisions

### The tokenizer of `golang.org/x/net/html`

HTML is read with `html.NewTokenizer`, token by token. It handles attributes with and without quotes, the named and numeric character references of HTML, comments and the raw text of `script` and `style`. `golang.org/x/net` is maintained with the Go project, as `golang.org/x/sys` and `golang.org/x/term`, which termd already uses.

Alternatives:
- `html.Parse`. Rejected: it builds a whole document, adds `html`, `head` and `body`, and moves elements by the tree construction rules of HTML5, while goldmark gives fragments that must be matched with markdown blocks between them.
- A tokenizer of our own. Rejected: the named character references alone are more than 2000 entries, and quoting rules and raw text add more edge cases than the rest of the change.

The text and the attribute values of tokens are passed through `controlPicture`, because the tokenizer decodes `&#27;` into ESC. When the tokenizer ends before the end of its input, the rest is shown as text.

### The HTML mode in `Options`

`render.Options` gets `HTMLSource bool`. The zero value is the default mode `render`, so `internal/site` and the tests do not change. With `HTMLSource`, `renderer.block` and `renderer.inline` take the current paths for `HTMLBlock` and `RawHTML`, and `renderer.blocks` does not group `details`. `cmd/termd/main.go` validates `--html` like `--hyperlinks`, with the message `--html must be render or source, got "strip"`.

Alternative: a string field with the flag value. Rejected: the renderer would have to validate it again; the flag is validated once in `main.go`.

### Inline HTML: open tags per inline container

`renderer.inlines` keeps the open tags while it walks the children of one node, and a `RawHTML` child changes them instead of producing text:

- `b`, `strong`, `i`, `em`, `kbd` and `code` change the style of the following children: bold, italic, or `style.Layer` with `pal.InlineCode`.
- `a` with `href` starts collecting the spans of the following children; its end tag, or the end of the node, passes them to `text.LinkSpans` with the destination.
- `img` gives the spans of an image, `br` gives `text.Break`, and `p`, `div`, `h1` to `h6`, `details` and `summary` give `text.Break`.
- An end tag closes the innermost open tag of its name and every tag opened after it; an end tag without an open tag of its name is ignored. Tags still open at the end of the node end there.

The state belongs to one call of `inlines`, so a tag opened inside an `Emphasis` ends with it. This covers paragraphs, headings, table cells and terms of definition lists, which all go through `inlines`. `renderer.alert` walks the children of the first paragraph of an alert itself and gets the same state.

Alternative: one state for the whole document. Rejected: a tag that is not closed would style the rest of the document, and the spec ends it with its paragraph.

### HTML blocks: tokens to parts

An `HTMLBlock` with its closure line is tokenized into parts: paragraphs and headings as spans, and the summary and the content of a `details` that starts and ends in the same block. Text and inline tags fill the current part with the same tag table as inline HTML; `p`, `div` and the end of a heading or summary close it. Each part is wrapped with `text.Wrap`, a heading with the `#` prefix and `pal.Heading` as in `renderer.block`; the parts are joined with blank lines, empty parts are dropped, and a block without parts returns `false`, so it adds no blank line.

### `details` across blocks: grouping in `renderer.blocks`

`renderer.blocks` walks the children of a container and keeps a stack of open `details`. An `HTMLBlock` that leaves a `details` open pushes it with its summary; the following siblings are rendered at the width minus 2 for every open `details` and become its content, up to the `HTMLBlock` that closes it. A `details` that is still open at the end of the container is closed there.

A closed `details` becomes one block:

- `▾ ` in `pal.Marker` and the summary wrapped at the width minus 2, its further lines indented by 2, as `renderer.items` does with the lines of an item. Without a summary, the text is `Details`.
- The content blocks joined with blank lines, every non-empty line prefixed with 2 spaces, as `renderer.definitions` does with 4.
- `Wide` when a block of the content is wide.

A nested `details` is a content block of the outer one and is indented again, so its marker is at column 2 and its content at column 4.

Alternatives:
- `▸` or no marker. Decided with the user: the content is always shown, and `▾` is the marker of an open `details`; without the indent the end of a `details` that spans blocks is not visible.
- A blank line between the summary and the content. Rejected: an alert title is followed by its text without a blank line, and the indent already separates the content.

### Images without alt text

The `[image: alt]` label moves into a helper used by `ast.Image` and by `img`: an alt text that is empty or only whitespace gives `[image]`. A missing `src` gives an empty destination, which `text.LinkSpans` shows without a link and without ` ()`.

### Destinations from HTML

The value of `href` and `src` is already decoded by the tokenizer, so it goes to `resolveLink` without `unescape`, which would decode `&amp;amp;` a second time. Apart from that it follows `renderer.destination`: resolved against `Options.Dir` with hyperlinks, as written without them.

## Risks / Trade-offs

- [goldmark reads a placeholder in text, such as `ssh user@<host>`, as a tag, and the mode `render` removes it] → GitHub removes it too. Such placeholders outside code are rare (about 10 in 2058 READMEs); `--html=source` keeps them, and README describes the behavior. A scenario of the spec makes it explicit.
- [Tags outside the table lose their layout: `<hr>` gives nothing (5 of 2058 READMEs), the whitespace of `<pre>` collapses (3), the items of `<ul>`/`<li>` run together (2), HTML tables lose their columns (5, out of scope)] → Listed in README "Limitations" with `--html=source` as the workaround. A row of the tag table can be added later without changing the design.
- [A `details` that is never closed indents the rest of its container] → The same as in a browser, and limited to the container: a list item or a block quote ends it.
- [`▾` is a character of ambiguous East Asian width, which some terminals with CJK settings show in 2 columns] → The same as `•` and `│`, which termd already uses as markers.
- [In inline HTML, `p` and `div` can only break the line, not start a paragraph with a blank line] → A block tag in the middle of a line is rare; at the start of a line it starts an HTML block, which has real paragraphs.
- [goldmark ends an HTML block at a blank line, so `<a href="x">`, a blank line, a markdown image, a blank line and `</a>` give no link] → Only `details` spans blocks, as the spec says; the image is still shown.

## Migration Plan

The golden files of `testdata/code.md`, `testdata/inline.md` and `testdata/control.md` change at every width and mode: `<details>`, `<kbd>`, `<span>` and `<div>` are interpreted. `--html=source` gives the previous output for HTML. Rollback is a revert of the change.
