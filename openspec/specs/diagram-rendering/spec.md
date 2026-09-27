# diagram-rendering Specification

## Purpose

Defines how termd detects diagram code blocks and renders them in the terminal: mermaid sequence, flowchart and ER diagrams are drawn as text, and every other diagram is shown as its source in a labelled frame instead of failing.

## Requirements

### Requirement: Diagram block detection
termd SHALL treat a fenced code block as a diagram when the first word of its info string is `mermaid`, `plantuml` or `puml` (case-insensitive). For mermaid blocks, termd SHALL determine the diagram type from the first line that is not empty, not a `%%` comment and not part of a leading `---` frontmatter block.

#### Scenario: Mermaid block with frontmatter
- **WHEN** a `mermaid` block starts with a `---` frontmatter block containing `title: Login` followed by `sequenceDiagram`
- **THEN** the block is detected as a sequence diagram

#### Scenario: Other fenced code
- **WHEN** a fenced block has the info string `go`
- **THEN** it is rendered as a code block, not as a diagram

### Requirement: Supported mermaid types
termd SHALL render mermaid `sequenceDiagram`, `flowchart`, `graph` and `erDiagram` blocks as text diagrams drawn with Unicode box-drawing characters, in place of the code block.

#### Scenario: Flowchart
- **WHEN** a document contains a `mermaid` block starting with `flowchart LR`
- **THEN** the output shows the nodes as boxes connected by arrows, and the mermaid source is not shown

#### Scenario: Legacy graph keyword
- **WHEN** a `mermaid` block starts with `graph TD`
- **THEN** it is rendered as a flowchart

### Requirement: Sequence diagram constructs
termd SHALL render the following sequence diagram constructs: participants and actors with aliases, solid and dotted messages with and without arrowheads, self-messages, notes (`left of`, `right of`, `over` one or two participants), `loop`, `alt`/`else`, `opt`, `par`/`and`, `critical`, `break` and `rect` blocks, activation via `activate`/`deactivate` and the `+`/`-` message shorthand, `autonumber` and `box` participant groups. Labels with non-ASCII text SHALL keep the diagram aligned.

#### Scenario: Notes and loops
- **WHEN** a sequence diagram contains `Note right of A: проверка токена` and a `loop retry 3x` block with two messages
- **THEN** the note is drawn as a box next to A's lifeline with the Cyrillic text intact, and the loop is drawn as a labelled frame around its two messages

#### Scenario: Autonumber and activation
- **WHEN** a sequence diagram declares `autonumber` and uses `C->>+A: call` and `A-->>-C: result`
- **THEN** the messages are numbered in order and A's lifeline is drawn as active between them

#### Scenario: Alternative branches
- **WHEN** a sequence diagram contains `alt found` ... `else not found` ... `end`
- **THEN** both branches are drawn inside one frame separated by a divider labelled `not found`

### Requirement: Flowchart constructs
termd SHALL render flowchart directions (`TD`, `TB`, `BT`, `LR`, `RL`), subgraphs, edge labels and node shapes, including decision (diamond) nodes.

#### Scenario: Subgraph and labelled edges
- **WHEN** a flowchart contains a `subgraph Parse` with two nodes and a decision node with edges labelled `yes` and `no`
- **THEN** the subgraph is drawn as a labelled frame around its nodes and both edges carry their labels

### Requirement: ER diagram constructs
termd SHALL render ER entities and the relationships between them with their cardinality markers and labels.

#### Scenario: Relationships
- **WHEN** an ER diagram contains `USER ||--o{ DOCUMENT : owns`
- **THEN** USER and DOCUMENT are drawn as boxes connected by a line labelled `owns` with the `||` and `o{` markers at the corresponding ends

### Requirement: Diagram titles
termd SHALL print the `title` from a diagram's frontmatter above the rendered diagram.

#### Scenario: Title from frontmatter
- **WHEN** a mermaid block has frontmatter `title: Login flow`
- **THEN** `Login flow` is printed above the diagram

### Requirement: Diagram width
termd SHALL try to fit a flowchart within the output width by compacting its spacing. If a left-to-right or right-to-left flowchart still does not fit, termd SHALL lay it out top-to-bottom. A diagram that still does not fit, or a sequence or ER diagram wider than the output, SHALL be emitted in full as a wide block.

#### Scenario: Flowchart compacted to fit
- **WHEN** a flowchart is slightly wider than the output width with default spacing but fits with compact spacing
- **THEN** it is rendered with compact spacing within the output width

#### Scenario: Horizontal flowchart re-laid out
- **WHEN** a `flowchart LR` does not fit the output width even with compact spacing
- **THEN** it is rendered top-to-bottom

#### Scenario: Wide sequence diagram
- **WHEN** a sequence diagram with 8 participants is wider than the output width
- **THEN** it is emitted in full as a wide block and can be scrolled horizontally in the pager

### Requirement: Framed source fallback
termd SHALL show the source of a diagram block inside a frame, instead of rendering it, when the diagram language is not mermaid (for example PlantUML), when the mermaid type is not supported (for example `stateDiagram-v2`, `mindmap`, `classDiagram`, `gantt`), or when the diagram cannot be parsed or rendered. The frame SHALL carry a label naming the diagram language, the detected type when known, and the reason. A failure in one diagram SHALL NOT stop the rest of the document from rendering.

#### Scenario: Unsupported mermaid type
- **WHEN** a document contains a `mermaid` block starting with `stateDiagram-v2`
- **THEN** the block's source is shown in a frame labelled with `mermaid`, `stateDiagram-v2` and "not supported", and the following content is rendered normally

#### Scenario: PlantUML block
- **WHEN** a document contains a `plantuml` block
- **THEN** its source is shown in a frame labelled `plantuml`

#### Scenario: Invalid mermaid
- **WHEN** a `mermaid` block starts with `sequenceDiagram` but contains a syntax error
- **THEN** its source is shown in a frame labelled with `mermaid`, `sequenceDiagram` and the parse failure, and termd exits with status 0

#### Scenario: Renderer failure
- **WHEN** rendering a supported diagram fails unexpectedly
- **THEN** the diagram is shown as framed source and the rest of the document is rendered
