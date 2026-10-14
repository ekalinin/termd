# Design

## Context

`render.Parse` builds the goldmark parser with `extension.GFM`, `extension.Footnote` and `extension.DefinitionList`. `renderer.inline` renders the inline nodes it knows and, for any other node, the children of that node; `renderer.plainText` collects the text of `ast.Text` and `ast.String` nodes for the alt text of images and for code spans. HTML blocks are laid out from the tokens of their source and are never parsed as markdown.

goldmark-emoji v1.0.6 has three parts: `definition.Github()`, a map of 1913 GitHub short names to their emoji; the node `ast.Emoji` with the short name and the emoji; and an inline parser that the `emoji.Emoji` extension registers with the priority 999 and the trigger `:`. A spike with goldmark v1.8.6 and the extensions of termd showed:

- The parser reads letters, digits, `_`, `+` and `-` after a colon up to the next colon and looks the name up. It does not look at the character before the first colon, so `a:b:c` gives `a🅱️c` and `foo:rocket:bar` gives `foo🚀bar`.
- An `ast.Emoji` is found in headings, paragraphs, list items, block quotes, table cells, link text, image alt text, footnotes, definition terms and descriptions, and between `RawHTML` nodes (`<kbd>:sparkles:</kbd>`). It is not found in code spans, code blocks, link destinations, HTML blocks, or in a bare URL such as `https://example.com/:smile:/x`, which the linkify parser of GFM takes first.
- `:octocat:` is not in the map, and no entry of the map lacks a Unicode character.
- `ast.Emoji` has no children, so `renderer.inline` and `renderer.plainText` would drop it.

The markdown API of GitHub keeps `a:b:c`, `foo:rocket:bar`, `key:link:value` and `1:zap:` as text and replaces `x :b: y`, `:rocket::zap:` and `:zap:x`. It also keeps `(:rocket:)`, `:rocket:.` and `-:zap:-`, keeps `:zap:` in `x:rocket: :zap:`, keeps shortcodes in the `alt` attribute of an image and replaces them in HTML blocks. In 2058 READMEs of Go modules, all 46 shortcodes follow a space or start a line; 3 are followed by punctuation.

The requirements are in `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- A shortcode right after a letter or a digit stays text; any other GitHub shortcode outside code, destinations, alt text, HTML blocks and frontmatter becomes its emoji.
- Documents without shortcodes render byte for byte as before. No fixture, no example of the site and no README contains a shortcode, so the existing golden files do not change.

**Non-Goals:**
- The other rules of GitHub for shortcodes next to punctuation (`(:rocket:)`, `:rocket:.`) and after a failed shortcode (`x:rocket: :zap:`). They cannot be stated as one rule, and the author of `(:rocket:)` most likely wants the emoji.
- Shortcodes in HTML blocks, see proposal.md.

## Decisions

### The definitions and the node of goldmark-emoji, an inline parser of termd

`internal/render/emoji.go` holds an inline parser with the trigger `:`. It returns no node when `PrecendingCharacter()` of the reader is a letter or a digit (`unicode.IsLetter`, `unicode.IsDigit`), reads the name with the characters of goldmark-emoji, looks it up in `definition.Github()` and returns `ast.NewEmoji` of goldmark-emoji. It advances the reader only when it returns a node. `Parse` adds it with `parser.WithInlineParsers(util.Prioritized(..., 999))`, the priority of the extension, so code spans, links, autolinks, raw HTML and linkify keep their precedence as in the spike.

Alternatives:
- The `emoji.Emoji` extension as it is. Rejected: `a:b:c` in prose becomes `a🅱️c`.
- The extension, with the renderer showing an emoji as its shortcode when the text before it ends with a letter or a digit. Rejected: the renderer would look at sibling nodes to undo a decision of the parser, and the parser can see the character directly.
- A list of GitHub emoji of our own. Rejected: goldmark-emoji keeps the list up to date with GitHub, and termd would have to.

The parser accepts a name of any length, as goldmark-emoji does; a name without a closing colon on the same line is no shortcode.

### Rendering

`renderer.inline` gets an `ast.Emoji` case: a span with the emoji in the style of the surrounding text, so `**:tada:**` is bold and a heading keeps its style. `text.Width` measures it like any other emoji.

`renderer.plainText` writes an `ast.Emoji` as its shortcode, `:` + short name + `:`. It is used for the alt text of images, which keeps the shortcode as on GitHub, and for code spans, which never hold an emoji.

Alternative: the emoji in the alt text, `[image: logo ⭐]`. Rejected: GitHub keeps it in the `alt` attribute, and termd shows the alt text the way GitHub has it.

### No change for HTML blocks

HTML blocks are rendered from the tokens of their source, which the inline parser never sees, so their shortcodes stay text without code for it. Text between inline HTML tags in a paragraph is markdown text and gets emoji.

## Risks / Trade-offs

- [Short names such as `:b:`, `:x:`, `:m:`, `:o:` or `:on:` after a space become emoji, for example `x :b: y`] → GitHub does the same; inside code they stay text.
- [termd replaces `(:rocket:)`, `:rocket:.` and `:necktie:,` where GitHub keeps them] → Accepted, see Non-Goals. 3 of 46 shortcodes in the sample are followed by punctuation.
- [Emoji with a variation selector, such as `:warning:` (⚠️), are 2 columns for termd, and terminals disagree on their width] → The existing limitation of README, which now applies to shortcodes too.
- [The set is the one of goldmark-emoji v1.0.6; GitHub adds emoji over time] → An update of the dependency updates the set.

## Migration Plan

None. Documents without shortcodes render as before, so the existing golden files do not change. Rollback is a revert of the change.
