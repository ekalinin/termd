# table-rendering Specification

## Purpose

Defines how termd lays out GFM tables so that they stay readable and aligned at any output width: column width allocation, wrapping inside cells, alignment and behavior when a table cannot fit.

## Requirements

### Requirement: Table structure
termd SHALL render a table as rows of cells separated by vertical column separators, with a horizontal rule between the header row and the body. Every line of the table SHALL have its column separators at the same columns.

#### Scenario: Aligned separators
- **WHEN** a table with 3 columns and 4 body rows is rendered
- **THEN** the column separators are at identical display columns on every line of the table, including the header and the rule

### Requirement: Words are never split
termd SHALL break cell text only at whitespace. The minimum width of a column SHALL be the display width of the widest word in that column, including its header.

#### Scenario: Short word in a narrow table
- **WHEN** a column contains the value `string` and the table has to be narrowed to fit the output width
- **THEN** `string` is rendered on one line and never as `strin` + `g`

#### Scenario: Minimum width from the header
- **WHEN** a column header is `По умолчанию` and its values are short
- **THEN** the column is at least 9 columns wide (the width of `умолчанию`)

### Requirement: Headers are never truncated
termd SHALL render the full text of every header cell. Header text SHALL wrap only at whitespace and SHALL NOT be shortened or replaced with an ellipsis.

#### Scenario: Long header in a narrow table
- **WHEN** a table with the header `По умолчанию` is rendered at a width where that column cannot be 12 columns wide
- **THEN** the header is rendered as `По` and `умолчанию` on two lines, with no `…`

### Requirement: Column width allocation
When the sum of the column minimum widths plus separators fits the output width, termd SHALL fit the whole table within the output width. Columns whose full content fits on one line SHALL get their full natural width before columns with longer text, and the remaining width SHALL go to the columns with longer text, where wrapping happens.

#### Scenario: Parameters table at width 60
- **WHEN** a table with columns `Параметр`, `Тип`, `По умолчанию`, `Описание` holds rows `--width | int | 80 | <long description>` and `--theme | string | auto | <long description>` and the output width is 60
- **THEN** the table is at most 60 columns wide, `--width`, `--theme`, `string`, `auto` and the header `По умолчанию` are each on one line, and only the `Описание` column wraps

#### Scenario: Table that fits naturally
- **WHEN** a table's natural width is smaller than the output width
- **THEN** it is rendered at its natural width with no wrapping in any cell

### Requirement: Column alignment
termd SHALL apply the alignment from the delimiter row (`:---`, `:---:`, `---:`) to every line of every cell in the column, including wrapped lines. Columns without an alignment marker SHALL be left-aligned.

#### Scenario: Right-aligned numbers
- **WHEN** a column is declared with `---:` and holds `1`, `42` and `1000`
- **THEN** the numbers are right-aligned so their last digits share a column

#### Scenario: Centered column
- **WHEN** a column is declared with `:---:`
- **THEN** each cell's text is centered within the column width

### Requirement: Inline content in cells
termd SHALL render inline formatting, inline code, links and escaped pipes (`\|`) inside cells using the same rules as in paragraphs, and SHALL size cells by the visible width of their rendered content. When hyperlinks are disabled, links and images in cells SHALL follow "Links in cells without hyperlinks" instead of the `text (url)` form of paragraphs.

#### Scenario: Link in a cell
- **WHEN** a cell contains `[docs](https://example.com/very/long/path/to/documentation/page)` and hyperlinks are enabled
- **THEN** the cell shows `docs`, is sized as 4 columns and the link opens the full URL

#### Scenario: Escaped pipe
- **WHEN** a cell contains `` `code \| pipe` ``
- **THEN** the cell shows `code | pipe` and the row keeps its number of columns

### Requirement: Links in cells without hyperlinks
When hyperlinks are disabled, termd SHALL show a link or an image in a table cell, including a link or an image from HTML, as its text followed by a space and a reference marker `[N]`, where `N` is a number, and SHALL NOT write its URL in the cell. Below the table, after a blank line, termd SHALL list the URL of every marker on a line of its own as `[N] url`, ordered by `N`. Every table SHALL number its markers from 1, in the order in which their links appear: the header row first, then the body rows from top to bottom, and the cells of a row from left to right. Links and images with the same URL in one table SHALL share one marker, and the URL SHALL be listed once.

The URL in the list SHALL be the URL that termd writes in `text (url)` outside tables, in full. The lines of the list SHALL NOT be wrapped; when a line of the list is wider than the output width, the table SHALL be emitted as a wide block. The width of a cell SHALL include the marker and SHALL NOT include the URL.

