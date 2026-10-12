# Spec Delta

## MODIFIED Requirements

### Requirement: GFM parsing
termd SHALL parse documents as CommonMark with the GitHub Flavored Markdown extensions for tables, task lists, strikethrough, autolinks and footnotes, and with definition lists.

#### Scenario: GFM extensions are recognized
- **WHEN** a document contains a pipe table, a `- [x]` task item, `~~text~~`, a bare `https://` URL and a footnote reference `[^1]` with the definition `[^1]: Note.`
- **THEN** they are rendered as a table, a checked task item, struck-through text, a link and a footnote respectively

## ADDED Requirements

### Requirement: Definition lists
termd SHALL parse a definition list: a paragraph followed by one or more definitions, each starting with a line that begins with `:` followed by at least one space. Every line of that paragraph SHALL be a term of its own. A definition list can hold several entries, each with its terms and definitions. A line that starts with `:` without a following space SHALL NOT start a definition.

termd SHALL render every term on its own line, wrapped to the output width, in bold when styling is enabled, and the definitions of the terms below them, indented by 4 columns: wrapped lines and further blocks of a definition SHALL be indented by 4 columns as well. Inline formatting, inline code and links in terms and definitions SHALL be rendered as in paragraphs. No theme SHALL color a term.

When no definition of the list is preceded by a blank line in the source and every definition consists of one block, termd SHALL render the list without blank lines. Otherwise termd SHALL separate with a blank line every definition from the previous definition, every term that follows a definition from that definition, and the blocks inside a definition. A term SHALL NOT be separated by a blank line from its first definition or from the next term.

When styling is disabled, the layout SHALL be the same, without bold.

#### Scenario: Term and definition
- **WHEN** the document `Term`, `: Definition of the term` is rendered with styling disabled
- **THEN** the output is `Term` followed by `    Definition of the term`

#### Scenario: Term in bold
- **WHEN** the same document is rendered with styling enabled
- **THEN** `Term` is bold and the definition is not

#### Scenario: Term with a named theme
- **WHEN** the same document is rendered to a terminal with `COLORTERM=truecolor` and `--theme=dracula`
- **THEN** `Term` is bold without a color

#### Scenario: Several terms and definitions
- **WHEN** a document contains the lines `Apple`, `Pomme`, `: A fruit` and `: Red or green`
- **THEN** the output is `Apple`, `Pomme`, `    A fruit` and `    Red or green` on consecutive lines

#### Scenario: Several entries
- **WHEN** a document contains `Term A`, `: Def A`, a blank line, `Term B` and `: Def B`
- **THEN** the output is `Term A`, `    Def A`, `Term B` and `    Def B` on consecutive lines

#### Scenario: Loose list
- **WHEN** a document contains `Term A`, a blank line, `: Def A`, a blank line, `Term B` and `: Def B`
- **THEN** the output is `Term A`, `    Def A`, a blank line, `Term B` and `    Def B`

#### Scenario: Definition with several blocks
- **WHEN** a document contains `Term`, `: First paragraph.`, a blank line and the indented line `    Second paragraph.`
- **THEN** the output is `Term`, `    First paragraph.`, a blank line and `    Second paragraph.`

#### Scenario: Long definition
- **WHEN** a definition longer than the output width is rendered at width 40
- **THEN** it wraps at whitespace, every line of it starts with 4 spaces, and no line is wider than 40 columns

#### Scenario: Inline content
- **WHEN** a document contains `` `--width` `` followed by `: See [docs](https://example.com).` and is rendered with styling and hyperlinks disabled
- **THEN** the output is `--width` followed by `    See docs (https://example.com).`

#### Scenario: Paragraph lines become terms
- **WHEN** a document contains the lines `First line`, `second line` and `: Definition`
- **THEN** the output is `First line`, `second line` and `    Definition` on consecutive lines

#### Scenario: Colon without a space
- **WHEN** a document contains `Term` followed by `:not a definition`
- **THEN** the output is the paragraph `Term :not a definition`

#### Scenario: Definition list in a list item
- **WHEN** a list item `- item` contains, after a blank line, the indented lines `  Term` and `  : Def`
- **THEN** `Term` starts at the column of the item text and `Def` starts 4 columns to the right of it
