# Tasks

## 1. Baseline

- [x] 1.1 Run `make build` on `main` and record the size of `./termd` (`ls -l termd`); verify the number is written down for the PR description (baseline: 11014162 bytes)

## 2. Headerless table

- [x] 2.1 Add `TestHeaderless` to `internal/table/table_test.go`: a `Table` with a nil `Header` and the rows `title | Doc` and `description | <30 words>` rendered at width 40 has no rule line, its first line is the `title` row, the separators are at the same column on every line and no line is wider than 40; verify `go test ./internal/table` fails
- [x] 2.2 In `internal/table/table.go` take the number of columns from the widest row when `Header` is nil, measure and allocate widths over the rows only, and skip the header line and the rule in `Render`; verify `go test ./internal/table` passes and `go test ./internal/render ./internal/diagram ./internal/site` passes with unchanged golden files

## 3. Frontmatter detection

- [x] 3.1 Add tests for `splitFrontmatter` in `internal/render/render_test.go`, one case per rule of the design: a mapping, a `...` closing line, an empty block, a comment-only block, a scalar block (`Some text`), no closing line, a `---` block after a paragraph, a BOM with CRLF line endings, and invalid YAML (`title: [unclosed`); each case checks whether a table, a frame, nothing or the original source comes out, and the body that remains; verify `go test ./internal/render` fails
- [x] 3.2 Create `internal/render/frontmatter.go` with `splitFrontmatter`, parsing the block into a `yaml.Node` with `go.yaml.in/yaml/v3` under `recover`; add the module with `go get go.yaml.in/yaml/v3@v3.0.5`; verify the tests of 3.1 pass and `make tidy-check` passes

## 4. Frontmatter rendering

- [x] 4.1 Add value formatting tests to `internal/render/render_test.go`: a quoted scalar (`title: "Doc: x"` gives `Doc: x`), a `|` string (two lines in the cell), `tags: [a, b]` (`a, b`), a mapping of scalars (`name: Eugene` and `url: https://example.com` on two lines), a list of mappings (`[{name: A}, {name: B}]`) and a key without a value (empty cell); in styled mode the key is bold, in plain mode the output has no escape sequences; verify `go test ./internal/render` fails
- [x] 4.2 In `render.Render` call `splitFrontmatter` before `Parse`, render the body as `src`, and prepend the frontmatter table (bold keys, value cells per the design table) as the first block; verify the tests of 4.1 pass and the existing golden files do not change
- [x] 4.3 Add a test that `---`, `title: [unclosed`, `---`, `# Hello` renders `title: [unclosed` inside a frame labelled `frontmatter - invalid YAML` followed by `# Hello`, then show the block through `diagram.Frame`, marking it wide when the frame is wider than the output; verify `go test ./internal/render` passes
- [x] 4.4 Add a case to `cmd/termd/main_test.go`: a file with invalid frontmatter exits with status 0 and the output contains the frame label; verify `go test ./cmd/termd` passes

## 5. Golden fixture

- [x] 5.1 Add `testdata/frontmatter.md`: frontmatter with `title`, a quoted value with a colon, `date`, `draft`, `tags` as a list, `author` as a mapping, `authors` as a list of mappings, a multi-line `description` and a 30-word `summary`, followed by a heading and a paragraph; run `make golden`; verify `git status testdata/golden` shows only the six new `frontmatter.*.golden` files, and at width 40 the key column is on one line, the long values wrap and the styled keys are bold

## 6. Documentation

- [x] 6.1 In `README.md` add a frontmatter bullet to Features and a "Frontmatter" subsection to "How it works" (a YAML mapping becomes a key-value table, invalid YAML is shown in a frame, other blocks render as markdown), and note in Limitations that TOML (`+++`) frontmatter is rendered as markdown; verify every statement matches `specs/markdown-rendering/spec.md`

## 7. Verification

- [x] 7.1 Run `make check`; verify it passes
- [x] 7.2 Run `make build`, compare the size of `./termd` with 1.1 and note the difference for the PR description; open a document with frontmatter in a terminal (`./termd testdata/frontmatter.md`) and through the pager; verify the table is at the top, keys are bold and the rest of the document renders as before (after: 11464450 bytes, +450288 bytes, +4.1%)
