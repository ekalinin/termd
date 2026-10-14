# Proposal

## Why

GitHub shows emoji shortcodes such as `:rocket:` or `:white_check_mark:` as emoji, and READMEs use them in headings, lists and tables. termd shows them as text, `:rocket:`. glamour replaces them with the `WithEmoji` option, based on the goldmark-emoji extension.

## What Changes

- termd replaces a GitHub emoji shortcode in the text of a document with its emoji: in paragraphs, headings, list items, quotes, alerts, table cells, link text, footnotes and definition lists, including the text between inline HTML tags such as `<kbd>:zap:</kbd>`.
- A shortcode right after a letter or a digit stays as text, so `a:b:c`, `foo:rocket:bar` and `1:zap:` do not change, as on GitHub. goldmark-emoji replaces them (`a🅱️c`). In a sample of 2058 READMEs of Go modules, 28 use shortcodes, 46 in total, and every one of them follows a space or starts a line.
- Shortcodes in inline code, code blocks, diagrams, link destinations, bare URLs, image alt text, HTML blocks and frontmatter are not replaced. GitHub keeps them in image alt text too, which is an attribute there.
- A shortcode that is not a GitHub emoji, or a GitHub emoji without a Unicode character (for example `:octocat:`), stays as text.
- The emoji is part of the text, so it is shown in plain mode too, and tables measure it by its display width like any other emoji.
- Replacement is always on; there is no flag.

Out of scope:

- A flag that turns replacement off.
- Images for custom GitHub emoji.
- Shortcodes in HTML blocks, such as `<p align="center">:rocket: Fast</p>`. GitHub replaces them; no README of the sample has one.
- The rules of GitHub for a shortcode next to punctuation: GitHub keeps `(:rocket:)` and `:rocket:.` as text, termd replaces them.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `markdown-rendering`: a new requirement for emoji shortcodes; the "GFM parsing" requirement lists them.

## Impact

- A new dependency: `github.com/yuin/goldmark-emoji` for its GitHub emoji definitions and its emoji node. Its parser does not check the character before a shortcode, so termd uses an inline parser of its own.
- `internal/render`: the parser gets the inline parser; the inline walk renders the emoji node as text, and the alt text of an image shows the shortcode.
- Some shortcodes give an emoji with a variation selector, for example `:warning:`. termd counts such emoji as 2 columns, and terminals disagree on their width; this is an existing limitation of the README, which now applies to shortcodes too.
- Tests and golden output with shortcodes in a paragraph, a heading, a table and inline code.
- `README.md`: "Features".
