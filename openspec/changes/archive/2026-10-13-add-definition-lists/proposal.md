# Proposal

## Why

termd does not parse definition lists. A term followed by a line `: definition` is rendered as one paragraph, `Term : Definition of the term`. glamour parses them with the definition list extension of goldmark and shows the term and its definition on separate lines.

## What Changes

- termd parses definition lists (the PHP Markdown Extra syntax supported by goldmark): one or more term lines followed by one or more lines that start with `: `. Every line of the paragraph before the first definition is a term of its own.
- A term is shown on its own line, in bold when styling is enabled. Each definition is shown below its term, indented by 4 columns, and wrapped like the text of a list item, with wrapped lines indented to the start of the definition.
- A term can have several definitions, and a definition can contain several blocks.
- In plain mode the output has the same layout without styles.
- GitHub does not support definition lists: on GitHub, `Term` followed by `: Definition` is rendered as one paragraph. For such documents termd now differs from GitHub, which it has followed so far (GFM parsing, alerts, footnotes). No fixture, README or site example contains a line that starts with `: `.

Out of scope:

- A theme color for terms.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `markdown-rendering`: the "GFM parsing" requirement adds the definition list extension, and a new requirement describes how definition lists are rendered.

## Impact

- `internal/render`: the parser gets `extension.DefinitionList`; blocks for `DefinitionList`, `DefinitionTerm` and `DefinitionDescription`.
- No new dependency: the extension is part of goldmark.
- Tests and a golden fixture with definition lists; documents without definition lists render as before.
- `README.md`: a "Definition lists" subsection of "How it works", as for footnotes and alerts.
