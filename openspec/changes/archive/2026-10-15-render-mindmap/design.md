# Design

## Context

`diagram.Render` detects the mermaid type with `MermaidType` and passes `sequenceDiagram`, `erDiagram`, `flowchart` and `graph` to mermaid-ascii; any other type is shown by `fallback` as framed source labelled `mermaid - <type> - not supported`. A library error goes through `reason` into the label of the frame. The output of a supported diagram becomes plain `text.Line`s, and the block is wide when its widest line is wider than the output. The frontmatter `title` of the supported types is printed by mermaid-ascii itself. `internal/diagram` already imports `internal/text`; `go.yaml.in/yaml/v3` is a dependency of the module, used by `internal/render` for document frontmatter.

mermaid-ascii has no mindmap support, also at its latest commit, which termd uses. A spike rendered the example of the Mermaid documentation as a `graph LR` through mermaid-ascii: 110 columns at `MaxWidth` 80, boxes of 5 lines, and an edge from `Popularisation` to its child that runs across the diagram. Mermaid builds the tree from indentation: a node is a child of the nearest node before it with less indentation, a node without such a node next to the root is the error "There can be only one root", and a node line is a text or an id with one of six shapes.

The requirements are in `specs/diagram-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- A mindmap of any size fits the output width, except a tree so deep that its texts would be narrower than 10 columns.
- The other diagram types render byte for byte as before; mermaid-ascii does not change.

**Non-Goals:**
- Full parity with the mindmap grammar of Mermaid: termd reads the constructs of the spec and shows framed source for the errors it names; any other line is the text of a node.
- Styles in diagrams.

## Decisions

### A parser and a layout of termd in `internal/diagram/mindmap.go`

`renderMindmap(src string, width int) (string, error)` parses the source into a tree and lays it out; `Render` gets a `case "mindmap"` that calls it, so the rest of `Render` (plain lines, `Wide`, `fallback` with `reason` on an error) is shared with the other types. The errors are plain `errors.New` values with the reasons of the spec, such as `more than one root`; `reason` keeps them as they are.

Alternatives:
- A mindmap converted to a `graph LR` for mermaid-ascii. Rejected by the spike: it does not fit the width and its edges cross.
- A horizontal tree, the root on the left and the branches to the right. Rejected with the user: the documentation example is already 83 columns wide, texts cannot wrap, and a fallback to the indented tree would be needed anyway.

### Parsing

The parser walks the lines after the header line found by `headerLine`:

- Empty lines and lines that start with `%%` are skipped; lines that start with `::icon(` or `:::` are skipped as decorations of the node before them.
- The indentation is the number of leading spaces, a tab counting as 4. A stack of open nodes with their indentation gives the parent: nodes are popped while their indentation is not less than the new one. The first node is the root; a later node with an empty stack is the error `more than one root`.
- The node text: the parser looks for the first shape opener, trying `((`, `))` and `{{` before `(`, `)` and `[`. Without one, the trimmed line is the text. With one, the line, trimmed, has to end with the matching closer (`))`, `((`, `}}`, `)`, `(`, `]`), and the text is between them; otherwise it is the error `unclosed shape`.
- A markdown string starts with `"` and a backtick right after the opener. When the line does not hold the closing backtick and `"`, the next source lines are added to the text with line breaks until they do; the end of the source is the error `unclosed markdown string`. The backticks and `"` are removed, and so are the `**` and `*` markers.
- `<br>`, `<br/>` and `<br />` become line breaks of the text.
- A source without nodes is the error `no nodes`.

Like Mermaid, `Tools (optional)` is the id `Tools ` with the rounded text `optional`, and `f(x) = y` is an unclosed shape.

### Frontmatter title

The parser reads the leading `---` block with `yaml.Unmarshal` into a struct with a `Title` field. A title is printed, followed by a blank line, above the tree, as mermaid-ascii does for the other types. Invalid YAML in the frontmatter means no title, not an error.

Alternative: matching `title:` by hand. Rejected: quoted titles and comments would need YAML rules again.

### Layout

The layout walks the tree depth first and builds every line from:

- the guides of the ancestors below the root: `│   ` for an ancestor with a later sibling, 4 spaces otherwise;
- the connector of the node: `├── `, or `└── ` for the last child; none for the root.

The text of a node at depth `d` is split at its line breaks, and every part is wrapped with `text.Wrap` at `max(width - 4*d, 10)` columns without splitting words. The first line gets the connector; the other lines get the guides of the ancestors and the guide of the node itself, so they start at the column of the text and continue the `│` of the siblings of the node. A word longer than the width keeps the line long, and `Render` marks the block wide.

Alternative: splitting long words, as paragraphs do. Rejected: a split word in a tree reads as two nodes less clearly than a long line in the pager.

## Risks / Trade-offs

- [Text with parentheses in the middle, such as `f(x) = y`, makes the whole mindmap framed source] → Mermaid fails on it too; the frame shows the source and the reason.
- [termd accepts some lines that Mermaid rejects, for example a closer inside the text] → The tree still shows the text; a mindmap that Mermaid renders is drawn as the spec says.
- [The tree characters are not ASCII] → `add-ascii-mode` will need ASCII forms for `├──`, `└──` and `│`, as for the other diagrams.
- [No README or document of the module cache has a mindmap, so there is no real-world sample] → The tests use the examples of the Mermaid documentation and the constructs of its grammar.

## Migration Plan

None. Mindmaps were framed source; the other diagrams and the existing golden files do not change. Rollback is a revert of the change.
