# Tasks

## 1. Type detection

- [x] 1.1 Add `TestPlantumlType` to `internal/diagram/diagram_test.go` for the detection rules of the spec: `@startuml`, `Alice -> Bob: hello`, `@enduml` and a block with only `Alice -> Bob: hello` are `sequence`, `participant A` alone is `sequence`, `class User`, `interface I`, `enum E` and `abstract A` are `class`, `usecase U` is `usecase`, `state S` and `[*] --> A` are `state`, `component C` is `component`, `object O` is `object`, `node N` is `deployment`, `start` and `:step;` are `activity`, `@startmindmap` is `mindmap`, `@startgantt` is `gantt`, `@startwbs` is `wbs`, and a block with only `skinparam monochrome true` has no type; verify `go test ./internal/diagram -run TestPlantumlType` fails
- [x] 1.2 Add `internal/diagram/plantuml.go` with `plantumlType` and switch `Render` on it for `plantuml` and `puml`: every type is framed source labelled with the language, the type and `not supported` for now, and a block without a type keeps the label `plantuml - not supported`; add `internal/diagram/testdata/plantuml-class.puml` (`@startuml`, `class User`, `@enduml`) and use it in `TestFallbackLabels` with the label parts `plantuml`, `class` and `not supported`; verify the tests pass, and `go test ./internal/diagram -run TestGolden` fails only for `plantuml`, whose label now names `sequence`

## 2. Sequence diagrams

- [x] 2.1 Add `TestPlantumlSequence` to `internal/diagram/diagram_test.go`: for every sequence scenario of the spec (messages, participants, reversed and decorated arrows, blocks, group, divider and reference, multi-line note, notes attached to a message, activation and return, ignored lines, title and line breaks), the lines of `Render("plantuml", src, 80)` equal the lines of `Render("mermaid", mermaid, 80)` for its mermaid counterpart, and none falls back; `@startuml`, `A -> B : call`, `[-> A : in`, `@enduml` is a frame labelled `plantuml - sequence - unsupported line 3`; verify `go test ./internal/diagram -run TestPlantumlSequence` fails
- [x] 2.2 Translate PlantUML sequence diagrams into mermaid text in `plantuml.go` with the two passes of design.md, and draw them with `safeRender`; verify the tests of 2.1 pass

## 3. Mindmaps

- [x] 3.1 Add `TestPlantumlMindmap` to `internal/diagram/diagram_test.go` for the mindmap scenarios of the spec: stars (`Debian`, `├── Ubuntu`, `│   ├── Linux Mint`, `│   └── Kubuntu`, `└── LMDE`), sides, node decorations, title, and the errors `more than one root`, a `:` text without `;` and a line that is not a node (`unsupported line N`), each a frame labelled `plantuml - mindmap - <reason>`; verify `go test ./internal/diagram -run TestPlantumlMindmap` fails
- [x] 3.2 Split `renderMindmap` into the mermaid parser and a function that prints a title and lays out a `mindNode` tree, and add `parsePlantMindmap` with the depth of `*`, `+` and `-`; verify the tests of 3.1, `TestMindmap` and `TestMindmapWidth` pass

## 4. Golden fixtures

- [x] 4.1 Add `internal/diagram/testdata/plantuml-seq.puml` with a sequence diagram that uses every construct of the spec and `plantuml-mindmap.puml` with the Debian example, sides and node decorations; make `TestSupportedTypesAreRendered` render `.puml` fixtures as `plantuml` and add both; run `go test ./internal/diagram -run TestGolden -update`; verify `git status` shows the two new golden files and a changed `plantuml.golden`, which is now the diagram of `Alice -> Bob: hello`, and nothing else
- [x] 4.2 Add `testdata/plantuml.md` with a `plantuml` sequence diagram, a `puml` block without `@start` lines, a mindmap, a class diagram and a sequence diagram with an unsupported line, followed by a paragraph; run `make golden`; verify `git status testdata/golden` shows only the six new `plantuml.*.golden` files, the diagrams are drawn at every width, and the two frames carry their labels

## 5. Documentation

- [x] 5.1 In `README.md` add PlantUML sequence diagrams and mindmaps to "Features" and to the "Diagrams" section (translated into mermaid without external programs, the approximations of the spec in short, other PlantUML types shown as source with their type), and update "Limitations" and "Roadmap"; verify every statement matches `specs/diagram-rendering/spec.md`

## 6. Verification

- [x] 6.1 Run `make check` and `openspec validate render-plantuml --strict`; verify both pass
- [x] 6.2 Run `make build`, then `./termd --no-pager --width 80 testdata/plantuml.md` in a terminal and with `| cat`; verify the sequence diagram and the mindmap are drawn, the frames carry their labels, and the plain output has no escape sequences
