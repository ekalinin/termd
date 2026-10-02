# Spec Delta

## MODIFIED Requirements

### Requirement: GFM parsing
termd SHALL parse documents as CommonMark with the GitHub Flavored Markdown extensions for tables, task lists, strikethrough, autolinks and footnotes.

#### Scenario: GFM extensions are recognized
- **WHEN** a document contains a pipe table, a `- [x]` task item, `~~text~~`, a bare `https://` URL and a footnote reference `[^1]` with the definition `[^1]: Note.`
- **THEN** they are rendered as a table, a checked task item, struck-through text, a link and a footnote respectively

## ADDED Requirements

### Requirement: Footnotes
termd SHALL render a footnote reference as the number of its footnote in square brackets, for example `[1]`. Footnotes SHALL be numbered from 1 in the order of their first reference, regardless of their labels, and every reference to the same footnote SHALL show the same number. The reference SHALL NOT be a hyperlink, also when hyperlinks are enabled. A reference to a label without a definition SHALL be rendered as its source text.

termd SHALL render the definitions of the referenced footnotes at the end of the document, wherever they stand in the source: a horizontal line spanning the output width, the same as a thematic break, followed by a numbered list of the definitions in the order of their numbers. A definition SHALL be laid out like an item of an ordered list: wrapped lines and further blocks of the definition are indented to the start of the item text. When every definition consists of one block, the definitions SHALL follow each other without blank lines; when a definition has several blocks, the definitions and the blocks inside them SHALL be separated by blank lines, as in a loose list.

A definition that is never referenced SHALL NOT be shown, and when no definition is referenced, termd SHALL NOT render the horizontal line. A footnote definition SHALL NOT be treated as a link reference definition. termd SHALL NOT render back-references from a definition to its references.

#### Scenario: Footnote reference and definition
- **WHEN** the document `Text[^1].`, a blank line, `[^1]: Note.` is rendered with hyperlinks disabled at width 40
- **THEN** the output is `Text[1].`, a blank line, a horizontal line 40 columns wide, a blank line and `1. Note.`, and it contains neither `^1` nor `(Note.)`

#### Scenario: Reference is not a hyperlink
- **WHEN** the same document is rendered with hyperlinks enabled
- **THEN** the output contains `[1]` and no hyperlink escape sequence

#### Scenario: Numbering by first reference
- **WHEN** a document contains the paragraph `A[^b] B[^a] C[^b]` followed by the definitions `[^a]: Alpha.` and `[^b]: Beta.`
- **THEN** the paragraph is rendered as `A[1] B[2] C[1]` and the definitions as `1. Beta.` and `2. Alpha.` on consecutive lines

#### Scenario: Definitions in the middle of the document
- **WHEN** a document contains the paragraph `First[^1].`, the definition `[^1]: Note.` and then the paragraph `Last.`
- **THEN** `Last.` is rendered before the horizontal line and `1. Note.` after it

#### Scenario: Unreferenced definition
- **WHEN** a document contains `Text[^1].` and the definitions `[^1]: Used.` and `[^2]: Unused.`
- **THEN** the definitions are rendered as `1. Used.` only, and the output does not contain `Unused.`

#### Scenario: No referenced definition
- **WHEN** a document contains the paragraph `Text.` and the definition `[^1]: Unused.`
- **THEN** the output is `Text.` without a horizontal line

#### Scenario: Definition with several paragraphs
- **WHEN** a document contains `Text[^1].` and the definition `[^1]: First paragraph.` followed by a blank line and the indented line `    Second paragraph.`
- **THEN** the definitions are rendered as `1. First paragraph.`, a blank line and `   Second paragraph.`, indented to the start of the item text

#### Scenario: Undefined reference
- **WHEN** a document contains `Text[^x].` and no definition of `x`
- **THEN** the output is `Text[^x].` without a horizontal line

#### Scenario: Footnotes with frontmatter
- **WHEN** a document starts with the frontmatter `---`, `title: Doc`, `---` and continues with `Text[^1].` and the definition `[^1]: Note.`
- **THEN** the output starts with the frontmatter table, contains `Text[1].` and ends with `1. Note.`
