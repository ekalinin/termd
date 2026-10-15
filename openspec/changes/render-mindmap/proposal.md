# Proposal

## Why

A mermaid `mindmap` block is shown as its source in a frame labelled "not supported". mermaid-ascii, which draws the other mermaid diagrams, has no mindmap support, and neither does its latest commit, which termd already uses. Drawing a mindmap as a flowchart through mermaid-ascii does not work either: the 10 nodes of the example from the Mermaid documentation take 110 columns at an output width of 80, every node is a box of 5 lines, and an edge crosses half of the diagram.

## What Changes

- termd draws a mermaid `mindmap` as an indented tree, as the `tree` command does: the root on the first line, every other node on its own line after `├── ` or `└── `, with `│` guides that connect the children of a node.
- The text of a node is shown without its shape: `root((mindmap))`, `id[Square]`, `id(Rounded)`, `id))Bang((`, `id)Cloud(` and `id{{Hexagon}}` show `mindmap`, `Square`, `Rounded`, `Bang`, `Cloud` and `Hexagon`. `::icon(...)` and `:::class` lines are ignored. The text of a markdown string, such as ``id["`**Root** text`"]``, is shown without the backticks and the `**` and `*` markers. The output has no styles, like the other diagrams.
- The hierarchy follows the indentation of the source, as in Mermaid: a node is a child of the nearest node before it with less indentation. A tab counts as 4 spaces.
- `<br>`, `<br/>` and a line break inside a markdown string continue the text of the node on the next line, under the start of its text.
- A text longer than the output width wraps at whitespace, under the start of its text and with the `│` guides kept. The text is never narrower than 10 columns; a mindmap that is still wider than the output is emitted in full as a wide block.
- A `title` from the frontmatter of the block is printed above the tree, as for the other diagrams.
- A mindmap with more than one root, with a shape or a markdown string that is not closed, or without nodes is shown as its source in a frame labelled with the reason, as other diagrams that cannot be parsed.

Out of scope:

- A radial or horizontal layout of the branches, as Mermaid draws it.
- Shapes, icons, classes, and bold or italic text of markdown strings.
- An ASCII form of the tree characters; `add-ascii-mode` will need one for `├──`, `└──` and `│`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `diagram-rendering`: "Supported mermaid types" adds `mindmap`; "Framed source fallback" no longer names `mindmap` as an unsupported type; a new requirement "Mindmap layout" describes the tree.

## Impact

- `internal/diagram`: a parser and a layout for mindmaps, and a `mindmap` case in `Render`; mermaid-ascii does not change.
- Tests and a golden file `internal/diagram/testdata/mindmap.mmd`, and a new fixture `testdata/mindmap.md` with its golden files; the existing golden files do not change.
- `README.md`: "Features", the "Diagrams" section, "Limitations" and "Roadmap", which list `mindmap` as shown as source or as planned.
