# Proposal

## Why

The demo of the color themes is far from the top. In the README, the theme screenshots are at the end of "Examples", after the text examples (line 103 of 301). On the landing page, the theme showcase comes after all five examples. A reader should see the themes almost immediately after opening the README or the page.

## What Changes

- README: the theme screenshots, together with the sentence that introduces them, move from the end of "Examples" to a new section `## Color themes` between "Features" and "Install". The section is not called `Themes`, so the anchor `#themes` keeps pointing to "How it works > Themes". The flags table, the sentence before the screenshots and the link in `internal/site/themes.md` use that anchor, and the link is also part of the showcase and of the screenshots. The "Theme screenshots" subsection of "Development" links to `#color-themes` instead of `#examples`.
- Landing page: the theme showcase moves before the examples and becomes the first section of the page content, right under the sticky switchers.
- Landing page: the page theme switcher in the sticky bar is labelled `Page theme` instead of `Theme`, in the visible label and in `aria-label`. The showcase switcher `termd theme` now sits right under it, and both offer `dark` and `light`, which do different things.
- The content does not change: the screenshots, the showcase document and the examples stay the same.

Out of scope:

- New screenshots in `docs/themes/`: the images do not change.
- An HTML heading above the showcase. As in the examples, the title of the showcase is part of the document termd renders: the `# Color themes` heading of `themes.md`.
- Changes to `style.css`: the showcase and the examples have the same margins.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `project-site`: the theme showcase comes before the examples, and the page theme switcher is labelled `Page theme`.

## Impact

- `README.md`: the new `## Color themes` section with the screenshot block, which is removed from "Examples"; the link in "Theme screenshots".
- `internal/site/page.html`: the showcase section before the examples; the label and `aria-label` of the page theme switcher.
- `internal/site/site_test.go`: `TestShowcasePage` checks that the showcase comes before the first example instead of after the last one, and checks the labels `Page theme` and `termd theme` of the two switchers.
- `internal/site/site.go`, `internal/site/style.css`, `cmd/termd-site`, the `Makefile` and `docs/themes/` do not change. `TestThemeScreenshots` still passes, because it checks only that the README references every screenshot, not where.
