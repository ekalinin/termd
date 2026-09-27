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
termd SHALL render inline formatting, inline code, links and escaped pipes (`\|`) inside cells using the same rules as in paragraphs, and SHALL size cells by the visible width of their rendered content.

#### Scenario: Link in a cell
- **WHEN** a cell contains `[docs](https://example.com/very/long/path/to/documentation/page)` and hyperlinks are enabled
- **THEN** the cell shows `docs`, is sized as 4 columns and the link opens the full URL

#### Scenario: Escaped pipe
- **WHEN** a cell contains `` `code \| pipe` ``
- **THEN** the cell shows `code | pipe` and the row keeps its number of columns

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
