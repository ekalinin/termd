# Tasks

## 1. Lookup by file name

- [x] 1.1 Add `TestRecognizedFile` to `internal/highlight/highlight_test.go`: `main.go`, `dir/main.go`, `config.yaml`, `Makefile`, `Dockerfile` and `CMakeLists.txt` are recognized; `README.md`, `doc.markdown`, `notes.txt`, `README`, `MAIN.GO`, `flow.mmd`, `diagram.puml` and the empty name are not, and `HighlightFile` returns nil for them; `HighlightFile` of the `go` sample named `main.go` gives the same lines as `Highlight` with the info string `go`; verify `go test ./internal/highlight` fails
- [x] 1.2 In `internal/highlight/highlight.go` add `fileLexer` (`lexers.Match`, without the `markdown` and `plaintext` lexers), `RecognizedFile` and `HighlightFile`, and move the body of `Highlight` into `highlightWith`; verify `go test ./internal/highlight` passes and `git status internal/highlight/testdata` shows no changes

## 2. Whole file as one code block

- [x] 2.1 Add `TestCode` and `TestCodeThemeIsResolvedOnlyWhenNeeded` to `internal/render/render_test.go`, one case per rule of the spec: in plain mode `a\nb\n` named `x.go` gives `a\nb\n`, `a\nb` gives `a\nb\n`, `a\n\n` gives `a\n\n`, an empty file gives an empty string, `---\n# Server settings\nport: 8080\n` named `config.yaml` gives itself, a Go block comment with a line of three backticks and a line `# Title` gives itself, and a 120-column line at width 80 is output in full; in styled mode `func main() {}` named `main.go` has escape sequences and gives the source after stripping them; the theme is resolved once for the styled `main.go` and not at all in plain mode or for an empty file; verify `go test ./internal/render` fails
- [x] 2.2 Add `Code` to `internal/render/render.go` and move the loop that writes the lines of `Render` into an `output` helper shared by both; verify the tests of 2.1 pass and `go test ./internal/render ./internal/site` passes with unchanged golden files

## 3. Golden fixtures

- [x] 3.1 Add `TestCodeFileGolden` to `internal/render/golden_test.go` and the fixtures `testdata/codefiles/main.go` (a doc comment, tab indentation, a string, a block comment with a line of three backticks and a line `# Title`, a line wider than 80 columns), `testdata/codefiles/config.yaml` (`---`, `# comment` lines, a list and a mapping) and `testdata/codefiles/Makefile` (comments and tab-indented recipes); run `make golden`; verify `git status testdata` shows only the fixtures and the three new `<file>.styled.golden` files, the test checks that the plain output of each fixture is the fixture itself, and `cat` of the styled files shows colors

## 4. Command line

- [x] 4.1 Add `TestCodeFiles`, `TestStdinIsMarkdown` and `TestCodeFileInTerminal` to `cmd/termd/main_test.go`, one case per scenario of `specs/cli/spec.md` that the fake environment can check: `main.go` and `Makefile` in a pipe give the text of the file with exit status 0, `config.yaml` with `---`, `# Server settings` and `port: 8080` gives these three lines, `main.go` without a final line break gets one, an empty `empty.go` gives no output and exit status 0, `doc.md`, `notes.txt` and `README` with the lines `first line` and `second line` give the paragraph `first line second line`, and the YAML lines piped to stdin or read with `-` give a heading; in a 80-column terminal `main.go` is highlighted with one theme query, and a 120-column line goes to the pager; verify `go test ./cmd/termd` fails
- [x] 4.2 In `cmd/termd/main.go` keep the path of the input file, empty for stdin, and render the document with `render.Code` when `highlight.RecognizedFile` accepts the path, with `render.Render` otherwise; verify `go test ./cmd/termd` passes

## 5. Documentation

- [x] 5.1 In `README.md` add `termd main.go` to the Usage examples, a "Code files" subsection to "How it works" (a name recognized as a language other than markdown or plain text is shown as one code block; markdown, `.txt`, unrecognized names and stdin are rendered as markdown; the decision is made by the name only), and the code file golden files to the "Golden files" list; verify every statement matches `specs/cli/spec.md` and `design.md`

## 6. Verification

- [x] 6.1 Run `make check` and `openspec validate render-code-files --strict`; verify both pass and `git status testdata/golden` shows no modified files
- [x] 6.2 Run `make build` and `./termd --no-pager` on `cmd/termd/main.go`, `Makefile`, `.github/workflows/ci.yml`, `openspec/config.yaml` and `README.md`, in a pipe and in a pseudo-terminal (`script -q /dev/null ./termd --no-pager --theme=dark ...`); verify the code files are output line by line (plain output identical to the file, `cmp`), are highlighted in the terminal, the YAML `#` comments are not headings, and `README.md` renders as markdown as before

## 7. Control characters

- [x] 7.1 After the rebase on escape-control-characters, add the cases `control characters` (ESC and BEL become `␛` and `␇`) and `crlf and lone cr` (both become LF) to `TestCode` and a styled case with ESC and BEL, then call `cleanSource` in `render.Code`; verify `go test ./internal/render` fails before the call and passes after it, and `./termd --no-pager` on a `.go` file with ESC, BEL and CRLF outputs no raw control characters
