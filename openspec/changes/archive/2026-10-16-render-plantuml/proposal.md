# Proposal

## Why

A `plantuml` or `puml` block is shown as its source in a frame labelled "not supported", whatever it contains. The roadmap of README lists PlantUML sequence diagrams. Running PlantUML itself, which has a text output for sequence diagrams, needs Java and PlantUML on the machine; Java is not installed by default on macOS, and a JVM would start for every block. termd already draws mermaid sequence diagrams with mermaid-ascii and mindmaps with a tree layout of its own, and the sequence and mindmap syntax of PlantUML map onto them.

No markdown file of the Go module cache (66 mermaid blocks) has a PlantUML block: GitHub does not render PlantUML in markdown. PlantUML is used in GitLab, wikis and internal documentation.

## What Changes

- termd translates a PlantUML sequence diagram into a mermaid `sequenceDiagram` and draws it with mermaid-ascii, as it draws mermaid sequence diagrams. No external program is run.
- Constructs without an exact mermaid counterpart are approximated:

  | PlantUML | termd |
  |---|---|
  | `participant`, `actor` | the same |
  | `boundary`, `control`, `entity`, `database`, `collections`, `queue` | a participant |
  | `participant "Long name" as L` | the participant `L` labelled `Long name` |
  | `A -> B`, `A --> B`, `A ->> B`, `A -> A` | a solid, dotted or asynchronous message, a self-message |
  | `B <- A` | the message `A -> B` |
  | `->x`, `->o`, `-\`, `-[#red]>`, `<->` | a plain arrow; the decoration is lost |
  | `alt`/`else`, `loop`, `opt`, `break`, `critical` | the same |
  | `par`/`else` | `par`/`and` |
  | `group Retry` ... `end` | a note `Retry` over all participants, without a box |
  | `== Phase 2 ==` | a note `Phase 2` over all participants |
  | `ref over A, B : Login` | a note `ref Login` over `A` and `B` |
  | `note left of`/`right of`/`over`, `hnote`, `rnote`, multi-line `note` ... `end note` | a note; the lines of a multi-line note are joined with spaces |
  | `activate`/`deactivate`, `++`/`--`, `autonumber`, `box` ... `end box`, `title` | the same |
  | `return ok` | the message `ok` back to the caller of the latest activation, which ends |
  | `...`, `|||`, `skinparam`, `hide`, `show`, `header`, `footer`, `legend`, `caption`, `newpage`, `!` directives, `create`, `destroy` | ignored |

- A line that termd cannot read as a sequence construct shows the diagram as framed source labelled `plantuml - sequence - unsupported line N`.
- termd draws a PlantUML mindmap (`@startmindmap`) as the indented tree of mermaid mindmaps. The depth of a node is the number of its `*`, `+` or `-` characters. Colors (`*[#Orange]`), boxless nodes (`*_`) and multi-line nodes (`*:` ... `;`) show their text; `left side` is ignored.
- The type of a PlantUML block comes from its `@start` line. In a `@startuml` block, or a block without `@start`, the keywords of other diagram types (`class`, `interface`, `enum`, `usecase`, `state`, `[*]`, `component`, `start`, an activity `:...;`, `object`, `node`) give that type, which is shown as framed source labelled, for example, `plantuml - class - not supported`; a block with participants or messages is a sequence diagram. `@startgantt`, `@startwbs`, `@startjson` and the other `@start` types are framed source labelled with their type.
- `title` is printed above the diagram, as for mermaid diagrams. `\n` in the text of a message becomes a space, because mermaid-ascii does not wrap message texts; in the text of a mindmap node it continues the text on the next line. Creole markup of PlantUML is shown as written.

Out of scope:

- Running PlantUML or any other external program.
- Other PlantUML diagram types: class, activity, use case, state, component, object, deployment, gantt, wbs and the rest.
- Incoming and outgoing messages (`[-> A`, `A ->]`), which are unsupported lines.
- Creole markup, colors and skin parameters.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `diagram-rendering`: "Diagram block detection" determines the type of a PlantUML block; "Framed source fallback" no longer frames every PlantUML block; new requirements "PlantUML sequence diagrams" and "PlantUML mindmaps".

## Impact

- `internal/diagram`: the type detection and a translator of PlantUML sequence diagrams into mermaid text; a parser of PlantUML mindmaps into the tree of `mindmap.go`; `Render` handles `plantuml` and `puml`.
- `internal/diagram/testdata/plantuml.puml` (`Alice -> Bob: hello`) is now drawn, so its golden file changes; `TestFallbackLabels` takes a PlantUML class diagram for the framed case.
- New golden fixtures in `internal/diagram/testdata` and a new `testdata/plantuml.md` with its golden files.
- `README.md`: "Features", the "Diagrams" section, "Limitations" and "Roadmap".
