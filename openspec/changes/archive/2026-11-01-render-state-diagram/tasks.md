# Tasks

## 1. mermaid-ascii

- [x] 1.1 Add a test to `internal/diagram/diagram_test.go` that `Render("mermaid", "flowchart LR\n    a[\"Tom & Jerry\"] --> b", 200)` has a line with `│ Tom & Jerry ├`; verify `go test ./internal/diagram -run` with its name fails
- [x] 1.2 Update mermaid-ascii to `5a9df1560871` with `go get github.com/AlexanderGrooff/mermaid-ascii@5a9df1560871` and `go mod tidy`; verify the test of 1.1 passes, `make golden` leaves `git status` without changed golden files, and `make check` passes

## 2. classDiagram as the unsupported example

- [x] 2.1 Add `internal/diagram/testdata/class.mmd` (`classDiagram`, `    class User`) and use it in `TestFallbackLabels` with the label parts `mermaid`, `classDiagram` and `not supported` in place of `state.mmd`; switch the `stateDiagram-v2` blocks of `testdata/diagrams.md`, `testdata/control.md` (keep its CRLF line endings and the ESC character inside the block), the "unsupported diagram" case of `internal/render/render_test.go`, `TestUnsupportedDiagramExitsZero` in `cmd/termd/main_test.go` and `internal/site/examples/4-unsupported.md` to `classDiagram` blocks; run `make golden`; verify `git status` shows only the new `class.golden` and the changed `diagrams.*.golden` and `control.*.golden`, which frame a `classDiagram`, and `make check` passes

## 3. Mermaid state diagrams

- [x] 3.1 Add `TestStateDiagram` to `internal/diagram/diagram_test.go` for the scenarios of "State diagram constructs" and the state scenario of "Supported mermaid types": transitions, descriptions, choice, fork and join, compared with `Render("mermaid", flowchart, 80)` of the flowchart of the scenario; ignored lines compared with the diagram without them; `stateDiagram` drawn the same as `stateDiagram-v2`; and the frames labelled `mermaid - stateDiagram-v2 - unsupported line 3`, `mermaid - stateDiagram-v2 - unclosed block on line 2` and `mermaid - stateDiagram - no states`; verify `go test ./internal/diagram -run TestStateDiagram` fails
- [x] 3.2 Add `internal/diagram/state.go` with the model, the mermaid parser and the drawer of a single region of design.md: generated node ids, quoted node texts with `<br>` between description lines, the `‑` and `›` look-alikes, edge labels with `<br>` as spaces; draw the region with `renderFlowchart` and add `stateDiagram` and `stateDiagram-v2` to the switch of `Render`; verify the tests of 3.1 pass
- [x] 3.3 Add `TestStateLayout` for the scenarios of "State diagram layout" and the state scenario of "Diagram width": composite state, transition into a composite state, concurrent regions with the `┈` line, nested composite states in the order `A`, `B`, `C`, direction, notes, links in the text of a state, title, a note longer than the output that wraps with continuation lines that start with 2 spaces, and a `direction LR` diagram that is laid out top-to-bottom at a narrow width; verify `go test ./internal/diagram -run TestStateLayout` fails
- [x] 3.4 Draw the sections of composite states, the regions, the innermost region of a transition, the directions, the notes and the title as in design.md; verify the tests of 3.3 and `TestFlowchartWidthFitting` pass

## 4. PlantUML state diagrams

- [x] 4.1 Add `TestPlantumlState` to `internal/diagram/diagram_test.go` for the scenarios of "PlantUML state diagrams": transitions and arrows, declarations, composite state with regions and special states, compared with `Render("mermaid", ...)` of the mermaid state diagram of the scenario; history, notes, and direction, title and ignored lines; the frames labelled `plantuml - state - unsupported line 3` and `plantuml - state - unclosed block on line 2` for `@startuml`, `state Active {`, `[*] --> Running`, `@enduml`; verify `go test ./internal/diagram -run TestPlantumlState` fails
- [x] 4.2 Add the PlantUML state parser to `internal/diagram/plantuml.go`, building the model of 3.2, and draw the type `state` in `renderPlantUML` with the drawer of `state.go`; verify the tests of 4.1, `TestPlantumlType` and `TestPlantumlSequence` pass

## 5. Golden fixtures

- [x] 5.1 Add `internal/diagram/testdata/state-composite.mmd` with a composite state, two regions, a nested composite state, choice, fork and join, notes and `direction LR`, and `plantuml-state.puml` with the constructs of "PlantUML state diagrams"; add them and `state.mmd` to `TestSupportedTypesAreRendered`; run `go test ./internal/diagram -run TestGolden -update`; verify `git status` shows the two new golden files and a changed `state.golden`, which is now a diagram, and nothing else
- [x] 5.2 Add `testdata/state.md` with a mermaid state diagram with a composite state, regions and notes, a PlantUML state diagram, a state diagram with an unsupported line, and a paragraph after them; run `make golden`; verify `git status testdata/golden` shows only the six new `state.*.golden` files, the diagrams are drawn at every width, and the frame carries its label

## 6. Documentation

- [x] 6.1 In `README.md` add state diagrams to "Features" and to the "Diagrams" section (drawn as flowcharts, the symbols of special states, composite states and regions as separate diagrams, notes listed under the diagrams, PlantUML state diagrams), show a framed `classDiagram` in "Examples", remove `stateDiagram` from "Limitations" and "Roadmap"; verify every statement matches `specs/diagram-rendering/spec.md`

## 7. Verification

- [x] 7.1 Run `make check` and `openspec validate render-state-diagram --strict`; verify both pass
- [x] 7.2 Run `make build`, then `./termd --no-pager --width 80 testdata/state.md` in a terminal and with `| cat`; verify the state diagrams and their sections are drawn, the frame carries its label, and the plain output has no escape sequences
