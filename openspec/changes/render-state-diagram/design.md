# Design

## Context

`diagram.Render` frames `stateDiagram` and `stateDiagram-v2` as "not supported", and `plantumlType` already gives a PlantUML block with a `state ` or `[*]` line the type `state`, which `renderPlantUML` frames the same way. Flowcharts go through `renderFlowchart`, which calls mermaid-ascii with `MaxWidth`, and rewrites `LR` and `RL` to `TD` when the result is still too wide. Mindmaps and gantt charts take their title with `frontmatterTitle`, and `text.Wrap` wraps text at whitespace.

mermaid-ascii has no state diagrams (issue #61), and its graph package only exports `Parse(text)` and `Draw`, so its input is flowchart text. Spikes with the termd binary and with mermaid-ascii at its latest commit showed:

- A state diagram translated by hand into a flowchart draws well: `[*]` as a node `●` or `◉`, transitions as labelled edges, `<br>` in a node label as two lines, self-transitions as loops. `((●))` is not a shape of the library and is shown as written, and every node is a box of 5 lines.
- A composite state as a `subgraph` that is also the end of an edge breaks the layout: the library treats the id as a node inside the frame, the inner `●` was drawn outside the frame, and edge labels overlapped. The same states drawn as a top diagram and a separate diagram of the composite state are clean.
- A note as a node linked with `-.-` is not connected, and with `-.->` it takes the place of the next state and widens its column.
- The pinned version splits `Tom & Jerry` into two nodes even in a quoted label, and `x & y` in an edge label. The commit of 2026-10-19 (`5a9df15`) keeps both. In both versions `-->` inside a quoted node label splits the line, a literal `\n` in an edge label starts a new line, and `|` and `-->` in an edge label work because the target is a bare id.
- `‑` (U+2011), `›` (U+203A), `●`, `◉`, `◇` and `━` are one column wide for `uniseg` and keep the boxes aligned.

The requirements are in `specs/diagram-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- One state model for mermaid and PlantUML, drawn by one function, so that composite states, regions, `[*]`, notes and width fitting behave the same for both languages.
- The diagrams of a state diagram drawn by the flowchart renderer, with its width fitting.
- Every other diagram renders byte for byte as before; the only changed golden files are those whose fixtures switch from `stateDiagram-v2` to `classDiagram` or become drawn state diagrams.

**Non-Goals:**
- A layout engine for state diagrams, or changes to mermaid-ascii.
- Parity with the grammars of Mermaid and PlantUML: lines outside the spec are unsupported lines.

## Decisions

### Both parsers build one model; the model is drawn through flowchart text

`internal/diagram/state.go` holds the model, the mermaid parser and the drawer; the PlantUML parser goes into `plantuml.go` next to the parsers of sequence diagrams and mindmaps. The model:

- a diagram with a title and a top region;
- a region with its states in the order of their first appearance, flags for a start and an end, its direction and its notes;
- a state with an id, the lines of its description, a kind (plain, choice, fork, join) and, for a composite state, its regions;
- the transitions in the order of the block, each with its two ends (a state, or the start or end of a region) and its label.

This follows the mindmaps, whose two parsers build one `mindNode` tree.

Alternatives:
- Translating PlantUML into mermaid state text, as PlantUML sequence diagrams are translated into mermaid sequence text. Rejected after a look at the syntax: `[H]`, `<<start>>`, `<<end>>` and `note on link` have no mermaid counterpart, so the translation would need made-up mermaid lines, and the text would be parsed twice.
- Building the AST of mermaid-ascii. Not possible: the graph package exports no constructor besides `Parse`; and the `render-plantuml` design already chose text over the AST of the library.

### A diagram per region, drawn by `renderFlowchart`

The drawer walks the regions in the order of the spec: the top region, then, for every composite state with states inside it, in the order of first appearance and depth first, its regions. Each region becomes the text of one flowchart, which goes to `renderFlowchart(src, width)`, so it is compacted and re-laid out top to bottom on its own. The pieces are joined with the blank lines, the line of the composite state and the `┈` divider of the spec, and `result(out, width)` marks the block wide when any line is wider than the output.

A transition is drawn in the innermost region that has both ends: the drawer takes the chain of regions of each end, finds the last region they share, and replaces each end by its ancestor in that region. `[*]` belongs to the region of its line, so `[*] --> Running` inside `Active` stays in the diagram of `Active`, and `Idle --> Running` at the top becomes `Idle --> Active`.

The direction of a region comes from its own `direction` line, then from the composite state of its parent region, then `TB`. PlantUML `left to right direction` sets the direction of the top region, which the others inherit.

Alternative: composite states as subgraphs in one flowchart. Rejected with the user, after the spike above.

### Flowchart text that the library cannot misread

For each region the drawer writes `flowchart <dir>`, then one line per node and one line per edge:

- Nodes get generated ids `n1`, `n2`, ... in the order of the spec (start, states, end), so state ids such as `end`, `graph`, `subgraph` or ids with brackets cannot clash with flowchart syntax. A node line is `nK["text"]`, the lines of a description joined with `<br>`.
- In the text of a node, a `-` after a `-` becomes `‑` and a `>` after a `-` or `=` becomes `›`. This breaks every link operator of the library (`-->`, `---`, `--o`, `--x`, `-.->`, `==>`, `<-->`) with characters of the same width that look alike.
- An edge line is `nA --> nB`, or `nA -->|label| nB` when the transition has a label. `<br>` and PlantUML `\n` in a label become spaces before the line is written, since a literal `\n` splits the line.
- The title is not passed as frontmatter. The drawer prints it, followed by a blank line, as mindmaps and gantt charts do, which keeps YAML quoting out of the generated text.

Alternatives:
- Escaping characters in the labels. Not possible: the library decodes no escapes, `#quot;` is shown as written.
- Placeholders of the same width replaced in the output. Rejected: the library wraps labels and replaces spaces in edge labels by `─`, so the placeholders would not map back reliably.

### mermaid-ascii at `5a9df15`

The bump fixes `&` in quoted node labels and in edge labels, which is common in state descriptions (`Save & exit`). It is the only upstream commit after the pinned one, and it touches only the handling of `&`. The bump is the first task, and its check is that `make golden` changes no file.

### Parsing

Both parsers read the block line by line with a stack of open composite states and regions, like the stack of the gantt and the mindmap parsers. The first appearance of an id creates the state in the current region; later lines find it through a map from id to state. A multi-line note collects its lines up to `end note`. An unknown line returns the error `unsupported line N`; a stack that is not empty at the end returns `unclosed block on line N` with the line that opened the innermost open block; a diagram without states returns `no states`. `Render` frames the source with the error as the reason.

The mermaid parser skips the frontmatter with `headerLine` and takes the title with `frontmatterTitle`. A transition is split at the first `:` and then at `-->`, after lines that start with `state ` or `note ` have been handled, so `state "a --> b" as T` and `S : a --> b` stay a declaration and a description.

The PlantUML parser recognises an arrow with one regular expression: `<?` then `-+`, an optional direction and an optional `[...]` in any order between dashes, and `>?`, with at least one of `<` and `>`. Ids, `[*]`, `[H]`, `[H*]` and `id[H]` are the ends. `<<start>>` and `<<end>>` states are plain states with the description `●` and `◉`, and a history is a plain state `H` or `H*` in the first region of its composite state, which it creates when the composite state has no region yet.

### Fixtures

`internal/diagram/testdata/state.mmd` becomes a drawn diagram and its golden file changes. A new `class.mmd` with a `classDiagram` takes its place in `TestFallbackLabels`. New fixtures cover a mermaid diagram with every construct of the spec and a PlantUML one; `TestSupportedTypesAreRendered` adds them. The scenarios that compare a state diagram with a flowchart, or a PlantUML diagram with a mermaid one, are tests that compare two `Render` calls, as `TestPlantumlSequence` does.

## Risks / Trade-offs

- [Every `●`, `◉` and `━━━` is a box of 5 lines, so state diagrams are taller than in Mermaid] → Accepted with the user as the cost of reusing the flowchart renderer.
- [A transition into an inner state from outside loses its exact target on the outer diagram] → Accepted with the user; the inner diagram shows the state by name.
- [Notes are far from their states] → Each note line starts with the text of its state.
- [Look-alike characters in a node text differ from the source when copied] → They appear only in texts with `--` or `->`, `=>`, which are rare in state names.
- [A new mermaid-ascii commit could change the layout of flowcharts] → The bump is a separate task whose check is that no golden file changes.
- [A PlantUML state diagram without `state ` and `[*]` lines is detected as a sequence diagram] → This is how PlantUML picks the type too, and detection does not change.

## Migration Plan

The golden files of `state.mmd`, `testdata/diagrams.md` and `testdata/control.md` change: the first is now drawn, the other two frame a `classDiagram` instead. All other output stays the same. Rollback is a revert of the change, including the bump of mermaid-ascii.
