# Spec Delta

## ADDED Requirements

### Requirement: YAML frontmatter
When the first line of a document is `---`, a later line is `---` or `...`, and the lines between them form a YAML mapping, termd SHALL NOT render these lines as markdown. termd SHALL render the mapping as the first block of the document: a table with two columns, the key and its value, with one row per key in source order, without a header row and without a horizontal rule. Keys SHALL be bold when styling is enabled. The table SHALL follow the table layout rules: column separators at the same columns on every line, words never split, the value column wrapping first, and a table that does not fit at its minimum widths emitted as a wide block. A UTF-8 byte order mark before the first line and CRLF line endings SHALL NOT prevent the detection.

termd SHALL show a scalar value as its text without quotes, a list of scalars as its items separated by `, `, a mapping of scalars as one `key: value` line per entry, and a multi-line string with its line breaks. Any other value SHALL be shown as YAML in flow style on one line. A key without a value SHALL have an empty value cell.

When the lines between the delimiters are not valid YAML, termd SHALL show them as source in a frame labelled `frontmatter - invalid YAML`, render the rest of the document normally and exit with status 0. When the lines between the delimiters are empty or contain only YAML comments, termd SHALL omit them from the output. When they are valid YAML but not a mapping, when there is no closing line, or when the `---` line is not the first line of the document, termd SHALL render the document as CommonMark.

#### Scenario: Frontmatter table
- **WHEN** a document starts with `---`, `title: Doc`, `tags: [a, b]`, `---` followed by `# Hello`
- **THEN** the output starts with a row `title` and `Doc` and a row `tags` and `a, b`, separated by the column separator, then a blank line and `# Hello`, and it contains no horizontal line and no `## title` heading

#### Scenario: Nested mapping
- **WHEN** the frontmatter contains `author:` with the indented entries `name: Eugene` and `url: https://example.com`
- **THEN** the value cell of `author` holds `name: Eugene` and `url: https://example.com` on two lines

#### Scenario: Multi-line string
- **WHEN** the frontmatter contains `description: |` followed by the indented lines `line one` and `line two`
- **THEN** the value cell of `description` holds `line one` and `line two` on two lines

#### Scenario: Deeper nesting
- **WHEN** the frontmatter contains `authors:` with the list items `- name: A` and `- name: B`
- **THEN** the value cell of `authors` holds `[{name: A}, {name: B}]`

#### Scenario: Long value
- **WHEN** the frontmatter contains a `description` of 30 words and the output width is 40
- **THEN** the description wraps within the value column, `description` stays on one line and no line of the table is wider than 40 columns

#### Scenario: Plain mode
- **WHEN** a document with frontmatter is rendered to a pipe
- **THEN** the table contains no escape sequences

#### Scenario: Invalid YAML
- **WHEN** a document starts with `---`, `title: [unclosed`, `---` followed by `# Hello`
- **THEN** `title: [unclosed` is shown in a frame labelled `frontmatter - invalid YAML`, `# Hello` is rendered after it, and termd exits with status 0

#### Scenario: Empty frontmatter
- **WHEN** a document starts with `---`, `---` followed by `# Hello`
- **THEN** the output starts with `# Hello`

#### Scenario: Not a mapping
- **WHEN** a document starts with `---`, `Some text`, `---`
- **THEN** it is rendered as CommonMark: a horizontal line followed by the heading `## Some text`

#### Scenario: No closing line
- **WHEN** a document starts with `---`, `title: Doc` and has no other `---` or `...` line
- **THEN** it is rendered as CommonMark: a horizontal line followed by the paragraph `title: Doc`

#### Scenario: Byte order mark and CRLF
- **WHEN** a document starts with a UTF-8 byte order mark and uses CRLF line endings, and its first lines are `---`, `title: Doc`, `---`
- **THEN** the output starts with the frontmatter table
