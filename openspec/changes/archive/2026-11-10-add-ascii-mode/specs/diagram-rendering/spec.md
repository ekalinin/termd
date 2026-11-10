## MODIFIED Requirements

### Requirement: Supported mermaid types
termd SHALL render mermaid `sequenceDiagram`, `flowchart`, `graph` and `erDiagram` blocks as text diagrams drawn with Unicode box-drawing characters, or with ASCII characters in ASCII mode, mermaid `stateDiagram` and `stateDiagram-v2` blocks as flowcharts, mermaid `mindmap` blocks as indented trees, and mermaid `gantt` blocks as bars on a time axis, in place of the code block.

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

## ADDED Requirements

### Requirement: ASCII mode
In ASCII mode, termd SHALL draw diagrams and framed source with ASCII characters only, apart from the text that comes from the diagram source, which SHALL NOT change:

- Sequence diagrams, flowcharts, ER diagrams and the flowcharts of state diagrams, mermaid and PlantUML, SHALL be drawn by the diagram library with its ASCII characters instead of box-drawing characters and arrows.
- In state diagrams, mermaid and PlantUML, the start state SHALL be `*` instead of `●`, the end state `(*)` instead of `◉`, a choice `<>` instead of `◇`, and a fork or a join `===` instead of `━━━`. The line between concurrent regions SHALL be made of `.` instead of `┈`. In the text of a node, every `-` after a `-` and every `>` after a `-` or a `=` SHALL be kept and preceded by a space, instead of being shown as `‑` and `›`.
- In mindmaps, mermaid and PlantUML, a node SHALL follow `|-- ` instead of `├── `, or a backtick and `-- ` instead of `└── `, and an ancestor with a later sibling SHALL give `|   ` instead of `│   `.
- In gantt charts, the bar of a task SHALL be made of `#` instead of `█`, of an active task of `=` instead of `▓`, of a done task of `-` instead of `░`, and a milestone SHALL be `*` instead of `◆`.
- The frame of diagram source SHALL have `+` in its corners, `-` in its top and bottom lines and `|` at its sides, instead of `┌`, `┐`, `└`, `┘`, `─` and `│`, at the same columns.

The fitting of a diagram to the output width and the choice of wide blocks SHALL follow the same rules as without ASCII mode.

#### Scenario: Flowchart in ASCII mode
- **WHEN** a `mermaid` block with the lines `flowchart TD` and `A --> B` is rendered in ASCII mode
- **THEN** every character of the output is ASCII, the boxes are drawn with `+`, `-` and `|`, and the output has as many lines as without ASCII mode, each of the same width

#### Scenario: Sequence diagram in ASCII mode
- **WHEN** a `mermaid` block with the lines `sequenceDiagram` and `Alice->>Bob: Hi` is rendered in ASCII mode
- **THEN** every character of the output is ASCII, and the output has as many lines as without ASCII mode, each of the same width

#### Scenario: State symbols in ASCII mode
- **WHEN** a state diagram has the lines `state c <<choice>>`, `state f [[fork]]`, `state j <<join>>`, `[*] --> c`, `c --> f : ok`, `f --> j` and `j --> [*]`, and it is rendered in ASCII mode
- **THEN** the output is the same as for the flowchart with the nodes `*`, `<>`, `===`, `===` and `(*)` and the same edges in ASCII mode

#### Scenario: Concurrent regions in ASCII mode
- **WHEN** a state diagram has the lines `[*] --> Active`, `state Active {`, `[*] --> A`, `--`, `[*] --> B` and `}`, and it is rendered in ASCII mode
- **THEN** the diagrams of the regions of `Active` are separated by a blank line, a line of `.` and a blank line

#### Scenario: Links in the text of a state in ASCII mode
- **WHEN** a state diagram has the lines `state "a --> b" as T` and `[*] --> T`, and it is rendered in ASCII mode
- **THEN** the node shows `a - - > b`

#### Scenario: Mindmap in ASCII mode
- **WHEN** a mindmap has the lines `mindmap`, `  root((Plan))`, `    Goals`, `      Faster`, `      Simpler` and `    Risks`, and it is rendered in ASCII mode
- **THEN** the output is `Plan`, `|-- Goals`, `|   |-- Faster`, ``|   `-- Simpler`` and `` `-- Risks ``

#### Scenario: Gantt in ASCII mode
- **WHEN** a gantt chart has the tasks `Design :done, d, 2024-01-01, 10d`, `Build :active, b, after d, 10d`, `Test :t, after b, 10d` and `Release :milestone, after t, 0d`, and it is rendered in ASCII mode
- **THEN** the bar of Design is made of `-`, the bar of Build of `=`, the bar of Test of `#`, and the line of Release has a single `*`

#### Scenario: Framed source in ASCII mode
- **WHEN** a `mermaid` block with the lines `classDiagram` and `    class User` is rendered in ASCII mode
- **THEN** the output is `+- mermaid - classDiagram - not supported +`, `| classDiagram                            |`, `|     class User                          |` and a line of `+`, 41 `-` and `+`

#### Scenario: Text of the diagram in ASCII mode
- **WHEN** a `mermaid` block with the lines `flowchart TD` and `A[Café] --> B` is rendered in ASCII mode
- **THEN** the node shows `Café`, and every other character of the output is ASCII
