# diagram-rendering Specification

## Purpose

Defines how termd detects diagram code blocks and renders them in the terminal: mermaid sequence, flowchart and ER diagrams and PlantUML sequence diagrams are drawn as text, mermaid and PlantUML mindmaps are drawn as indented trees, mermaid gantt charts are drawn as bars on a time axis, and every other diagram is shown as its source in a labelled frame instead of failing.

## Requirements

### Requirement: Diagram block detection
termd SHALL treat a fenced code block as a diagram when the first word of its info string is `mermaid`, `plantuml` or `puml` (case-insensitive). For mermaid blocks, termd SHALL determine the diagram type from the first line that is not empty, not a `%%` comment and not part of a leading `---` frontmatter block. For PlantUML blocks, termd SHALL take the type from the `@start` line: `@startmindmap` is a mindmap, and any other `@start` keyword except `@startuml`, such as `@startgantt`, `@startwbs` or `@startjson`, names its own type. In a `@startuml` block, or a block without a `@start` line, a line that starts with `class `, `interface `, `enum ` or `abstract ` SHALL give the type `class`, `usecase ` the type `usecase`, `state ` or `[*]` the type `state`, `component ` the type `component`, `object ` the type `object`, `node ` the type `deployment`, and a line that is `start` or an activity step `:...;` the type `activity`. Otherwise a block with a participant declaration or a message SHALL be a sequence diagram, and any other block SHALL have no type.

#### Scenario: Mermaid block with frontmatter
- **WHEN** a `mermaid` block starts with a `---` frontmatter block containing `title: Login` followed by `sequenceDiagram`
- **THEN** the block is detected as a sequence diagram

#### Scenario: PlantUML sequence diagram
- **WHEN** a `plantuml` block has the lines `@startuml`, `Alice -> Bob: hello` and `@enduml`, or a `puml` block has only the line `Alice -> Bob: hello`
- **THEN** the block is detected as a sequence diagram

#### Scenario: Other PlantUML types
- **WHEN** a `plantuml` block has the lines `@startuml`, `class User` and `@enduml`, and another one starts with `@startgantt`
- **THEN** the first is detected as a `class` diagram and the second as a `gantt` diagram

#### Scenario: Other fenced code
- **WHEN** a fenced block has the info string `go`
- **THEN** it is rendered as a code block, not as a diagram

### Requirement: Supported mermaid types
termd SHALL render mermaid `sequenceDiagram`, `flowchart`, `graph` and `erDiagram` blocks as text diagrams drawn with Unicode box-drawing characters, mermaid `mindmap` blocks as indented trees, and mermaid `gantt` blocks as bars on a time axis, in place of the code block.

#### Scenario: Flowchart
- **WHEN** a document contains a `mermaid` block starting with `flowchart LR`
- **THEN** the output shows the nodes as boxes connected by arrows, and the mermaid source is not shown

#### Scenario: Legacy graph keyword
- **WHEN** a `mermaid` block starts with `graph TD`
- **THEN** it is rendered as a flowchart

#### Scenario: Mindmap
- **WHEN** a document contains a `mermaid` block starting with `mindmap`
- **THEN** the output shows the nodes as an indented tree, and the mermaid source is not shown

#### Scenario: Gantt
- **WHEN** a document contains a `mermaid` block starting with `gantt`
- **THEN** the output shows the tasks as bars on a time axis, and the mermaid source is not shown

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
termd SHALL print the `title` from a diagram's frontmatter above the rendered diagram. A gantt chart with a `title` line SHALL print the title of that line instead.

#### Scenario: Title from frontmatter
- **WHEN** a mermaid block has frontmatter `title: Login flow`
- **THEN** `Login flow` is printed above the diagram

#### Scenario: Gantt title line
- **WHEN** a `gantt` block has frontmatter `title: Old` and the line `title Plan`
- **THEN** `Plan` is printed above the chart, and `Old` is not printed

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
termd SHALL show the source of a diagram block inside a frame, instead of rendering it, when a PlantUML block is not a sequence diagram or a mindmap (for example a class diagram), when the mermaid type is not supported (for example `stateDiagram-v2`, `classDiagram`), or when the diagram cannot be parsed or rendered. The frame SHALL carry a label naming the diagram language, the detected type when known, and the reason. A failure in one diagram SHALL NOT stop the rest of the document from rendering.

