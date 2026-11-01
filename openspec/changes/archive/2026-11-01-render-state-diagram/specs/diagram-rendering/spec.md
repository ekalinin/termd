# Spec Delta

## MODIFIED Requirements

### Requirement: Supported mermaid types
termd SHALL render mermaid `sequenceDiagram`, `flowchart`, `graph` and `erDiagram` blocks as text diagrams drawn with Unicode box-drawing characters, mermaid `stateDiagram` and `stateDiagram-v2` blocks as flowcharts, mermaid `mindmap` blocks as indented trees, and mermaid `gantt` blocks as bars on a time axis, in place of the code block.

#### Scenario: Flowchart
- **WHEN** a document contains a `mermaid` block starting with `flowchart LR`
- **THEN** the output shows the nodes as boxes connected by arrows, and the mermaid source is not shown

#### Scenario: Legacy graph keyword
- **WHEN** a `mermaid` block starts with `graph TD`
- **THEN** it is rendered as a flowchart

#### Scenario: State diagram
- **WHEN** a document contains a `mermaid` block starting with `stateDiagram-v2`, and another one starting with `stateDiagram` with the same lines
- **THEN** the output shows the states as boxes connected by arrows, the mermaid source is not shown, and both blocks are drawn the same

#### Scenario: Mindmap
- **WHEN** a document contains a `mermaid` block starting with `mindmap`
- **THEN** the output shows the nodes as an indented tree, and the mermaid source is not shown

#### Scenario: Gantt
- **WHEN** a document contains a `mermaid` block starting with `gantt`
- **THEN** the output shows the tasks as bars on a time axis, and the mermaid source is not shown

### Requirement: Diagram width
termd SHALL try to fit a flowchart, and every diagram of a state diagram, within the output width by compacting its spacing. If a left-to-right or right-to-left one still does not fit, termd SHALL lay it out top-to-bottom. A diagram that still does not fit, a state diagram with such a diagram, or a sequence or ER diagram wider than the output, SHALL be emitted in full as a wide block.

#### Scenario: Flowchart compacted to fit
- **WHEN** a flowchart is slightly wider than the output width with default spacing but fits with compact spacing
- **THEN** it is rendered with compact spacing within the output width

#### Scenario: Horizontal flowchart re-laid out
- **WHEN** a `flowchart LR` does not fit the output width even with compact spacing
- **THEN** it is rendered top-to-bottom

#### Scenario: Horizontal state diagram re-laid out
- **WHEN** a `stateDiagram-v2` with the line `direction LR` does not fit the output width even with compact spacing
- **THEN** it is rendered top-to-bottom

#### Scenario: Wide sequence diagram
- **WHEN** a sequence diagram with 8 participants is wider than the output width
- **THEN** it is emitted in full as a wide block and can be scrolled horizontally in the pager

### Requirement: Framed source fallback
termd SHALL show the source of a diagram block inside a frame, instead of rendering it, when a PlantUML block is not a sequence diagram, a mindmap or a state diagram (for example a class diagram), when the mermaid type is not supported (for example `classDiagram`, `gitGraph`), or when the diagram cannot be parsed or rendered. The frame SHALL carry a label naming the diagram language, the detected type when known, and the reason. A failure in one diagram SHALL NOT stop the rest of the document from rendering.

#### Scenario: Unsupported mermaid type
- **WHEN** a document contains a `mermaid` block starting with `classDiagram`
- **THEN** the block's source is shown in a frame labelled with `mermaid`, `classDiagram` and "not supported", and the following content is rendered normally

#### Scenario: PlantUML block
- **WHEN** a document contains a `plantuml` block with the lines `@startuml`, `class User` and `@enduml`
- **THEN** its source is shown in a frame labelled with `plantuml`, `class` and "not supported"

#### Scenario: Invalid mermaid
- **WHEN** a `mermaid` block starts with `sequenceDiagram` but contains a syntax error
- **THEN** its source is shown in a frame labelled with `mermaid`, `sequenceDiagram` and the parse failure, and termd exits with status 0

#### Scenario: Renderer failure
- **WHEN** rendering a supported diagram fails unexpectedly
- **THEN** the diagram is shown as framed source and the rest of the document is rendered

## ADDED Requirements

### Requirement: State diagram constructs
termd SHALL read the lines of a mermaid `stateDiagram` or `stateDiagram-v2` block after its first line as follows:

- A line that does not start with `state ` or `note ` and has `-->` before its first `:` is a transition: the state before the `-->`, the state after it, and the text after the `:` as its label. A state is `[*]` or an id, a run of characters without spaces and `:`; a `:::class` after an id is dropped.
- An id alone on a line, and `state id`, declare a state. `id : text` declares a state and adds the line `text` to its description. `state "text" as id` declares a state with the description `text`.
- `state id <<choice>>` and `state id [[choice]]` declare a choice, `<<fork>>` and `[[fork]]` a fork, and `<<join>>` and `[[join]]` a join.
- `state id {`, also `state "text" as id {`, declares a composite state whose lines follow up to the matching `}`. They are read as the lines at the top. A line `--` inside a composite state starts a new concurrent region of it.
- A state belongs to the composite state, or region, of the line where it first appears, and to the top of the diagram otherwise. `[*]` belongs to the composite state, or region, of its line: it is the start when it is the source of a transition and the end when it is the target.
- `direction TB`, `TD`, `BT`, `LR` or `RL` sets the direction of the diagrams of the composite state of its line, or of the top diagram. A composite state without one takes the direction of its parent, and the top diagram is `TB` by default.
- `note left of id : text` and `note right of id : text` are notes of the state `id`. A note without `: text` continues on the next lines up to `end note`, and its lines are joined with spaces.
- Empty lines, lines that start with `%%`, `classDef`, `class`, `style`, `accTitle:`, `accDescr:` and `accDescr {` ... `}` are ignored.

termd SHALL show the block as framed source labelled with `mermaid`, the type of its first line and the reason:

- `unsupported line N` for any other line, a `}` without a composite state and a `--` outside of one;
- `unclosed block on line N` for a composite state without its `}` and a note without its `end note`, N being the line that opens it;
- `no states` for a block without states.

N SHALL be the number of the line in the block, starting with 1.

#### Scenario: Transitions
- **WHEN** a state diagram has the lines `[*] --> Idle`, `Idle --> Loading : open` and `Loading --> [*]`
- **THEN** the output is the same as for the flowchart `flowchart TB` with the nodes `●`, `Idle`, `Loading` and `◉`, and the edges from `●` to `Idle`, from `Idle` to `Loading` labelled `open`, and from `Loading` to `◉`

#### Scenario: Descriptions
- **WHEN** a state diagram has the lines `state "Waiting" as W`, `W : press a key`, `L : Loading data` and `W --> L`
- **THEN** the output is the same as for the flowchart with the node `Waiting<br>press a key`, the node `Loading data` and an edge between them

#### Scenario: Choice, fork and join
- **WHEN** a state diagram has the lines `state c <<choice>>`, `state f [[fork]]`, `state j <<join>>`, `[*] --> c`, `c --> f : ok`, `f --> j` and `j --> [*]`
- **THEN** the output is the same as for the flowchart with the nodes `●`, `◇`, `━━━`, `━━━` and `◉` and the same edges

#### Scenario: Ignored lines
- **WHEN** a state diagram has the lines `%% a comment`, `classDef warn fill:#f00`, `class Idle warn`, `style Idle fill:#f00`, `accTitle: States` and `accDescr {`, `text`, `}` besides its transitions, and writes a state as `Idle:::warn`
- **THEN** the output is the same as for the diagram without these lines and with `Idle`

#### Scenario: Unsupported line
- **WHEN** a `mermaid` block has the lines `stateDiagram-v2`, `[*] --> A` and `A -> B`
- **THEN** its source is shown in a frame labelled with `mermaid`, `stateDiagram-v2` and `unsupported line 3`

#### Scenario: Unclosed composite state
- **WHEN** a `mermaid` block has the lines `stateDiagram-v2`, `state Active {` and `[*] --> Running`
- **THEN** its source is shown in a frame labelled with `mermaid`, `stateDiagram-v2` and `unclosed block on line 2`

#### Scenario: No states
- **WHEN** a `mermaid` block has only the line `stateDiagram`
- **THEN** its source is shown in a frame labelled with `mermaid`, `stateDiagram` and `no states`

### Requirement: State diagram layout
termd SHALL draw a state diagram as the following lines, without styles:

- The title from the frontmatter, followed by a blank line.
- The top diagram, with the states and the transitions of the top of the diagram, followed by its notes.
- A section for every composite state with states inside it: a blank line, the lines of its description joined with spaces, or its id, a blank line, and the diagram of every region of it with its notes. The diagrams of two regions SHALL be separated by a blank line, a line of `┈` as wide as the widest diagram of the composite state, and a blank line. The sections of the composite states of a diagram SHALL follow it in the order in which these states first appear, and the sections of the composite states inside a composite state SHALL come right after its section.

Every diagram SHALL be drawn as the flowchart of its direction with these nodes and edges:

