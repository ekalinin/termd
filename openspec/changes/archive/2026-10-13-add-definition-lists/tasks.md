# Tasks

## 1. Parsing and layout

- [x] 1.1 Extend `TestParseGFM` in `internal/render/render_test.go` with `Term\n: Definition.\n`; the AST must contain an `east.DefinitionList`, an `east.DefinitionTerm` and an `east.DefinitionDescription`; verify `go test ./internal/render -run TestParseGFM` fails
- [x] 1.2 Add `TestDefinitionLists` to `internal/render/render_test.go`, comparing the full plain output at width 40 for the scenarios of the spec: term and definition (`Term`, `    Definition of the term`), several terms and definitions (`Apple`, `Pomme`, `    A fruit`, `    Red or green`), several entries (no blank lines), loose list (a blank line before `Term B` only), definition with several blocks, inline content with hyperlinks disabled (`--width`, `    See docs (https://example.com).`), paragraph lines become terms, colon without a space (`Term :not a definition`) and a definition list in a list item (`Term` at the column of the item text, `Def` 4 columns further); for a long definition check that every line of it starts with 4 spaces and no line is wider than 40; verify `go test ./internal/render -run TestDefinitionLists` fails
- [x] 1.3 In one step, add `extension.DefinitionList` to `Parse` (and its doc comment) and an `east.DefinitionList` case to `renderer.block` in `internal/render/render.go`: terms wrapped like paragraphs with the base style `Bold`, descriptions laid out at `width-4` with their blocks joined by a blank line and every non-empty line prefixed with 4 spaces, blank lines between descriptions and before a term that follows a description only when the list is loose (a description with `IsTight` false or with more than one block), `Wide` from the blocks of the descriptions; verify the tests of 1.1 and 1.2 pass and `go test ./internal/render ./internal/site` passes with unchanged golden files

## 2. Styles

- [x] 2.1 Add `TestDefinitionListStyles` to `internal/render/render_test.go`: in styled mode `Term` is bold (SGR `1`) and the definition line has no bold; with the `dracula` theme and truecolor `Term` is bold without a color; inline code in a term keeps the inline code color of the theme and is bold; verify it passes

## 3. Golden fixture

- [x] 3.1 Add `testdata/definitions.md`: a single entry, two terms sharing two definitions, several entries, a loose list, a definition with a second paragraph and a code block, a definition longer than 80 columns, a term with inline code and a definition with a link and emphasis, a definition list inside a list item and inside a block quote, and a `:not a definition` line; run `make golden`; verify `git status testdata/golden` shows only the six new `definitions.*.golden` files, and in them terms are on their own lines, definitions and their wrapped lines start 4 columns to the right of their term at every width, blank lines appear only in the loose list, and terms are bold in the styled files

## 4. Documentation

- [x] 4.1 In `README.md` add a "Definition lists" subsection to "How it works": the syntax (a paragraph, then lines that start with `: `), every line of the paragraph is a term, terms in bold, definitions indented by 4 columns, when blank lines separate entries, no theme colors a term, and GitHub renders such text as one paragraph; verify every statement matches `specs/markdown-rendering/spec.md`

## 5. Verification

- [x] 5.1 Run `make check` and `openspec validate add-definition-lists --strict`; verify both pass
- [x] 5.2 Run `make build`, then `./termd --no-pager --width 60 testdata/definitions.md` in a terminal and with `| cat`; verify terms are bold only in the terminal, definitions are indented by 4 columns in both, and the plain output has no escape sequences
