# Design

## Context

`diagram.Render` returns `fallback(src, width, lang, "not supported")` for every language other than `mermaid`, so a `plantuml` or `puml` block is always framed source. Mermaid sequence diagrams go through `safeRender` into mermaid-ascii, which prints the frontmatter `title` above the diagram; `internal/diagram/mindmap.go` parses mermaid mindmaps into a `mindNode` tree and lays it out with `mindNode.layout`. mermaid-ascii also exports `sequence.Parse` and `sequence.Render` over its own AST, with a fixed set of fragment types (`loop`, `opt`, `alt`, `par`, `critical`, `break`, `rect`) whose label always starts with the keyword.

Spikes with the mermaid-ascii version of termd showed:

- A PlantUML login flow translated by hand into a mermaid `sequenceDiagram` (an actor, `alt`/`else`, activation, a note, a title) draws within 80 columns.
- `-)` draws an asynchronous arrow, `par` ... `and` works, and `note over A,C` spans the participants from `A` to `C`, which serves as a divider.
- `<br/>` in a message text is printed as written; in a note it becomes a space.
- `rect` is labelled `[rect]` without a text, and `critical X` is labelled `[critical X]`, so no fragment type is neutral for a PlantUML `group`.

Java is not installed on the development machine, and no PlantUML block is in the markdown files of the module cache. The requirements are in `specs/diagram-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- PlantUML sequence diagrams and mindmaps drawn without any program besides termd, by the same code that draws their mermaid counterparts.
- Every PlantUML block keeps a precise frame label when it is not drawn: its type, or the line termd could not read.
- Mermaid diagrams render byte for byte as before.

**Non-Goals:**
- Parity with the PlantUML grammar. Lines outside the spec are unsupported lines.
- Other PlantUML diagram types.

## Decisions

### Translation into mermaid text, not into the AST of mermaid-ascii

`internal/diagram/plantuml.go` translates a PlantUML sequence diagram into the text of a mermaid `sequenceDiagram`, with the title as mermaid frontmatter, and passes it to `safeRender` like a mermaid block. A translation error, the first line it cannot read, becomes the frame label through `fallback`.

Alternatives:
- Building a `sequence.SequenceDiagram` and calling `sequence.Render`. Rejected: termd would depend on the AST of the library, which changes between commits, instead of its input language; text keeps the translation testable against a mermaid source, which the scenarios of the spec use.
- Running `plantuml -tutxt`. Rejected with the user: it needs Java and PlantUML, starts a JVM per block, and cannot run in CI without Java.

### Type detection

`plantumlType(src)` reads the `@start` line: `@startmindmap` is `mindmap`, `@startuml` or no `@start` line goes on, any other `@start<x>` is `<x>`. For `@startuml` it scans the lines for the keywords of the spec in one pass; the first keyword of another type wins, as PlantUML picks the type from such keywords. Without one, a participant declaration or a line with an arrow makes the block `sequence`; anything else is `""`, which keeps today's label `plantuml - not supported`.

`Render` switches on the language: `mermaid` as before; `plantuml` and `puml` on `plantumlType`; the label of a frame keeps the language as written in the info string, as it does today.

### Two passes over a sequence diagram

The first pass collects the participants in the order of their first appearance, in a declaration, a message or a note, with the `box` around them and their kind and label. The second pass writes the mermaid text: the frontmatter title, `sequenceDiagram`, the declarations of all participants in that order, each `box` ... `end` around its participants, then the translated lines. Declaring every participant first fixes the order of the columns and lets a note over all participants name the first and the last of them, even when the note comes before a participant is used.

The other lines translate one by one:

- A message is matched by a regular expression: a sender, an arrow, a receiver, optional activation marks and an optional `: text`. The arrow is a run of `-`, `<`, `>`, `/`, `\`, `x` and `o` around one or two `-`, with an optional `[...]` color; `x` and `o` count as decorations only next to an arrowhead or a dash. The arrow gives the mermaid arrow (`->>`, `-->>`, `-)`, `--)`), its direction gives the order of the participants. Participant names are identifiers or quoted strings.
- `group`, `== ==`, `ref over` and `note across` become `note over first,last` or `note over A,B`; the `end` of a `group` is dropped, which needs a stack of open blocks to tell it from the `end` of `alt` or `loop`.
- A multi-line `note` collects its lines up to `end note` and joins them with spaces.
- `return` needs the activations: a stack of the activated participant and the sender of the message that activated it, pushed by `++` and `activate`, popped by `--`, `deactivate` and `return`.
- The ignored lines of the spec are skipped, including the `{` ... `}` block of `skinparam`, `legend` ... `endlegend` and `/'` ... `'/` comments.

### PlantUML mindmaps share the tree of mermaid mindmaps

`parsePlantMindmap` builds the same `mindNode` tree: the depth comes from the count of `*`, `+` or `-`, a stack of open nodes gives the parent, and `:` ... `;` texts and `\n` give the lines of a node. `renderMindmap` is split into the mermaid parser and a function that prints a title and lays out a tree, which both parsers use, so wrapping, guides, width and the `more than one root` error stay the same.

### Fixtures

`internal/diagram/testdata/plantuml.puml` (`Alice -> Bob: hello`) becomes a drawn sequence diagram, so its golden file changes, and `TestFallbackLabels` uses a new `plantuml-class.puml` for the framed case. New `.puml` fixtures cover a sequence diagram with every construct of the spec and the mindmap examples; `TestGolden` already renders `.puml` files as `plantuml`.

## Risks / Trade-offs

- [The arrow grammar of PlantUML is large; an arrow form outside the expression is an unsupported line and frames the whole diagram] → The frame names the line; forms can be added to the expression without changing the design.
- [`group` loses its box and `->x` its cross] → Accepted with the user: a wrong keyword such as `[critical Retry]` would mislead more than a missing box.
- [Hoisting the participant declarations to the top can reorder a participant that PlantUML would place by its position inside a `box`] → PlantUML requires the participants of a box to be declared together, and the first pass keeps them together.
- [No real PlantUML documents in the sample] → The fixtures follow the examples of the PlantUML documentation.

## Migration Plan

The golden file of `plantuml.puml` changes from a frame to a diagram, and `TestFallbackLabels` changes its PlantUML case. Mermaid output does not change. Rollback is a revert of the change.
