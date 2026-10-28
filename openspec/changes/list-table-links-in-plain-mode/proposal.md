# Proposal

## Why

When hyperlinks are disabled, a link in a table cell is written as `text (url)`. A cell never splits a word, so a long URL becomes the minimum width of its column and pushes the table beyond the output width. The README lists this as a limitation: "In plain mode a long URL is an unbreakable word and can push a table beyond the output width". glamour avoids it by printing the links of a table as a numbered list below the table.

## What Changes

- When hyperlinks are disabled, a link or an image in a table cell, also from HTML, shows its text with a reference marker `[1]`, `[2]`, ... instead of `text (url)`, and the URLs are listed below the table after a blank line, one per line as `[1] url`, in the order of their markers. The markers start from 1 in every table.
- A bare URL in a cell (an autolink, or a link whose text is its URL) shows a short text with the marker: `owner/repo#N` and the other GitHub texts that termd shows with hyperlinks, and the host of the URL for any other URL. A bare URL without a host, such as an email address, stays in the cell without a marker.
- The cell is sized by its text and marker only, so a table with long URLs fits the output width when its words fit.
- URLs in the list are never truncated; a line of the list wider than the output is a wide block, as today.
- Two links to the same URL in one table share one marker.
- When hyperlinks are enabled, nothing changes: the cell shows only the text and the URL is in the hyperlink.
- Links outside tables do not change.

Out of scope:

- An option that keeps `text (url)` in cells.
- Truncating URLs with an ellipsis, as glamour does in the list.
- Markers that differ from footnote references. Both are shown as `[1]`, so in a document with footnotes and links in tables they look the same.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `table-rendering`: the "Inline content in cells" requirement refers to a new requirement, "Links in cells without hyperlinks", which moves the URLs of links and images below the table when hyperlinks are disabled.
- `markdown-rendering`: the "Hyperlinks" requirement, which writes a link as `text (url)` when hyperlinks are disabled, refers to the table exception.

## Impact

- `internal/render`: the table renders link spans of cells without the URL and collects the URLs for the list.
- `internal/text`, `internal/table`: no change; the list is added after the lines of the table.
- Golden files of fixtures with links in tables, in plain mode: `tables`, `regression` and `autolinks`.
- `README.md`: the "Links" section; the limitation is removed from "Limitations".