- A node for every state of the diagram, in the order in which the states first appear, with the lines of the description of the state, or its id when it has none. A choice SHALL be a node `◇`, and a fork and a join a node `━━━`. The start of the diagram SHALL be one node `●` before the other nodes, and its end one node `◉` after them.
- An edge for every transition of which this is the innermost diagram that has both states, as nodes or inside the composite states of its nodes, in the order of the block. The edge SHALL go from the node that is or contains the source of the transition to the node that is or contains its target, with the label of the transition.

In the text of a node, every `-` after a `-` SHALL be shown as `‑` (U+2011) and every `>` after a `-` or a `=` as `›` (U+203A), so that the text is not read as a link of the flowchart. `<br>`, `<br/>` and `<br />` SHALL continue the text of a node on the next line, and SHALL be shown as a space in the label of an edge, the line of a section and a note.

A note SHALL be listed under the diagram that has its state as a node: a blank line before the first note of the diagram, then every note on its own line, in the order of the block, as the text of its state with its lines joined with spaces, `: ` and the text of the note. A note and the line of a section wider than the output SHALL wrap at whitespace, and their continuation lines SHALL start with 2 spaces.

#### Scenario: Composite state
- **WHEN** a state diagram has the lines `[*] --> Idle`, `Idle --> Active : start`, `state Active {`, `[*] --> Running`, `Running --> Paused : pause`, `Paused --> Running : resume`, `}` and `Active --> [*] : stop`
- **THEN** the output is the flowchart with the nodes `●`, `Idle`, `Active` and `◉` and the edges labelled `start` and `stop`, a blank line, the line `Active`, a blank line, and the flowchart with the nodes `●`, `Running` and `Paused` and the edges labelled `pause` and `resume`

#### Scenario: Transition into a composite state
- **WHEN** a state diagram has a composite state `Active` with the state `Running` inside it, and the line `Idle --> Running : resume` at the top
- **THEN** the top diagram has an edge from `Idle` to `Active` labelled `resume`, and the diagram of `Active` has no edge labelled `resume`

#### Scenario: Concurrent regions
- **WHEN** a state diagram has the lines `[*] --> Active`, `state Active {`, `[*] --> A`, `--`, `[*] --> B` and `}`
- **THEN** the section of `Active` has the diagram with the nodes `●` and `A`, a blank line, a line of `┈`, a blank line and the diagram with the nodes `●` and `B`

#### Scenario: Nested composite states
- **WHEN** a state diagram has the composite states `A` and `C` at the top, in this order, and the composite state `B` inside `A`
- **THEN** the sections follow the top diagram in the order `A`, `B` and `C`

#### Scenario: Direction
- **WHEN** a state diagram has the lines `direction LR`, `[*] --> A` and `A --> B`
- **THEN** the output is the same as for the flowchart `flowchart LR` with the nodes `●`, `A` and `B` and the same edges

#### Scenario: Notes
- **WHEN** a state diagram has the lines `[*] --> Idle`, `note right of Idle : waits for input`, `note left of Idle`, `first`, `second` and `end note`
- **THEN** the output is the top diagram, a blank line, the line `Idle: waits for input` and the line `Idle: first second`

#### Scenario: Links in the text of a state
- **WHEN** a state diagram has the lines `state "Save & exit" as S`, `state "a --> b" as T` and `S --> T : x --> y`
- **THEN** the nodes show `Save & exit` and `a -‑› b`, and the edge is labelled `x --> y`

#### Scenario: Title
- **WHEN** a state diagram has the frontmatter `title: Player`
- **THEN** the output starts with `Player`, a blank line and the top diagram

### Requirement: PlantUML state diagrams
termd SHALL draw a PlantUML state diagram, a block of the type `state`, with the layout of "State diagram layout", and SHALL NOT run PlantUML or any other external program. termd SHALL read the lines of the block as follows:

- `state id`, `state "text" as id` and `state id as "text"` declare a state, the last two with the description `text`, and `id : text` adds the line `text` to the description of the state `id`. A color such as `#pink`, `##[dashed]` or `#line:red` after a declaration is ignored.
- `state id {`, also with a description, a color or both, declares a composite state whose lines follow up to the matching `}`. A line of `--` or `||` inside it starts a new concurrent region.
- `state id <<choice>>`, `<<fork>>` and `<<join>>` are the same as in mermaid. `<<start>>` is drawn as `●` and `<<end>>` as `◉`. A state with any other stereotype is a plain state.
- `[*]` is the same as in mermaid. `[H]` and `[H*]` are the history of the composite state of their line, and `id[H]` and `id[H*]` the history of the composite state `id`; a history is a state inside its composite state drawn as `H` or `H*`.
- A line that does not start with `state ` or `note ` and has an arrow before its first `:` is a transition: a source, the arrow, a target and the text after the `:` as its label. An arrow is one or more `-` with an optional direction (`up`, `down`, `left`, `right`, `u`, `d`, `l`, `r`, `do`, `le`, `ri`) and an optional style in `[` and `]` among them, ending with `>`. An arrow that starts with `<` and has no `>` swaps the source and the target.
- `left to right direction` makes every diagram `LR`, and `top to bottom direction` makes every diagram `TB`.
- `note left of id`, `note right of id`, `note top of id` and `note bottom of id` are notes of the state `id`, and `note on link` is a note of the transition before it, listed with the texts of the source and the target of the transition joined by ` --> `. A note without `: text` continues on the next lines up to `end note`, and its lines are joined with spaces.
- `title text` is printed above the diagrams. `\n` in a description continues it on the next line, and in a label or a note it is shown as a space.
- `@startuml`, `@enduml`, empty lines, `'` comments, `/'` ... `'/` comments, `skinparam` (also with a `{` ... `}` block), `hide`, `show`, `scale`, `header`, `footer`, `legend` ... `endlegend`, `caption` and lines that start with `!` are ignored.

termd SHALL show the block as framed source labelled with `plantuml`, `state` and the reason: `unsupported line N` for any other line, a `}` without a composite state and a region separator outside of one; `unclosed block on line N` for a composite state without its `}` and a note without its `end note`; `no states` for a block without states. N SHALL be the number of the line in the block, starting with 1.

#### Scenario: Transitions and arrows
- **WHEN** a PlantUML state diagram has the lines `[*] --> Idle`, `Idle -> Loading : open`, `Loading -down-> Ready`, `Ready -[#red]-> Done`, `Done <- Loading` and `Done --> [*]`
- **THEN** the output is the same as for the mermaid state diagram `[*] --> Idle`, `Idle --> Loading : open`, `Loading --> Ready`, `Ready --> Done`, `Loading --> Done` and `Done --> [*]`

#### Scenario: Declarations
- **WHEN** a PlantUML state diagram has the lines `state "Waiting" as W`, `state L as "Loading data"`, `state X #pink`, `W : press a key`, `W --> L` and `L --> X`
- **THEN** the output is the same as for the mermaid state diagram `state "Waiting" as W`, `state "Loading data" as L`, `state X`, `W : press a key`, `W --> L` and `L --> X`

#### Scenario: Composite state with regions
- **WHEN** a PlantUML state diagram has the lines `[*] --> Active`, `state Active {`, `[*] -> A`, `||`, `[*] -> B` and `}`
- **THEN** the output is the same as for the mermaid state diagram with the same lines, `--` in place of `||`

#### Scenario: Special states
- **WHEN** a PlantUML state diagram has the lines `state s <<start>>`, `state c <<choice>>`, `state p <<entryPoint>>`, `state e <<end>>`, `s --> c`, `c --> p` and `p --> e`
- **THEN** the output is the same as for the mermaid state diagram `state "●" as s`, `state c <<choice>>`, `state p`, `state "◉" as e`, `s --> c`, `c --> p` and `p --> e`

#### Scenario: History
- **WHEN** a PlantUML state diagram has the lines `state Active {`, `[*] --> Running`, `Running --> [H]`, `}` and `Idle --> Active[H]`
- **THEN** the diagram of `Active` has a node `H` with an edge from `Running`, and the top diagram has an edge from `Idle` to `Active`

#### Scenario: Notes
- **WHEN** a PlantUML state diagram has the lines `[*] --> A`, `A --> B`, `note on link : retry`, `note top of A : first`, `note bottom of B`, `waits` and `end note`
- **THEN** the output is the diagram, a blank line, and the lines `A --> B: retry`, `A: first` and `B: waits`

#### Scenario: Direction, title and ignored lines
- **WHEN** a PlantUML state diagram has the lines `title Player`, `left to right direction`, `skinparam monochrome true`, `hide empty description`, `' a comment` and `!theme plain` besides the transitions `[*] --> A` and `A --> B`
- **THEN** the output is `Player`, a blank line and the output of the mermaid state diagram `direction LR`, `[*] --> A` and `A --> B`

#### Scenario: Unsupported line
- **WHEN** a `plantuml` block has the lines `@startuml`, `[*] --> A`, `note "free" as N1` and `@enduml`
- **THEN** its source is shown in a frame labelled with `plantuml`, `state` and `unsupported line 3`
