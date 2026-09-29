# Tasks

## 1. ANSI to HTML converter

- [x] 1.1 Add a package under `internal/` that converts termd's styled output to HTML: SGR `0`, `1`, `2`, `3`, `4`, `9` to CSS classes, basic colors `30`-`37` and `90`-`97` to CSS variables, `38;2;R;G;B` to an inline `color`, OSC 8 open and close to `<a href>` and `</a>`, text and attribute values HTML-escaped; verify with unit tests for each sequence, for combined codes such as `1;3;38;2;R;G;B`, and for `<`, `>`, `&` and quotes in text and URLs
- [x] 1.2 Make the converter return an error for any escape sequence outside that set; verify with unit tests for `38;5;N`, a cursor movement sequence and an unterminated OSC 8
- [x] 1.3 Wrap every grapheme cluster that `text.Width` measures as 2 columns in a two-cell box, and leave clusters of width 1 as they are; verify with unit tests for a CJK character, a ZWJ emoji sequence, an emoji with a variation selector (`⚠️`) and plain ASCII
- [x] 1.4 Add golden tests for the converter using `internal/golden`, with ANSI inputs in the package's `testdata/`, and add the package to `GOLDEN_PKGS` in the Makefile; verify `make golden` creates the files and `make test` passes

## 2. Examples

- [x] 2.1 Write the five example documents for the site (table, `sequenceDiagram`, `flowchart`, unsupported diagram, highlighted code block), each with its caption inside the document, separate from `testdata/`; drafts are reviewed by the maintainer; verify each renders with `go run ./cmd/termd --width 80` and `--width 40`
- [x] 2.2 Tune the table example so the long column wraps at 40 columns while the short columns keep their width, and the link cell shows only its text; verify in the 40- and 80-column output
- [x] 2.3 Tune the flowchart example so it is left-to-right at 80 columns and top-to-bottom at 40; verify in the 80- and 40-column output

## 3. Page generation

- [x] 3.1 Render every example at 40, 60 and 80 columns with styles, hyperlinks and 24-bit colors, and the code example with both the dark and the light highlighting theme, then convert each result to an HTML fragment; a conversion error names the example; verify with a test that for every example, width and theme the fragment text without tags equals the renderer's output without escape sequences
- [x] 3.2 Add the page template with the project name, the one-line description (text chosen by the maintainer), the install command, a link to `https://github.com/ekalinin/termd`, the examples and a footer linking to the repository and the license; verify with a test that the page contains each element and that its install command equals the `go install` command in the README
- [x] 3.3 Add the width switcher: one radio group 40, 60, 80 above the examples, 80 checked, in a sticky row with the theme switcher, and CSS that shows the fragments of the checked width; verify with a test that the page has exactly one width group with 80 checked, and by switching widths in a browser
- [x] 3.4 Add the theme switcher: radio inputs auto (checked), light, dark; light and dark sets of CSS variables selected by `prefers-color-scheme` under auto and by `:has(...:checked)` otherwise; the dark or light code fragment follows the same rules; verify by emulating both color schemes in the browser dev tools and switching manually
- [x] 3.5 Add the CSS for terminal blocks: monospace font, `white-space: pre`, `line-height: 1.2` (1 lets box-drawing glyphs overlap the neighbouring lines), `overflow-x: auto`, two-cell boxes of width `2ch`, classes for bold, faint, italic, underline and strikethrough; verify visually in step 4.2
- [x] 3.6 Embed the template, the CSS and the example documents with `embed`, and add an entry point under `cmd/` that takes the output directory and writes `index.html` and the CSS; verify with a test that the generated page contains no `<script`, and that `go run` writes both files

## 4. Local build and review

- [x] 4.1 Add a Makefile target that builds the site into a directory listed in `.gitignore`; verify `git status` shows no new files after the build and `make check` passes
- [x] 4.2 Review the local page in at least two browsers: box-drawing lines are continuous, table columns with emoji and CJK line up, links open their destination, the switchers work with JavaScript disabled, a narrow window scrolls blocks horizontally; choose the font stack and the basic ANSI color values of each theme during this review

## 5. Deployment

- [x] 5.1 Add `.github/workflows/` workflow that runs on push to `main` with permissions `contents: read`, `pages: write`, `id-token: write` and a concurrency group: checkout, `setup-go` with `go-version-file: go.mod`, the generator, `upload-pages-artifact`, `deploy-pages`, using the current major versions of these actions; verify the workflow file parses (for example with `actionlint`)
- [x] 5.2 Maintainer sets Settings -> Pages -> Source to "GitHub Actions" before the merge; verify `gh api repos/ekalinin/termd/pages` returns `"build_type": "workflow"`
- [x] 5.3 After the merge, check the deployment run and open `https://ekalinin.github.io/termd/`; verify the page shows all five examples at 80 columns and the width and theme switchers work
- [x] 5.4 Maintainer fills the repository Homepage field with the site URL; verify `gh repo view --json homepageUrl` returns it
