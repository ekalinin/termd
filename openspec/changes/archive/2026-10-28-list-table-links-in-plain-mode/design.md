# Design

## Context

`renderer.table` in `internal/render/render.go` turns every cell into spans with `r.inlines` and passes them to `table.Table`, which measures a cell by the visible width of its spans: the minimum width is the widest word, and a cell never splits a word.

Four places build the spans of a link, all with `text.LinkSpans(label, url, hyperlinks)`: the `ast.Link` and `ast.AutoLink` cases of `renderer.inline`, `renderer.image`, which also renders `<img>`, and `inlineHTML.pop` in `internal/render/html.go` for `<a href>`. Without hyperlinks, `LinkSpans` appends the span ` (url)` to the label, unless the label reads as the URL; for an email autolink, the `ast.AutoLink` case passes the label as the URL, so the address is shown once. `renderer.destination` returns a destination as written when hyperlinks are disabled. For a `www.` autolink, the label is `www.example.com/x` and the URL is `http://www.example.com/x`.

goldmark gives a table the header row as its first child, so the cells are rendered in reading order. GFM tables cannot nest.

Three fixtures have links in tables: `testdata/tables.md` and `testdata/regression.md` (`[docs](https://example.com/very/long/path/to/documentation/page)`) and `testdata/autolinks.md` (two bare URLs of pull requests). The golden tests render plain output without hyperlinks and styled output with them; the site and the theme golden files render with hyperlinks.

The requirements are in `specs/table-rendering/spec.md` and `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- Output with hyperlinks and output of documents without links in tables stay byte for byte as before. Only the plain golden files of the three fixtures above change.
- `internal/text` and `internal/table` do not change.

**Non-Goals:**
- Joining the marker to the last word of the text. The space before the marker is a break opportunity like any other, so in a narrow column the marker can go to the next line.
- Styling the markers and the list. They are written without a style, like the ` (url)` span today.

## Decisions

### The table collects the URLs on the renderer

`renderer` gets a field `refs *linkRefs`. `linkRefs` holds the URLs of the markers in order, and its method `marker(url string) int` returns the number of a URL, adding it when it is new. `renderer.table` sets `r.refs` to a new `linkRefs` before it renders the cells when hyperlinks are disabled, and sets it back to `nil` after them. Outside a table, `r.refs` is `nil`.

Alternatives:
- Remove the ` (url)` spans from the cells after rendering. Rejected: the span cannot be told from text such as `(x)`, and `LinkSpans` leaves it out when the label reads as the URL.
- An option of `text.LinkSpans` that leaves out the URL, as the proposal first planned. Rejected: the numbering needs state per table, which `internal/text` does not have, and the four call sites would still need the state from the renderer.

### One method for the spans of a link

`renderer.link(label []text.Span, url string, bare bool) []text.Span` takes the place of the four calls of `text.LinkSpans`. When `r.refs` is `nil`, which also covers hyperlinks, it returns `text.LinkSpans(label, url, hyperlinks)`, so output outside tables does not change. In a cell without hyperlinks:

- An empty URL returns the label, as `LinkSpans` does.
- A bare link, which is `bare` or has a label that reads as the URL, gets a short text from `shortText(url)`: `githubText(url)` when it matches, else the `Host` of `url.Parse(url)`. The short text takes the style of the first span of the label. Without a short text, the label is returned without a marker; this covers email addresses and relative destinations such as `a.md`.
- Otherwise the label is followed by the span ` [N]` with `N` from `r.refs.marker(url)`.

The `ast.AutoLink` case passes `bare` as `n.AutoLinkType == ast.AutoLinkURL`, so a `www.` autolink, whose label differs from its URL, is bare too. The other call sites pass `false`.

Alternatives:
- Treat only autolinks as bare. Rejected: a link whose text is its URL is already shown as the URL alone without hyperlinks, so in a cell it has the same width problem.

### The list is added to the lines of the table

After `t.Render(width)`, `renderer.table` appends an empty line and then `text.Plain("[N] " + url)` for every URL of `r.refs`, when there is one. The block is wide when the table overflows or a line of the list is wider than `width`. The lines of the list are not wrapped, so a URL is never split.

Alternatives:
- Return the list as a second `Block` and let `join` add the blank line. Rejected: `block` returns one block per node, and a table inside a list item or a quote has to keep its list next to it.

### Tests and golden files

`TestTableLinks` in `internal/render/render_test.go` renders the scenarios of `specs/table-rendering/spec.md` with `plain(80)` and compares whole outputs; the scenario of the output width renders with `plain(40)` and checks the width of the table lines and the `Wide` flag of the block from `renderer.block`. A case with `styled(80)` checks that the tables of the scenarios have no markers and no list. `make golden` updates the plain files of `tables`, `regression` and `autolinks`; the fixtures do not change.

## Risks / Trade-offs

- [The marker `[1]` looks like a footnote reference] → Accepted in the proposal. The list right below the table shows which URL a marker means.
- [A linked image, such as `[![CI](badge.svg)](https://ci.example.com)`, gets two markers, `[image: CI] [1] [2]`] → Today it is written with two URLs, `[image: CI] (badge.svg) (https://ci.example.com)`, so the cell only gets shorter.
- [The host of a bare URL hides its path in the cell] → The full URL is in the list below the table, and with hyperlinks the cell shows the full URL as before.

## Migration Plan

None. Output with hyperlinks does not change. Rollback is a revert of the change.
