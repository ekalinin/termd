# Tasks

## 1. Page

- [x] 1.1 In `TestShowcasePage` in `internal/site/site_test.go`, require the showcase section to come before the first `<section class="example"` instead of after the last one, with the message `no showcase section before the examples`, and require exactly one `role="radiogroup" aria-label="Page theme"` and one `role="radiogroup" aria-label="termd theme"` on the page; verify `go test ./internal/site -run TestShowcasePage` fails on the current page
- [x] 1.2 In `internal/site/page.html`, move the `<section class="showcase">` block unchanged before the range over the examples, and change the visible label and the `aria-label` of the page theme radiogroup from `Theme` to `Page theme`, keeping the ids `theme-*` and the name `theme`; verify `go test ./internal/site` passes
- [x] 1.3 Run `make site` and open `_site/index.html` in a browser; verify the showcase is the first section under the sticky bar and shows `dark` at 80 columns, the bar reads `Width` and `Page theme`, selecting `dracula` and then 40 columns updates the showcase, the page theme switcher still changes the page and the code example, and both switchers work with JavaScript disabled

## 2. README

- [x] 2.1 In `README.md`, move the sentence "The same document in every [theme](#themes) at `--width 60`:" and the four screenshot tables unchanged from the end of "Examples" to a new `## Color themes` section between "Features" and "Install"; verify `go test ./internal/site -run TestThemeScreenshots` passes and "Examples" ends with the unsupported diagram block
- [x] 2.2 In the "Theme screenshots" subsection of "Development", change `[Examples](#examples)` to `[Color themes](#color-themes)`; verify `grep -n '#examples' README.md` finds nothing

## 3. Checks

- [x] 3.1 Run `make check`; verify it passes
- [x] 3.2 After the branch is pushed, open `README.md` of the branch on GitHub; verify "Color themes" with twelve screenshots follows "Features", the `Themes` link in the `--theme` row and the `theme` link above the screenshots open "How it works > Themes", and the `Color themes` link in "Theme screenshots" opens the new section
- [x] 3.3 After the `Pages` run for the merge commit passes, open `https://ekalinin.github.io/termd/`; verify the showcase is the first section under the sticky bar and the page theme switcher is labelled `Page theme`
