# Proposal

## Why

A YAML frontmatter block at the start of a document is rendered as markdown today: the opening `---` becomes a horizontal line and the keys turn into a heading, for example `## title: Doc tags: [a, b]`. glow strips the block and loses the metadata; GitHub shows it as a table. termd should show it as a table too, because tables are what termd lays out well.

## What Changes

- A document whose first line is `---`, which has a closing `---` or `...` line, and whose block between them is a YAML mapping, starts with a two-column table: the key and its value, one row per key, in the order of the source. The table has no header row; keys are bold when styling is enabled.
- Values are shown as text: a scalar as its value, a list as its items separated by commas, a mapping as one `key: value` line per entry, a multi-line string with its line breaks. Deeper nesting is shown as compact YAML.
- The table follows the table layout rules: words are never split, the value column wraps, and a table that does not fit at its minimum widths is emitted as a wide block.
- A block that starts like frontmatter but is not valid YAML is shown as its source in a frame labelled `frontmatter - invalid YAML`, the same way an unsupported diagram is shown. The rest of the document renders normally and termd exits with status 0.
- A block between `---` lines that is valid YAML but not a mapping, or a document without a closing line, is rendered as markdown, as today.
- YAML is parsed with `go.yaml.in/yaml/v3`, the maintained fork of `gopkg.in/yaml.v3`.

Out of scope:

- TOML (`+++`) and other frontmatter formats.
- Stripping the frontmatter instead of showing it, as glow does.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `markdown-rendering`: a new requirement for YAML frontmatter shown as a key-value table at the start of the document, with a framed source fallback for invalid YAML.

## Impact

- `internal/render`: frontmatter is detected and removed from the source before parsing, and its table or frame becomes the first block of the document.
- `internal/table`: a table without a header row, for the frontmatter table. GFM tables do not change.
- `internal/diagram.Frame` is reused for the invalid YAML frame.
- `go.mod`: new direct dependency `go.yaml.in/yaml/v3`.
- `testdata`: new golden fixtures for frontmatter.
- `README.md`: frontmatter in the features and in "How it works".
