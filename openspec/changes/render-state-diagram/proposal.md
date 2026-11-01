# Proposal

## Why

A mermaid `stateDiagram` or `stateDiagram-v2` block and a PlantUML state diagram are shown as their source in a frame labelled "not supported". The roadmap of README lists the rendering of `stateDiagram`. mermaid-ascii, which draws the flowcharts of termd, has no state diagrams: its issue #61 "Support stateDiagram-v2" has been open since 2026-04-26 and proposes to draw them as `graph TD`, with `[*]` as synthetic start and end nodes. A state diagram is a graph of states and labelled transitions, which the flowchart renderer of termd already draws and fits into the output width.

The markdown files of the Go module cache have no state diagram besides the fixtures of termd, so the syntax follows the Mermaid and PlantUML documentation.

## What Changes

- termd draws mermaid `stateDiagram` and `stateDiagram-v2` blocks and PlantUML state diagrams as flowcharts, with the renderer and the width fitting of mermaid flowcharts. No external program is run.
- A state is a box with its name or its description, and a transition is an arrow with its label. Special states are boxes with a symbol:

  | State | termd |
  |---|---|
  | `[*]` as the source of a transition, PlantUML `<<start>>` | `●` |
  | `[*]` as the target of a transition, PlantUML `<<end>>` | `◉` |
  | `<<choice>>` | `◇` |
  | `<<fork>>`, `<<join>>` | `━━━` |
  | PlantUML `[H]`, `[H*]` | `H`, `H*` |

- A composite state (`state X { ... }`) is a box in the diagram of its parent, and its inside is drawn as a separate diagram below, under a line with its name. Every concurrent region of a composite state (`--`, PlantUML `||`) is a separate diagram. A transition between states of different composite states is drawn between their ancestors in the innermost diagram that has both.
- Notes are listed as text under the diagram of their state, because a flowchart has no notes.
- `direction` and PlantUML `left to right direction` set the direction of the diagrams. Styles, classes, colors, arrow directions such as `-down->` and the layout lines of PlantUML are ignored. A line that termd cannot read shows the block as framed source labelled, for example, `mermaid - stateDiagram-v2 - unsupported line 4` or `plantuml - state - unsupported line 4`.
- `classDiagram` replaces `stateDiagram-v2` as the example of an unsupported mermaid type in the specs, the tests, the fixtures, the landing page and README.
- mermaid-ascii is updated to its commit of 2026-10-19, which keeps `&` inside quoted node labels and edge labels. With the pinned version, a state described as `Save & exit` would be split into two nodes.

Out of scope:

- A layout of its own for state diagrams: rounded boxes, small start and end markers, diamonds and bars.
- Composite states drawn as frames around their states inside one diagram.
- PlantUML floating notes (`note "text" as N`) and their links, which are unsupported lines.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `diagram-rendering`: "Supported mermaid types" adds state diagrams; "Diagram width" fits every diagram of a state diagram like a flowchart; "Framed source fallback" no longer frames state diagrams and takes `classDiagram` as its example; new requirements "State diagram constructs", "State diagram layout" and "PlantUML state diagrams".
- `cli`: the scenario of "Diagram fallbacks do not fail the run" uses a `classDiagram` block.

## Impact

- `internal/diagram`: a state model with a mermaid parser and a PlantUML parser, and a drawer that turns every diagram of the model into flowchart text for `renderFlowchart`; `Render` handles `stateDiagram`, `stateDiagram-v2` and the PlantUML type `state`.
- `go.mod`, `go.sum`: mermaid-ascii `v0.0.0-20261019220251-5a9df1560871`.
- `internal/diagram/testdata/state.mmd` is now drawn, so its golden file changes; `TestFallbackLabels` takes a new `class.mmd` for the unsupported mermaid type. New state fixtures in `internal/diagram/testdata` and a new `testdata/state.md` with its golden files.
- `testdata/diagrams.md` and `testdata/control.md` frame a `classDiagram` instead, so their golden files change. `internal/render/render_test.go`, `cmd/termd/main_test.go` and `internal/site/examples/4-unsupported.md` use a `classDiagram`.
- `README.md`: "Features", "Examples", the "Diagrams" section, "Limitations" and "Roadmap".