#### Scenario: Unsupported mermaid type
- **WHEN** a document contains a `mermaid` block starting with `stateDiagram-v2`
- **THEN** the block's source is shown in a frame labelled with `mermaid`, `stateDiagram-v2` and "not supported", and the following content is rendered normally

#### Scenario: PlantUML block
- **WHEN** a document contains a `plantuml` block with the lines `@startuml`, `class User` and `@enduml`
- **THEN** its source is shown in a frame labelled with `plantuml`, `class` and "not supported"

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

### Requirement: PlantUML sequence diagrams
termd SHALL draw a PlantUML sequence diagram as the mermaid sequence diagram that it translates to, with the constructs of "Sequence diagram constructs", and SHALL NOT run PlantUML or any other external program. termd SHALL read the lines of the block as follows:

- `participant`, `actor`, `boundary`, `control`, `entity`, `database`, `collections` and `queue` declare a participant, drawn as a mermaid `actor` for `actor` and as a mermaid `participant` otherwise. `"Long name" as L` and `L as "Long name"` give the participant `L` the label `Long name`.
- A message is a sender, an arrow, a receiver and an optional `: text`. An arrow with `--` is dotted, an arrow with `->>` or `-->>` is asynchronous, and any other arrow is solid. An arrow that points only to the left, such as `<-`, swaps the sender and the receiver. The decorations `x`, `o`, `\`, `/`, a color such as `[#red]` and a second arrowhead are ignored. A sender equal to the receiver is a self-message. `++` after the receiver activates it, and `--` after the receiver deactivates the sender.
- `alt`/`else`, `loop`, `opt`, `break` and `critical` with their `end` are the same blocks; the `else` of a `par` block is `and`.
- `group label` and its `end` become a note `label` over all participants, and the messages inside are drawn without a box. `== label ==` becomes a note `label` over all participants. `ref over A, B : label` becomes a note `ref label` over `A` and `B`. A note over all participants spans the first and the last participant of the diagram.
- `note left of`, `note right of` and `note over`, also as `hnote` and `rnote`, are notes. `note left` and `note right` without `of` are notes left of the sender and right of the receiver of the message before them, and `note across` is a note over all participants. A note without `: text` continues on the next lines up to `end note`, and its lines are joined with spaces.
- `activate`, `deactivate`, `autonumber` and `box` ... `end box` are the same as in mermaid. `return text` is a dotted message `text` from the participant of the latest activation that is still active to the sender of the message that activated it, and it deactivates that participant.
- `title text` is printed above the diagram. `\n` in a text becomes a space.
- `@startuml`, `@enduml`, empty lines, `'` comments, `/'` ... `'/` comments, `...` and `... text ...`, `|||` and `||n||`, `skinparam` (also with a `{` ... `}` block), `hide`, `show`, `header`, `footer`, `legend` ... `endlegend`, `caption`, `newpage`, lines that start with `!`, `create` and `destroy` are ignored.

Any other line SHALL make termd show the diagram as framed source labelled with `plantuml`, `sequence` and `unsupported line N`, where N is the number of the line in the block, starting with 1.

#### Scenario: Messages
- **WHEN** a PlantUML sequence diagram has the lines `Alice -> Bob : hello`, `Bob --> Alice : ok`, `Alice ->> Bob : async` and `Alice -> Alice : self`
- **THEN** the output is the same as for the mermaid sequence diagram `Alice->>Bob: hello`, `Bob-->>Alice: ok`, `Alice-)Bob: async` and `Alice->>Alice: self`

#### Scenario: Participants
- **WHEN** a PlantUML sequence diagram has the lines `actor User`, `database "Main DB" as DB` and `User -> DB : query`
- **THEN** the output is the same as for the mermaid sequence diagram `actor User`, `participant DB as Main DB` and `User->>DB: query`

#### Scenario: Reversed and decorated arrows
- **WHEN** a PlantUML sequence diagram has the lines `participant Alice`, `participant Bob`, `Bob <- Alice : hi`, `Alice ->x Bob : lost`, `Alice -[#red]> Bob : red` and `Alice <-> Bob : both`
- **THEN** the output is the same as for the mermaid sequence diagram `participant Alice`, `participant Bob`, `Alice->>Bob: hi`, `Alice->>Bob: lost`, `Alice->>Bob: red` and `Alice->>Bob: both`

#### Scenario: Blocks
- **WHEN** a PlantUML sequence diagram has an `alt valid` block with an `else invalid` part, a `loop 3 times` block and a `par` block with an `else` part, each with a message in every part
- **THEN** the output is the same as for the mermaid sequence diagram with the same blocks, the `else` of the `par` block written as `and`

#### Scenario: Group, divider and reference
- **WHEN** a PlantUML sequence diagram with the participants `A`, `B` and `C` has the lines `group Retry`, `A -> B : try`, `end`, `== Phase 2 ==` and `ref over A, B : Login`
- **THEN** the output is the same as for the mermaid sequence diagram `note over A,C: Retry`, `A->>B: try`, `note over A,C: Phase 2` and `note over A,B: ref Login`

#### Scenario: Multi-line note
- **WHEN** a PlantUML sequence diagram has the lines `note right of B`, `first`, `second` and `end note`
- **THEN** the output is the same as for the mermaid note `note right of B: first second`

#### Scenario: Notes attached to a message
- **WHEN** a PlantUML sequence diagram with the participants `A`, `B` and `C` has the lines `A -> B : call`, `note left : sent`, `note right : received` and `note across : all`
- **THEN** the output is the same as for the mermaid sequence diagram `A->>B: call`, `note left of A: sent`, `note right of B: received` and `note over A,C: all`

#### Scenario: Activation and return
- **WHEN** a PlantUML sequence diagram has the lines `A -> B ++ : call` and `return ok`
- **THEN** the output is the same as for the mermaid sequence diagram `A->>+B: call` and `B-->>-A: ok`

#### Scenario: Ignored lines
- **WHEN** a PlantUML sequence diagram has the lines `skinparam monochrome true`, `skinparam participant {`, `BorderColor red`, `}`, `hide footbox`, `' a comment`, `...`, `|||` and `!theme plain` besides its messages
- **THEN** the output is the same as for the diagram without these lines

#### Scenario: Title and line breaks
- **WHEN** a PlantUML sequence diagram has the lines `title Login` and `A -> B : first\nsecond`
- **THEN** the output starts with `Login` above the diagram, and the message reads `first second`

#### Scenario: Unsupported line
- **WHEN** a `plantuml` block has the lines `@startuml`, `A -> B : call`, `[-> A : in` and `@enduml`
- **THEN** its source is shown in a frame labelled with `plantuml`, `sequence` and `unsupported line 3`

### Requirement: PlantUML mindmaps
termd SHALL draw a PlantUML mindmap, a block with the line `@startmindmap`, as the indented tree of "Mindmap layout". A node line SHALL start with one or more `*`, or with one or more `+` or `-`, and the number of these characters SHALL be the depth of the node, 1 for the root. A node SHALL be a child of the nearest node before it with a smaller depth.

The text of a node SHALL follow the depth characters: a color such as `[#Orange]` and a `_` right after them SHALL be ignored, and a text that starts with `:` SHALL continue on the next lines up to a line that ends with `;`, without the `:` and the `;`. `\n` in a text SHALL continue the text on the next line. `title text` SHALL be printed above the tree. `@startmindmap`, `@endmindmap`, empty lines, `'` comments and `left side` SHALL be ignored.

A mindmap with more than one root, with a `:` text that does not end with `;`, or without nodes, and a line that is none of these SHALL make termd show the block as framed source labelled with `plantuml`, `mindmap` and the reason, such as `more than one root` or `unsupported line N`.

#### Scenario: Stars
- **WHEN** a `plantuml` block has the lines `@startmindmap`, `* Debian`, `** Ubuntu`, `*** Linux Mint`, `*** Kubuntu`, `** LMDE` and `@endmindmap`
- **THEN** the output is `Debian`, `├── Ubuntu`, `│   ├── Linux Mint`, `│   └── Kubuntu` and `└── LMDE`

#### Scenario: Sides
- **WHEN** a PlantUML mindmap has the lines `+ OS`, `++ Linux`, `-- Windows`
- **THEN** the output is `OS`, `├── Linux` and `└── Windows`

#### Scenario: Node decorations
- **WHEN** a PlantUML mindmap has the root `* Root` and the children `**[#Orange] Colored`, `**_ Boxless` and a node that starts with `**:First` and ends with `second;` on the next line
- **THEN** the output is `Root`, `├── Colored`, `├── Boxless`, `└── First` and a continuation line `    second`

#### Scenario: Title
- **WHEN** a PlantUML mindmap has the lines `title Ideas` and `* Root`
- **THEN** the output starts with `Ideas`, a blank line and `Root`

#### Scenario: More than one root
- **WHEN** a PlantUML mindmap has the lines `* A` and `* B`
- **THEN** its source is shown in a frame labelled with `plantuml`, `mindmap` and `more than one root`

### Requirement: Gantt task dates
termd SHALL read the lines of a mermaid `gantt` block after its `gantt` line as follows:

- `dateFormat F` sets the format of the dates of the block, `YYYY-MM-DD` by default. F SHALL be made of the tokens `YYYY`, `YY`, `MM`, `M`, `DD`, `D`, `HH`, `H`, `mm`, `m`, `ss` and `s`, and of other characters that stand for themselves, or F SHALL be `X`, seconds since 1970-01-01, or `x`, milliseconds since 1970-01-01. `M`, `D`, `H`, `m` and `s` accept one or two digits, the other tokens their full number of digits, and `YY` is a year from 1969 to 2068. The parts of a date that F does not have SHALL be those of 1970-01-01 00:00:00. Dates SHALL be read and shown in UTC, so the output does not depend on the time zone of the machine.
- `title text`, `section name` and `axisFormat` are used by "Gantt layout".
- `excludes` and `includes` name days by tokens separated by spaces or commas, in any case: `weekends`, weekday names such as `sunday`, and dates written in the `dateFormat` or as `YYYY-MM-DD`. Several `excludes` and `includes` lines add up. `weekend friday` makes the weekend Friday and Saturday, and `weekend saturday` Saturday and Sunday, the default.
- `inclusiveEndDates` makes every end given as a date one day later.
- `click`, `todayMarker`, `tickInterval`, `weekday`, `topAxis`, `accTitle:`, `accDescr:`, `accDescr {` ... `}`, empty lines and lines that start with `%%` are ignored.
- Any other line with a `:` is a task: the text before the first `:` is its name, and the text after it is a list of items separated by commas. The items SHALL start with any of the tags `done`, `active`, `crit`, `milestone` and `vert`, followed by one, two or three items: an end; a start and an end; or an id, a start and an end.

The start of a task SHALL be a date, or `after` followed by ids separated by spaces, which is the latest end of these tasks. A task without a start SHALL start at the end of the task before it in the block, in any section. The end of a task SHALL be a date; `until` followed by ids, which is the earliest start of these tasks; or a duration, a number with an optional fractional part followed by `ms`, `s`, `m`, `h`, `d`, `w`, `M` or `y`, added to the start. `d` and `w` SHALL add days, the number of days rounded to the nearest whole number; `M` and `y` SHALL add calendar months and years without the fractional part, ending on the last day of the month when the month has fewer days; the other units SHALL add exact time. An end that is not a date, an `until` or a duration SHALL be the start of the task. An id SHALL name a task before or after the task that uses it.

A day SHALL be excluded when `excludes` names its date or its weekday, or has `weekends` and the day is a day of the weekend, unless `includes` names its date. When the end of a task is a duration and a day is excluded, termd SHALL check the days after the start one by one up to the end, and every excluded day SHALL move the end one day later. The tasks that start at the end of this task SHALL start at the moved end, and its bar SHALL end before the excluded days right before the moved end.

termd SHALL show the block as framed source labelled with `mermaid`, `gantt` and the reason:

- `unsupported line N` for any other line, a task without items after its tags, and a task with more than three of them;
- `invalid date on line N` for a start that is not a date in the `dateFormat` and not an `after`;
- `unsupported dateFormat` for a `dateFormat` with other tokens of Mermaid, such as `MMM`, `Q`, `Do`, `DDD`, `hh`, `A`, `SSS` or `Z`;
- `no start date on line N` for a task without a start and without a task before it;
- `unknown task` followed by the id for an `after` or an `until` with an id that no task has;
- `circular dependency` for tasks whose starts or ends depend on each other;
- `no tasks` for a block without tasks, not counting the tasks with the `vert` tag.

N SHALL be the number of the line in the block, starting with 1.

#### Scenario: Task without a start
- **WHEN** a gantt chart has the tasks `A :a, 2024-01-01, 10d` and `B :10d`
- **THEN** the axis shows `2024-01-01` and `2024-01-21`, and the bar of B starts in the column where the bar of A ends

#### Scenario: After several tasks
- **WHEN** a gantt chart has the tasks `A :a, 2024-01-01, 5d`, `B :b, 2024-01-01, 10d` and `C :after a b, 10d`
- **THEN** the axis shows `2024-01-01` and `2024-01-21`, and the bar of C starts in the column where the bar of B ends

#### Scenario: Until a later task
- **WHEN** a gantt chart has the tasks `A :a, 2024-01-01, until b` and `B :b, 2024-01-11, 10d`
- **THEN** the axis shows `2024-01-01` and `2024-01-21`, and the bar of A ends in the column where the bar of B starts

#### Scenario: Durations
- **WHEN** gantt charts with `dateFormat YYYY-MM-DD` have the single tasks `A :2024-01-01, 1w`, `A :2024-01-01, 1.5d`, `A :2024-01-31, 1M` and `A :2024-01-01, 36h`
- **THEN** their axes end at `2024-01-08`, `2024-01-03`, `2024-02-29` and `2024-01-02`

#### Scenario: Times
- **WHEN** a gantt chart has the lines `dateFormat HH:mm`, `axisFormat %H:%M`, `A :a, 17:49, 10m` and `B :5m`
- **THEN** the axis shows `17:49` and `18:04`

#### Scenario: Unix timestamps
- **WHEN** a gantt chart has the lines `dateFormat X`, `axisFormat %s` and `A :0, 71`
- **THEN** the axis shows `0` and `71`

#### Scenario: Excluded weekend
- **WHEN** a gantt chart has the lines `excludes weekends`, `A :a, 2024-01-01, 5d` and `B :after a, 1d`
- **THEN** the axis shows `2024-01-01` and `2024-01-09`, the bar of A ends in the column of `2024-01-06`, the bar of B starts in the column of `2024-01-08`, and the columns between them are empty

#### Scenario: Included day
- **WHEN** a gantt chart has the lines `excludes weekends`, `includes 2024-01-06` and `A :2024-01-01, 6d`
- **THEN** the axis ends at `2024-01-08`

#### Scenario: Inclusive end dates
- **WHEN** a gantt chart has the lines `inclusiveEndDates` and `A :2024-01-01, 2024-01-10`
- **THEN** the axis ends at `2024-01-11`

#### Scenario: Ignored lines
- **WHEN** a gantt chart has the lines `todayMarker off`, `tickInterval 1week`, `weekday monday`, `topAxis`, `accTitle: Plan`, `click a href "https://example.com"` and `%% a comment` besides its tasks
- **THEN** the output is the same as for the chart without these lines

#### Scenario: Unsupported line
- **WHEN** a `mermaid` block has the lines `gantt`, `dateFormat YYYY-MM-DD` and `A 2024-01-01, 3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `unsupported line 3`

#### Scenario: Invalid date
- **WHEN** a `mermaid` block has the lines `gantt` and `A :2024-13-01, 3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `invalid date on line 2`

#### Scenario: Unsupported date format
- **WHEN** a `mermaid` block has the lines `gantt`, `dateFormat MMM D YYYY` and `A :Jan 1 2024, 3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `unsupported dateFormat`

#### Scenario: No start date
- **WHEN** a `mermaid` block has the lines `gantt` and `A :3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `no start date on line 2`

#### Scenario: Unknown task
- **WHEN** a gantt chart has the task `A :after x, 3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `unknown task x`

#### Scenario: Circular dependency
- **WHEN** a gantt chart has the tasks `A :a, after b, 1d` and `B :b, after a, 1d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `circular dependency`

#### Scenario: No tasks
- **WHEN** a `mermaid` block has the lines `gantt` and `title Plan`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `no tasks`

### Requirement: Gantt layout
termd SHALL draw a mermaid `gantt` block as lines without styles:

- The title, from the `title` line or else from the frontmatter, followed by a blank line.
- The axis line: the start of the chart, the earliest start of a task, at the first column of the bars, and the end of the chart, the latest end of a task, ending at the last column of the bars. Both SHALL be formatted with `axisFormat`, `%Y-%m-%d` by default. When the bars have fewer columns than the two dates and a space between them, only the start SHALL be shown.
- Every section with tasks: its name on its own line, wrapped at the output width, followed by its tasks in the order of the block. The tasks before the first section SHALL come first, without a section line.
- Every task on its own line: its name, indented by 2 spaces in a section, padded to the width of the names column, 2 spaces, and its bar. The names column SHALL be as wide as the widest name with its indentation.

The bars SHALL take the columns from the end of the names column and its 2 spaces up to the output width. They SHALL map the time from the start to the end of the chart evenly onto their columns: the instant t SHALL fall at column round(C × (t − start) / (end − start)), counting from 0, where C is the number of columns of the bars, and at column 0 when the start and the end of the chart are the same. A bar SHALL fill the columns from the column of the start of its task up to, not including, the column of the end of its bar, and at least one column; a bar that starts at column C SHALL fill column C − 1. A bar SHALL be made of `▓` for an `active` task, `░` for a `done` task that is not `active`, and `█` for any other task. A `milestone` SHALL be a single `◆` at the column of the middle between its start and its end, and at column C − 1 when that column is C. A task with the `vert` tag SHALL NOT be drawn and SHALL NOT take a line. The `crit` tag SHALL NOT change the bar.

When the bars would have fewer than 20 columns, the names column SHALL shrink to leave them 20 columns, but SHALL NOT be narrower than 10 columns, and the bars SHALL have at least 20 columns. A name wider than its column SHALL wrap at whitespace, a word wider than the column SHALL be split, and the continuation lines SHALL start at the column of the name and have no bar. A chart that is still wider than the output SHALL be emitted in full as a wide block.

`axisFormat` SHALL support `%Y`, `%y`, `%m`, `%d`, `%e`, `%H`, `%I`, `%M`, `%S`, `%L`, `%p`, `%a`, `%A`, `%b`, `%B`, `%j`, `%s` and `%%` as in Mermaid, with English names of months and weekdays, and a `-` after the `%` SHALL drop the padding of a number. Any other `%` code SHALL be shown as written.

#### Scenario: Bars
- **WHEN** a gantt chart with the lines `dateFormat YYYY-MM-DD`, `section Build`, `Design :d, 2024-01-01, 10d` and `Code :after d, 10d` is rendered at width 40
- **THEN** the output is the axis line with `2024-01-01` at columns 11 to 20 and `2024-01-21` at columns 31 to 40, the line `Build`, the line `  Design` followed by 2 spaces and 15 `█`, and the line `  Code` followed by 19 spaces and 15 `█`

#### Scenario: Title
- **WHEN** a gantt chart has the line `title Plan`
- **THEN** the output starts with `Plan`, a blank line and the axis line

#### Scenario: Task states
- **WHEN** a gantt chart has the tasks `Design :done, d, 2024-01-01, 10d`, `Build :active, b, after d, 10d`, `Test :t, after b, 10d`, `Fix :crit, f, after t, 10d` and `Release :milestone, after f, 0d`
- **THEN** the bar of Design is made of `░`, the bar of Build of `▓`, the bars of Test and Fix of `█`, and the line of Release has a single `◆` in the last column of the bars

#### Scenario: Tasks without a section
- **WHEN** a gantt chart has the task `Kickoff :k, 2024-01-01, 1d`, then `section Work` and `Build :after k, 5d`
- **THEN** the line of Kickoff comes first and starts with `Kickoff`, followed by the line `Work` and the line of Build, which starts with `  Build`

#### Scenario: Vertical marker
- **WHEN** a gantt chart has the tasks `A :a, 2024-01-01, 10d` and `Deadline :vert, 2024-01-05, 0d`
- **THEN** the output has no line for Deadline

#### Scenario: Long names
- **WHEN** a gantt chart has a section with the task `Implement the parser and the tests :2024-01-01, 5d` and is rendered at width 40
- **THEN** the names column is 18 columns wide, the name wraps with the bar on its first line, the bars take 20 columns, and no line is wider than 40 columns

#### Scenario: Narrow output
- **WHEN** a gantt chart with a section and the task `Implement the parser :2024-01-01, 5d` is rendered at width 25
- **THEN** the names column is 10 columns wide, the bars take 20 columns, and the chart is emitted as a wide block

#### Scenario: Axis format
- **WHEN** a gantt chart with the line `axisFormat %-d %b` has the task `A :2024-01-01, 20d`
- **THEN** the axis shows `1 Jan` and `21 Jan`

#### Scenario: Narrow bars
- **WHEN** a gantt chart with the line `axisFormat %Y-%m-%dT%H:%M` and the task `A :2024-01-01, 5d` is rendered at width 30
- **THEN** the axis line shows only the start of the chart
