# Tasks

## 1. Parsing and rendering

- [x] 1.1 Extend `TestParseGFM` in `internal/render/render_test.go` with `:rocket:`; the AST must contain an `ast.Emoji` of goldmark-emoji; add `TestEmojiShortcodes`, comparing the full plain output at width 80 for the scenarios of the spec: shortcodes in a paragraph (`Fast 🚀 and tested ✅`), in a heading (`# Features ✨`), in an alert (`│ Note`, `│ ⚠️ Read this.`), link text and destination (`docs 📖 (https://example.com/:book:)`), between inline HTML tags (`Press ⬆️ to scroll.`), after a letter or a digit (`a:b:c, foo:rocket:bar, 1:zap: and x 🅱️ y`), adjacent shortcodes (`🚀⚡`), names that are not emoji (unchanged), code and bare URLs (unchanged), image alt text (`[image: logo :star:] (logo.png)`), an HTML block (`:rocket: Fast`), frontmatter (`title │ :rocket: Launch`), and a definition list (`Term :one:`, `: Def :two:` gives `Term 1️⃣`, `    Def 2️⃣`); verify `go test ./internal/render -run 'TestParseGFM|TestEmojiShortcodes'` fails
- [x] 1.2 In one step, add `github.com/yuin/goldmark-emoji` v1.0.6 to `go.mod` without upgrading other modules, `internal/render/emoji.go` with the inline parser of design.md, the parser to `Parse` (and its doc comment), an `ast.Emoji` case to `renderer.inline` and the shortcode of an `ast.Emoji` to `renderer.plainText`; verify the tests of 1.1 pass, `go test ./internal/render ./internal/site` passes with unchanged golden files and `make tidy-check` passes

## 2. Styles and width

- [x] 2.1 Add `TestEmojiStyles` to `internal/render/render_test.go`: in styled mode `**:tada:**` renders as `**🎉**` and `# Features :sparkles:` as `# Features ✨`; the table with the header `Feature | Status` and the rows `Tables | :white_check_mark:` and `Diagrams | :warning:` has its column separator at the same column on every line, and the cells `✅` and `⚠️` are 2 columns wide by `text.Width`; verify it passes

## 3. Golden fixture

- [x] 3.1 Add `testdata/emoji.md` with shortcodes in a heading, a paragraph, a list, an alert, a table with `:white_check_mark:` and `:warning:`, link text, a footnote, a definition list and `<kbd>`, and with cases that stay text: `a:b:c`, `:octocat:`, inline code, a code block, a bare URL, image alt text and an HTML block; run `make golden`; verify `git status testdata/golden` shows only the six new `emoji.*.golden` files, and in them every shortcode of the first group is an emoji, every case of the second group is unchanged, and the table separators are aligned at every width

## 4. Documentation

- [x] 4.1 In `README.md` add an "Emoji shortcodes" item to "Features": GitHub shortcodes such as `:rocket:` become emoji, a shortcode right after a letter or a digit stays text, and shortcodes in code, link destinations, image alt text, HTML blocks and frontmatter are not replaced; verify every statement matches `specs/markdown-rendering/spec.md`

## 5. Verification

- [x] 5.1 Run `make check` and `openspec validate add-emoji-shortcodes --strict`; verify both pass
- [x] 5.2 Run `make build`, then `./termd --no-pager --width 60 testdata/emoji.md` in a terminal and with `| cat`; verify the emoji are shown in both, styled text keeps its style around them, and the plain output has no escape sequences
- [x] 5.3 Render the READMEs of `github.com/aws/smithy-go` and `github.com/fxamacker/cbor/v2` from the module cache; verify their shortcodes are shown as emoji and no text outside shortcodes changes compared with the previous release of termd
