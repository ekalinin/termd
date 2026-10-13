# Proposal

## Why

HTML in a document is shown as its source. In a sample of 914 READMEs of Go modules, 118 contain HTML, most often in the header: a centered logo (`<p align="center"><img ...></p>`), a row of badges (`<a><img></a>`), `<br>`, `<details>`/`<summary>`. termd prints these lines tag by tag, so the first screen of such a README is markup instead of text. glamour removes all tags and keeps only their text, which also loses the `alt` of images and the `href` of links, the most useful part of this markup.

## What Changes

- termd interprets the most common HTML tags in HTML blocks and in inline HTML instead of showing their source:

  | Tag | Output |
  |---|---|
  | `img` | `[image: alt]` linked to `src`, as a markdown image; `[image]` without `alt` |
  | `a href` | a link around its content, as a markdown link |
  | `p`, `div` | paragraph boundaries; `align` and other attributes are ignored |
  | `br` | a hard line break |
  | `h1` to `h6` | a heading, as a markdown heading |
  | `details`, `summary` | the summary on its own line after the marker `▾`, followed by the content indented by 2 columns, which is always shown |
  | `picture`, `source` | only the `img` inside |
  | `kbd`, `code` | the inline code style |
  | `b`, `strong`, `i`, `em` | bold, italic |
  | `sup`, `sub` | their text |

- Tags not in the table are removed and their text is kept. The content of `script` and `style` is removed. HTML comments (`<!-- -->`) are removed.
- goldmark gives HTML in fragments: an opening and a closing tag can be in different inline nodes (`<kbd>`, `Ctrl`, `</kbd>`) or in different blocks (`<details>` before a markdown paragraph, `</details>` after it). Tags are matched across these fragments; a tag that is never closed ends with its block.
- Links and images from HTML follow the rules of markdown links and images: OSC 8 hyperlinks, `text (url)` when hyperlinks are disabled, relative destinations resolved against the directory of the file.
- Character references in HTML text are decoded, and control characters are shown as control pictures, as in the rest of the document.
- Broken HTML never fails the run: what cannot be parsed as a tag is shown as text.
- In plain mode the output has the same structure without styles.
- A new flag `--html=render|source`, defaulting to `render`. `source` keeps today's output: HTML blocks and inline HTML are shown as their source. Any other value is a usage error with exit status 2.
- A markdown image without alt text, such as `![](logo.png)`, is shown as `[image]` instead of `[image: ]`, the same as an `img` without `alt`. In a sample of 2058 READMEs of Go modules, 125 of 2592 markdown images and 166 of 375 `<img>` tags have no alt text; the `<img>` tags without `alt` are mostly avatars of sponsors, logos and screenshots.

Out of scope:

- HTML tables (`<table>`, `<tr>`, `<td>`): the tags are removed and the text of the cells is kept, without a table layout.
- Centering by `align`, CSS, `width`/`height` of images.
- Collapsing `<details>`: a terminal output cannot be expanded.
- Removing all tags as glamour does: it loses `alt` and `href`.
- An environment variable for the HTML mode.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `cli`: a new requirement for the `--html` flag.
- `markdown-rendering`: the "Images and raw HTML" requirement interprets HTML tags instead of showing their source, unless the HTML mode is `source`, and shows an image without alt text as `[image]`; a new requirement "HTML tags" describes the tags; the scenario with raw HTML in "Control characters in document text" shows the text of the tags.

## Impact

- `cmd/termd/main.go`: the flag, its validation and the usage text.
- `internal/render`: `Options` gets the HTML mode; the inline walk keeps a stack of open HTML tags across sibling nodes; HTML blocks are tokenized into inline content and block boundaries; `<details>` spans several blocks.
- A new dependency for the HTML tokenizer, `golang.org/x/net/html`.
- Tests and a golden fixture `testdata/html.md` with the header patterns of real READMEs; the scenario "Raw HTML" of the spec is replaced. The golden files of `testdata/code.md`, `testdata/inline.md` and `testdata/control.md`, which contain `<details>`, `<kbd>`, `<span>` and `<div>`, change.
- `README.md`: the flags table, a section in "How it works" that lists the tags; "Limitations".
