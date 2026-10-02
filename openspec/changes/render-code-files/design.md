# Design

## Context

`cmd/termd` reads the input (a file or stdin) and passes it to `render.Render`, which always parses it as markdown. Code blocks are laid out by `renderer.code` in `internal/render/render.go`: a diagram info string goes to the diagram renderer, other code is shown verbatim, and in styled mode it is highlighted by `internal/highlight` when the info string names a chroma lexer (`lexers.Get`, never by content). The resulting `Block` is marked `Wide` when a line is wider than the output; only the containers (lists, quotes) look at that flag. The pager decision in `cmd/termd` measures the widest line of the final output, so a wide line opens `less -RS` wherever it comes from.

chroma v2.27.0 also matches lexers by file name: `lexers.Match(name)` takes the base name and tries the `Filenames` globs of every lexer (`*.go`, `Makefile`, `Dockerfile`), then the `AliasFilenames` globs, and picks the lexer with the highest priority. A name with an editor or packaging suffix (`~`, `.bak`, `.orig`, `.in` and a few others) matches like the name without it. The globs are matched case-sensitively.

The requirements are in `specs/cli/spec.md`.

## Goals / Non-Goals

**Goals:**
- A code file goes through the same verbatim and highlighting code as a fenced code block, so both look the same.
- Markdown files and stdin render byte for byte as before; the existing golden files do not change.

**Non-Goals:**
- Line numbers, a frame or a file name header around the code. The output is the file, as `cat` would show it, plus colors.
- Code files on the landing page (`internal/site`); it renders markdown examples only.
- Changing the `--help` text. The proposal names the README sections to update and leaves the usage text out.

## Decisions

### Lookup by file name in `internal/highlight`

A new `fileLexer(name)` sits next to `lexer(info)`: it calls `lexers.Match(name)` and returns nil when nothing matches or when the matched lexer is named `markdown` or `plaintext`. These are the two names in chroma v2.27.0: `markdown` matches `*.md`, `*.mkd` and `*.markdown`, `plaintext` matches `*.txt`. Two exported functions mirror the info string API: `RecognizedFile(name)` next to `Recognized(info)`, and `HighlightFile(code, name, theme)` next to `Highlight(code, info, theme)`. The tokenizing body of `Highlight` moves into an unexported `highlightWith(lexer, code, theme)` shared by both; it is not changed.

The rule is the one of the proposal, "recognized by the highlighter", taken literally. A spike over common names gave:

| Name | Lexer | Rendered as |
|---|---|---|
| `main.go`, `dir/main.go`, `main.go.bak` | Go | code |
| `config.yaml`, `config.yml` | YAML | code |
| `Makefile`, `makefile`, `GNUmakefile`, `Makefile.in` | Makefile | code |
| `Dockerfile`, `Containerfile` | Docker | code |
| `CMakeLists.txt` | CMake | code |
| `go.mod` | AMPL | code |
| `.env`, `.bashrc` | Bash | code |
| `x.rst`, `x.org` | reStructuredText, Org Mode | code |
| `README.md`, `x.markdown`, `README.md.orig` | markdown | markdown |
| `notes.txt`, `requirements.txt` | plaintext | markdown |
| `README`, `LICENSE`, `CHANGELOG`, `.gitignore`, `go.sum` | none | markdown |
| `x.mdx`, `x.mdown`, `x.qmd`, `x.Rmd` | none | markdown |
| `MAIN.GO`, `README.MD`, `NOTES.TXT`, `dockerfile` | none | markdown |
| `x.mmd`, `x.mermaid`, `x.puml`, `x.plantuml` | none | markdown |

Consequences that the proposal did not spell out:

- `CMakeLists.txt` is code. chroma's CMake lexer matches the whole name and outranks the `*.txt` pattern of `plaintext` (priority -1), so the highlighter recognizes it as CMake. "`.txt` files" in the proposal is read as "files the highlighter recognizes as plain text".
- Other markup languages (`.rst`, `.org`, `.gmi`) are code, because they are neither markdown nor plain text.
- Matching is case-sensitive, as chroma's globs are: `MAIN.GO` and `dockerfile` are rendered as markdown.
- `go.mod` matches the AMPL lexer (chroma maps `*.mod` to AMPL and Modula-2). The file is still shown line by line, only the colors follow AMPL.

