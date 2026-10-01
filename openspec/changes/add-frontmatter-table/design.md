# Design

## Context

`render.Render` parses the whole source with goldmark (CommonMark + GFM) and lays out the blocks. Both `cmd/termd` and the site generator (`internal/site`) go through it. goldmark has no frontmatter support, so a leading `---` block becomes a thematic break followed by a setext heading or a paragraph.

`internal/table` lays out GFM tables; the header row defines the number of columns, and the header is always followed by a rule. `internal/diagram` already has a framed source fallback (`diagram.Frame`) and guards the diagram library with `recover`.

The requirements are in `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- The frontmatter is cut off before goldmark sees the source, so the markdown parsing of the body does not change.
- The frontmatter table reuses the width allocation of `internal/table`.
- The documents without frontmatter render byte for byte as before; none of the current fixtures starts with `---`.

**Non-Goals:**
- A frontmatter example on the landing page. The site gets the feature through `render.Render`, but its examples stay as they are.
- Using frontmatter fields elsewhere, for example `title` as a heading.

## Decisions

### Split the frontmatter off in `render.Render`

`Render` calls a `splitFrontmatter(src)` step before `Parse`. It returns the frontmatter block (if any) and the body, and the renderer works on the body as its `src`, so goldmark segments keep pointing into the right slice. The frontmatter block, if shown, is prepended to the list of document blocks and joined with a blank line like any other block.

The code lives in a new file `internal/render/frontmatter.go`: it needs the table, the frame and the text spans, which `render` already imports. A separate package would only move these imports.

Alternative: a goldmark extension that parses frontmatter as an AST node. Rejected: it couples the detection to the parser, and the rule "valid YAML but not a mapping renders as CommonMark" would need the extension to give the lines back to the parser.

### Detection

1. A UTF-8 BOM is skipped for the detection only. When there is no frontmatter, the original `src` is rendered unchanged.
2. The first line, without a trailing `\r` and trailing spaces, is `---`.
3. The closing line is the first later line that is `---` or `...` after the same trimming. Without it there is no frontmatter.
4. The lines between are parsed into a `yaml.Node`:
   - parse error: the block is shown in a frame;
   - no document (empty block or only comments): the block is omitted;
   - the root is not a mapping: there is no frontmatter, the original `src` is rendered;
   - a mapping: the block is shown as a table.

`diagram.headerLine` also skips a leading `---` block, inside mermaid sources, but it is lenient (an unclosed block hides the rest of the source). It is not reused, since the document rule must fall back to CommonMark.

### YAML library: `go.yaml.in/yaml/v3`

A spike with v3.0.5 confirmed what the design needs: `yaml.Node` keeps the key order; an empty or comment-only block gives a node without a document; a scalar block gives a scalar root; `title: [unclosed` returns an error; a sequence or mapping node marshalled with `Style = yaml.FlowStyle` gives `[{name: A}, {name: B}]`.

Alternatives:
- `gopkg.in/yaml.v3`: the same API, but the repository is archived since April 2025. `go.yaml.in/yaml/v3` is its maintained continuation.
- `github.com/goccy/go-yaml`: more features than needed (its own AST, path queries, validation).
- A hand-written parser of `key: value` lines: rejected in the discussion. It cannot tell invalid YAML apart, and quoting, block scalars and anchors would be heuristics.

The parse runs under `recover`, like `diagram.safeRender`; a panic is treated as invalid YAML.

### Value formatting

For each key-value pair of the root mapping, the key cell is the key's scalar text, bold. The value cell is built from the value node:

| Value node | Cell |
|---|---|
| scalar | `Value` without the trailing newline; inner newlines become `text.Break` spans |
| sequence of scalars | the values joined with `, ` |
| mapping with scalar values | one `key: value` line per entry, separated by `text.Break` |
| anything else (nested collections, aliases) | a copy of the node with `Style = yaml.FlowStyle`, marshalled and trimmed |

`text.Wrap` already turns `text.Break` into a line break inside a table cell. Runs of spaces inside a multi-line string collapse to one space, as everywhere in wrapped text.

### Headerless table in `internal/table`

`Table` with a nil `Header` is a table without a header: the number of columns is the widest row, `Measure` and `Widths` work over the rows only, and `Render` emits neither a header line nor the rule. GFM tables always have a header, so their output does not change. The ragged rows rule for GFM tables stays as it is.

Alternative: a separate two-column layout for the frontmatter. Rejected: the spec requires the same layout rules as tables, which would be duplicated.

### Invalid YAML frame

The block's source (the lines between the delimiters) goes through `diagram.Frame("frontmatter - invalid YAML", src)`. The block is wide when the frame is wider than the output, as for diagram fallbacks. The label does not include the YAML error, so the label stays the one given in the spec.

## Risks / Trade-offs

- [A document that starts with a thematic break and continues with text that is a valid YAML mapping, for example `---`, `Note: read this first`, `---`, becomes a table] → Accepted: GitHub does the same, and such a document is rare.
- [A document that starts with a thematic break and continues with text that is invalid YAML ends up in the invalid YAML frame instead of rendering as markdown] → Accepted for the same reason; valid non-mapping YAML, the more likely case for plain text, still renders as CommonMark.
- [Control characters in frontmatter values reach the terminal] → The same as for the rest of the document today; it belongs to the separate fix for escape sequences in document text.
- [The new dependency grows the binary] → The size before and after is measured in the tasks and noted in the PR.
- [A panic in the YAML library on hostile input] → `recover` turns it into the invalid YAML frame.

## Migration Plan

None. Documents without frontmatter render as before, so the existing golden files do not change. Rollback is a revert of the change.
