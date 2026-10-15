# Tasks

## 1. Parsing and layout

- [x] 1.1 Add `TestMindmap` to `internal/diagram/diagram_test.go`, comparing the lines of `Render("mermaid", src, 80)` for the scenarios of the spec: tree (`Plan`, `├── Goals`, `│   ├── Faster`, `│   └── Simpler`, `└── Risks`), shapes, icons and classes, markdown string, markdown string with a line break, line break in a node (`│   and features`), nearest node with less indentation, comments, and title (`Ideas`, blank, the root); none of them falls back; and the errors: more than one root (a frame labelled `mermaid - mindmap - more than one root`), a shape that is not closed (`unclosed shape`), a markdown string that is not closed (`unclosed markdown string`) and no nodes (`no nodes`), each with `Fallback` set; verify `go test ./internal/diagram -run TestMindmap` fails
- [x] 1.2 Add `internal/diagram/mindmap.go` with the parser, the frontmatter title and the layout of design.md, splitting texts only at their line breaks, and a `case "mindmap"` in `Render`; verify the test of 1.1 passes and `go test ./internal/diagram ./internal/render` passes with unchanged golden files

## 2. Width

- [x] 2.1 Add `TestMindmapWidth` to `internal/diagram/diagram_test.go`: at width 30, the root `R` with the children `Short` and a text of 40 characters with spaces gives `├── Short` and wrapped lines of the second child that start with 4 spaces, no line wider than 30 and `Wide` unset; at width 20, a text of 30 characters with spaces at the fifth level below the root wraps at 10 columns and sets `Wide`; a word longer than the width is not split; verify `go test ./internal/diagram -run TestMindmapWidth` fails
- [x] 2.2 Wrap the texts with `text.Wrap` at `max(width - 4*depth, 10)` columns without splitting words, the continuation lines with the guides of design.md; verify the tests of 2.1 and 1.1 pass

## 3. Golden fixtures

- [x] 3.1 Add `internal/diagram/testdata/mindmap.mmd` with the example of the Mermaid documentation (`root((mindmap))`, `::icon(fa fa-book)`, `On effectiveness<br/>and features`, the children of `Uses` indented by 12 spaces) and add it to `TestSupportedTypesAreRendered`; run `go test ./internal/diagram -run TestGolden -update`; verify `git status` shows only the new `internal/diagram/testdata/mindmap.golden` and that it is the indented tree of the example
- [x] 3.2 Add `testdata/mindmap.md` with the example of the Mermaid documentation, a mindmap with a title, shapes and a markdown string, a mindmap with a text long enough to wrap at 40 and 60 columns, and a mindmap with two roots followed by a paragraph; run `make golden`; verify `git status testdata/golden` shows only the six new `mindmap.*.golden` files, wrapped lines keep the `│` guides at every width, and the broken mindmap is a frame followed by the paragraph

## 4. Documentation

- [x] 4.1 In `README.md` add `mindmap` to "Features" and to the "Diagrams" section (an indented tree, only the text of the nodes, shapes, icons and classes dropped, long texts wrapped), remove `mindmap` from "Limitations" and from "Roadmap"; verify every statement matches `specs/diagram-rendering/spec.md`

## 5. Verification

- [x] 5.1 Run `make check` and `openspec validate render-mindmap --strict`; verify both pass
- [x] 5.2 Run `make build`, then `./termd --no-pager --width 40 testdata/mindmap.md` in a terminal and `./termd --width 80 testdata/mindmap.md | cat`; verify the trees, the wrapped texts under their nodes, the frame of the broken mindmap, and no escape sequences in the pipe