Alternatives:
- Retry the lookup with a lower-cased name. Rejected: it adds a rule of termd's own on top of "recognized by the highlighter", and lower-casing alone breaks `Dockerfile`, whose pattern is case-sensitive.
- A list of markdown extensions in termd, everything else is code (glow's rule). Rejected by the proposal for `.txt` and unrecognized names.
- `lexers.Get` on the base name. Rejected: `Get` tries lexer names and aliases before the patterns, so files named `go` or `python` would be code, matched by a language name rather than by a file name pattern.
- Passing the matched lexer's name as an info string to `Highlight`. Rejected: 29 lexer names contain spaces (`Protocol Buffer`, `Go Template`), `Language` keeps only the first word, and the info string path goes through the diagram check.

### `render.Code` as a separate entry point

`render.Code(src, name, opts)` returns the terminal output of a whole file as one code block. It never calls `Parse`, the frontmatter split or `diagram.Language`, so fence lines, `---` lines and `#` lines in the file have no meaning, and the file can never reach the diagram renderer. It produces the lines the way `renderer.code` does: `verbatim` for plain text, or `highlight.HighlightFile` when styling is enabled and the name is recognized. The theme is resolved through `renderer.resolveTheme`, so it is asked for once and only when the file is highlighted.

`Code` does not compute `Block.Wide`: the file is the only block, no container lays it out, and the pager already measures the widest line of the output. Lines are never wrapped, so a wide line is output in full and scrolls in `less -RS`.

The loop that writes lines with their escape sequences and a line break moves from `Render` into a small `output(lines, style)` helper used by both entry points; the output of `Render` does not change.

Alternatives:
- Wrapping the file in a markdown fence and calling `Render`, as glow does. Rejected by the proposal: a file that contains a fence line closes the block early.
- A field in `Options` that makes `Render` skip the parser. Rejected: `Render` is the markdown entry point, and every markdown step would have to check the field.

### The decision in `cmd/termd`

`run` remembers the path it read the document from; it stays empty for stdin. After the file was read, a path recognized by `highlight.RecognizedFile` goes to `render.Code`, anything else to `render.Render`. The decision uses the path as given on the command line, before anything is read from the content. Read errors happen before the decision, so the exit statuses stay as they are, and the pager decision after it is unchanged.

Alternative: the decision inside `render`, for example `render.File(src, name, opts)`. Rejected: the proposal places the choice in `cmd/termd`, and `render` would need to know that stdin has no name.

### Line breaks and empty files

`Code` cuts one trailing line break from the file, as `renderer.code` does for a code block, and `output` ends every line with a line break. A file that ends with a line break is therefore output as it is, a file without one gets a line break after its last line, and a trailing empty line in the file (two line breaks at the end) is kept. An empty file returns an empty string, the same as an empty markdown document; without that rule it would be one empty line.

### Diagram-like extensions

chroma v2.27.0 has no lexer for mermaid or PlantUML, so `.mmd`, `.mermaid`, `.puml`, `.plantuml`, `.pu` and `.iuml` files are not recognized and are rendered as markdown, as today. If a later chroma version adds such a lexer, the file is shown as highlighted source and not as a diagram, because `Code` does not consult `diagram.Language`. Drawing a bare diagram file is a separate feature.

### Tests and fixtures

`TestGolden` only reads `testdata/*.md`. Code file fixtures go into `testdata/codefiles/` (`main.go`, `config.yaml`, `Makefile`), outside the glob and outside the Go packages (`testdata` is ignored by `go build`, `go vet ./...` and the `gofmt` target, which covers `cmd` and `internal` only). A new `TestCodeFileGolden` in `internal/render` checks that the plain output of every fixture is the file itself and compares the styled output with `testdata/golden/<file>.styled.golden`. A plain golden file would only repeat the fixture; the width is fixed at 80, since the output of a code file does not depend on it.

## Risks / Trade-offs

- [A file is highlighted by a wrong lexer because chroma's patterns overlap, for example `go.mod` as AMPL] → The text is unchanged and every line is kept; only the colors are off. Better patterns belong upstream.
- [A document with a code-like name, for example `notes.org`, is shown as code instead of markdown] → Intended by the proposal: only markdown and plain text names are markdown. `cat notes.org | termd` renders it as markdown.
- [A file with CRLF line endings: in plain mode the carriage returns are output as they are; in styled mode chroma converts the line endings to LF before tokenizing] → Both look the same in a terminal; the plain output stays the text of the file.
- [A file with lone CR line endings loses lines in styled mode: chroma turns each CR into a line break and `Highlight` keeps only as many lines as the code has LF line breaks] → The same happens today to a code block with such content. Control characters in document text belong to the separate escape-control-characters change; the code file path shares `verbatim` and `Highlight` with code blocks, so a fix there covers both.
- [Highlighting a very large file is slower than printing it] → The same tokenizer runs for code blocks today; no limit is added.

## Migration Plan

None. Markdown files and stdin render as before, so the existing golden files do not change. Rollback is a revert of the change.