A bare URL in a cell, which is an autolink or a link whose text is its URL, SHALL be shown as a short text followed by the marker. For a URL of an issue, a pull request, a discussion or a commit on GitHub, the short text SHALL be the text that termd shows for that URL with hyperlinks enabled, such as `owner/repo#N`. For any other URL with a host, the short text SHALL be the host of the URL as written, with the port, if any. A bare URL without a host, such as an email address or the destination of `[a.md](a.md)`, and a link with an empty destination SHALL be shown as in a paragraph, without a marker and without a line in the list.

When hyperlinks are enabled, and when no cell of a table has a marker, termd SHALL render the table without markers and without a list.

#### Scenario: Link in a pipe
- **WHEN** a table has the header `Ссылка` and the cell `[docs](https://example.com/very/long/path/to/documentation/page)`, and hyperlinks are disabled
- **THEN** the cell shows `docs [1]`, and the table is followed by a blank line and the line `[1] https://example.com/very/long/path/to/documentation/page`

#### Scenario: Table with a long URL fits the output width
- **WHEN** a table with the columns `Статус`, `Имя`, `Ссылка` and `Число` holds the rows `✅ | **жирный** | [docs](https://example.com/very/long/path/to/documentation/page) | 1` and `⚠️ | 日本語テキスト | x | 1000`, and it is rendered at an output width of 40 with hyperlinks disabled
- **THEN** every line of the table is at most 40 columns wide, `docs` is on one line, and the line `[1] https://example.com/very/long/path/to/documentation/page` follows the table in full as a wide block

#### Scenario: Several links in a table
- **WHEN** a table has the header `[Name](https://n.example) | Logo` and the rows `[a](https://a.example) | ![logo](logo.png)` and `[b](https://a.example) | x`, and hyperlinks are disabled
- **THEN** the cells show `Name [1]`, `a [2]`, `[image: logo] [3]` and `b [2]`, and the list below the table is `[1] https://n.example`, `[2] https://a.example` and `[3] logo.png`

#### Scenario: Markers start from 1 in every table
- **WHEN** a document has two tables with one link each, and hyperlinks are disabled
- **THEN** the cell of each table shows the marker `[1]`, and each table is followed by its own list

#### Scenario: GitHub autolink in a pipe
- **WHEN** a cell contains the bare URL `https://github.com/ekalinin/termd/pull/30` and hyperlinks are disabled
- **THEN** the cell shows `ekalinin/termd#30 [1]` and the list has the line `[1] https://github.com/ekalinin/termd/pull/30`

#### Scenario: Other bare URL in a pipe
- **WHEN** a cell contains the bare URL `https://example.com/docs/mindmaps`, or the link `[https://example.com/docs/mindmaps](https://example.com/docs/mindmaps)`, and hyperlinks are disabled
- **THEN** the cell shows `example.com [1]` and the list has the line `[1] https://example.com/docs/mindmaps`

#### Scenario: Bare URL without a host
- **WHEN** the cells of a table contain `<me@example.com>`, `[a.md](a.md)` and `[empty]()`, and hyperlinks are disabled
- **THEN** the cells show `me@example.com`, `a.md` and `empty`, and the table has no markers and no list

#### Scenario: HTML link in a pipe
- **WHEN** a cell contains `<a href="https://example.com/guide">guide</a>` and hyperlinks are disabled
- **THEN** the cell shows `guide [1]` and the list has the line `[1] https://example.com/guide`

#### Scenario: Hyperlinks enabled
- **WHEN** the tables of the other scenarios are rendered with hyperlinks enabled
- **THEN** no cell has a marker, no list follows a table, and every link with a destination is a hyperlink to its URL

### Requirement: Table overflow
When the sum of the column minimum widths plus separators exceeds the output width, termd SHALL render every column at its minimum width and emit the table as a wide block, without truncating cells and without splitting words.

#### Scenario: Table wider than the terminal
- **WHEN** a table with 12 columns of long words is rendered at an output width of 80 and its minimum widths add up to 140
- **THEN** the table is emitted 140 columns wide, every word is intact, and in the pager it can be scrolled horizontally

### Requirement: Ragged rows
termd SHALL render the number of columns defined by the header row. Missing cells SHALL be rendered empty and extra cells SHALL be ignored.

#### Scenario: Row with fewer cells
- **WHEN** a body row has 2 cells in a 3-column table
- **THEN** the third cell of that row is rendered empty and the separators stay aligned
