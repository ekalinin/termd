## ADDED Requirements

### Requirement: ASCII mode
In ASCII mode, termd SHALL draw the column separators of a table with `|` instead of `│`, and the rule under the header with `-` instead of `─` and `+` instead of `┼` where the rule crosses a column separator, at the same columns and with the same style. This SHALL apply to the tables of the document and to the key-value table of frontmatter. The cells SHALL NOT change.

#### Scenario: Table in ASCII mode
- **WHEN** a table with the header `A | B` and the row `1 | 2` is rendered in ASCII mode with styling disabled
- **THEN** the output is `A | B`, `--+--` and `1 | 2`

#### Scenario: Frontmatter table in ASCII mode
- **WHEN** a document starts with `---`, `title: Doc`, `---` and it is rendered in ASCII mode with styling disabled
- **THEN** the frontmatter table is `title | Doc`
