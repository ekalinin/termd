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
termd SHALL render mermaid `sequenceDiagram`, `flowchart`, `graph` and `erDiagram` blocks as text diagrams drawn with Unicode box-drawing characters, and mermaid `mindmap` blocks as indented trees, in place of the code block.

#### Scenario: Flowchart
- **WHEN** a document contains a `mermaid` block starting with `flowchart LR`
- **THEN** the output shows the nodes as boxes connected by arrows, and the mermaid source is not shown

#### Scenario: Legacy graph keyword
- **WHEN** a `mermaid` block starts with `graph TD`
- **THEN** it is rendered as a flowchart

#### Scenario: Mindmap
- **WHEN** a document contains a `mermaid` block starting with `mindmap`
- **THEN** the output shows the nodes as an indented tree, and the mermaid source is not shown

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
termd SHALL show the source of a diagram block inside a frame, instead of rendering it, when the diagram language is not mermaid (for example PlantUML), when the mermaid type is not supported (for example `stateDiagram-v2`, `classDiagram`, `gantt`), or when the diagram cannot be parsed or rendered. The frame SHALL carry a label naming the diagram language, the detected type when known, and the reason. A failure in one diagram SHALL NOT stop the rest of the document from rendering.

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

### Requirement: Mindmap layout
termd SHALL draw a mermaid `mindmap` as an indented tree: the root on the first line without a prefix, and every other node on its own line after `├── `, or after `└── ` when it is the last child of its parent. A line of a node SHALL start with 4 columns for every ancestor below the root: `│   ` when that ancestor has a later sibling, and 4 spaces otherwise.

The hierarchy SHALL follow the indentation of the source: a node SHALL be a child of the nearest node before it with less indentation, and a tab SHALL count as 4 spaces. Empty lines and `%%` comment lines SHALL be ignored.

A node line SHALL be either its text, or an optional id followed by a shape that ends the line: `[text]`, `(text)`, `((text))`, `))text((`, `)text(` or `{{text}}`. termd SHALL show only the text of a node, without its id and its shape. A line that starts with `::icon(` or `:::` SHALL be ignored. The text of a markdown string, a text between `"` and a backtick and a backtick and `"`, SHALL be shown without these characters and without the `**` and `*` markers, and it can span several source lines. The output SHALL have no styles.

`<br>`, `<br/>`, `<br />` and a line break inside a markdown string SHALL continue the text of the node on the next line. A text longer than the remaining width SHALL wrap at whitespace. A continuation line SHALL start at the column of the text of its node, with the same `│` guides as the lines of the children of the node, and the text SHALL never be narrower than 10 columns. A mindmap wider than the output SHALL be emitted in full as a wide block.

A `title` from the frontmatter of the block SHALL be printed above the tree, followed by a blank line. A mindmap with more than one root, with a shape or a markdown string that is not closed, or without nodes SHALL be shown as its source in a frame labelled with `mermaid`, `mindmap` and the reason.

#### Scenario: Tree
- **WHEN** a mindmap has the lines `mindmap`, `  root((Plan))`, `    Goals`, `      Faster`, `      Simpler` and `    Risks`
- **THEN** the output is `Plan`, `├── Goals`, `│   ├── Faster`, `│   └── Simpler` and `└── Risks`

#### Scenario: Shapes
- **WHEN** a mindmap has the root `root((A))` and the children `b[Square]`, `c(Rounded)`, `d))Bang((`, `e)Cloud(` and `f{{Hexagon}}`
- **THEN** the output is `A`, `├── Square`, `├── Rounded`, `├── Bang`, `├── Cloud` and `└── Hexagon`

#### Scenario: Icons and classes
- **WHEN** a mindmap has the root `R`, the child `Child`, and after it the lines `::icon(fa fa-book)` and `:::urgent large`
- **THEN** the output is `R` and `└── Child`

#### Scenario: Markdown string
- **WHEN** a mindmap has the root ``root["`**Root** text`"]``
- **THEN** the output is `Root text`

#### Scenario: Markdown string with a line break
- **WHEN** a mindmap has a root whose markdown string starts with ``["`First line`` on one source line and ends with ``second line`"]`` on the next
- **THEN** the output is `First line` followed by `second line`

#### Scenario: Line break in a node
- **WHEN** a mindmap has the root `R` and the children `On effectiveness<br/>and features` and `Next`
- **THEN** the output is `R`, `├── On effectiveness`, `│   and features` and `└── Next`

#### Scenario: Nearest node with less indentation
- **WHEN** a mindmap has the lines `mindmap`, `  R`, `    A`, `        B` and `      C`
- **THEN** `B` and `C` are children of `A`, and the output is `R`, `└── A`, `    ├── B` and `    └── C`

#### Scenario: Long text
- **WHEN** a mindmap has the root `R` and the children `Short` and a text of 40 characters with spaces, and it is rendered at width 30
- **THEN** the text of the second child wraps, its continuation lines start with 4 spaces, the line of `Short` starts with `├── `, and no line is wider than 30 columns

#### Scenario: Deep tree in a narrow output
- **WHEN** a mindmap has a node at the fifth level below the root with a text of 30 characters with spaces, and it is rendered at width 20
- **THEN** the text of that node wraps at 10 columns and the diagram is emitted as a wide block

#### Scenario: Title
- **WHEN** a mindmap block has the frontmatter `title: Ideas`
- **THEN** the output starts with `Ideas`, a blank line and the root of the tree

#### Scenario: Comments
- **WHEN** a mindmap has the lines `mindmap`, `  %% a comment`, `  R` and `    A`
- **THEN** the output is `R` and `└── A`

#### Scenario: More than one root
- **WHEN** a mindmap has the lines `mindmap`, `  A` and `  B`
- **THEN** its source is shown in a frame labelled with `mermaid`, `mindmap` and `more than one root`

#### Scenario: Shape that is not closed
- **WHEN** a mindmap has the root `root((Plan)`
- **THEN** its source is shown in a frame labelled with `mermaid`, `mindmap` and the reason

#### Scenario: No nodes
- **WHEN** a mindmap block has only the line `mindmap`
- **THEN** its source is shown in a frame labelled with `mermaid`, `mindmap` and the reason
