# Tasks

## 1. Control characters in the source

- [x] 1.1 Add `TestControlCharacters` to `internal/render/render_test.go`, with input built in Go (`"\x1b"`, `"\x07"`), one case per scenario of the spec that takes text from the source: the paragraph `hello \033[31mRED\033[0m and \033]0;pwned\007 title` gives exactly `hello ␛[31mRED␛[0m and ␛]0;pwned␇ title`; `**bold \033[31m**` styled gives `\033[1mbold ␛[31m\033[0m`; `a\000b\177c`, U+009B, `d`, 0xFF, `e` gives `a␀b␡c�d�e`; a heading; the table row `a\033b | x` reads `a␛b` with the separator at the same column on every line; a `go` block (styled, style sequences removed) and a plain block; the link `[docs](https://example.com/\033]0;pwned\007)` with hyperlinks and without styling gives exactly `\033]8;;https://example.com/␛]0;pwned␇\033\docs\033]8;;\033\`, and without hyperlinks `docs (https://example.com/␛]0;pwned␇)`; the image `![alt\033\007](x.png)`; inline HTML and an HTML block; a raw ESC in a frontmatter value; the invalid YAML frame of `title: [\033[31m`; a mermaid flowchart label (the right border of the box at the same column in the top line and in the label line) and the frame of an unsupported diagram; verify `go test ./internal/render` fails
- [x] 1.2 Add `TestPipeOutputWithControlCharacters` to `cmd/termd/main_test.go`: a piped document with `\033[31m` and `\033]0;pwned\007` in a paragraph, with default flags, gives output with no ESC and no BEL that contains `␛[31m` and `␛]0;pwned␇`; verify `go test ./cmd/termd` fails
- [x] 1.3 Create `internal/render/control.go` with `controlPicture` (the mapping table of the design) and `cleanSource`, applying it with `bytes.Map`, and call `cleanSource` at the start of `render.Render`, before `splitFrontmatter`; verify the tests of 1.1 and 1.2 pass, `go test ./...` passes and `git status testdata/golden internal` shows no changed golden file

## 2. Line endings

- [x] 2.1 Add `TestLineEndings` to `internal/render/render_test.go`: a document with CRLF line endings holding a paragraph of two lines and a fenced block with `code a` and `code b` renders exactly like the same document with LF endings and contains no `\r`; the paragraph `one\rtwo` renders as `one two` and a fenced block with `three\rfour` as the lines `three` and `four`; a document with a BOM and CRLF frontmatter still starts with the frontmatter table through `Render`; verify `go test ./internal/render` fails
- [x] 2.2 In `cleanSource` turn CRLF into LF and then every remaining CR into LF before the control characters are replaced; verify the tests of 2.1 pass and `TestSplitFrontmatter` still passes unchanged

## 3. Decoded text

- [x] 3.1 Add `TestDecodedControlCharacters` to `internal/render/render_test.go`: the paragraph `&#27;[31m red &#7; &#13; &#x9b;` gives `␛[31m red ␇ ␍ �`, `&#27;` in image alt text and in a table cell gives `␛`; frontmatter with `title: "\e]0;pwned\a"` gives the value `␛]0;pwned␇`, `tags: ["\e", b]` gives `␛, b`, a key `"\a": x` gives the key `␇`, and `nested: [{k: "\e"}]` gives `[{k: "␛"}]`; verify `go test ./internal/render` fails
- [x] 3.2 Map the result of `unescape` with `controlPicture`, and walk the frontmatter `yaml.Node` tree in `frontmatterBlock` to map every `Value` before the table is built; verify the tests of 3.1 pass and `go test ./...` passes

## 4. Golden fixture

- [x] 4.1 Add `testdata/control.md` with CRLF line endings: frontmatter with a YAML escape and a raw ESC, a heading, a paragraph with a color sequence, a title sequence, DEL, U+009B, the byte 0xFF and character references, a paragraph with a lone CR, a table, a link and an image with control characters, inline HTML and an HTML block, a `go` block, a plain block with a lone CR, a mermaid flowchart and an unsupported diagram; every line names in words the bytes it contains and the file contains no NUL; run `make golden`; verify `git status testdata/golden` shows only the six new `control.*.golden` files, the plain files contain no control character other than line feed, and removing the SGR and OSC 8 sequences from the styled files leaves no ESC

## 5. Documentation

- [x] 5.1 In `README.md` add a "Control characters" subsection to "How it works" (control pictures for C0 controls and DEL, `�` for C1 controls and invalid UTF-8, character references and YAML escapes included, CRLF and a lone CR as line endings); verify every statement matches `specs/markdown-rendering/spec.md`

## 6. Verification

- [x] 6.1 Run `make check` and `openspec validate escape-control-characters --strict`; verify both pass
- [x] 6.2 Run `make build`, then `./termd --no-pager testdata/control.md | od -c` and `./termd --no-pager --hyperlinks=always testdata/control.md | cat -v`; verify the output has no `\r`, no BEL and no ESC except in termd's own OSC 8 sequences, and that the control pictures appear where the fixture has control characters
