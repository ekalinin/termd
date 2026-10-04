# Design

## Context

See proposal.md - Why. The README shows the theme screenshots as twelve PNG files of 1095x940 pixels in four markdown tables of three, about 1100 pixels high on GitHub. The heading `### Themes` under "How it works" owns the anchor `#themes`, which three places link to: the `--theme` row of the flags table, the sentence before the screenshots and the link in `internal/site/themes.md`. That link is also part of the showcase on the page and of every screenshot.

On the page, `page.html` renders the sticky bar (`Width`, `Theme`), then `<main>` with the five `<section class="example">` and the `<section class="showcase">`. The page theme switcher is styled by the ids `theme-auto`, `theme-light` and `theme-dark` in `style.css`. `TestShowcasePage` fails when the showcase does not come after the last example. No test checks the label of the page theme switcher.

## Goals / Non-Goals

**Goals:**
- Move the existing blocks only: no new content in the README or on the page other than the section heading and the switcher label.
- Keep every existing anchor and link working without new screenshots.

**Non-Goals:**
- A different layout of the screenshot grid.
- Changes to the ids, names or CSS of the switchers.

## Decisions

### README: a section between "Features" and "Install"

The screenshot block moves to its own section right after "Features": the feature list names the themes, and the next thing the reader sees are the themes themselves. The introduction and the feature list stay at the top.

Alternatives:
- The start of "Examples": the block moves up only about 40 lines and is still below "Install" and "Usage".
- Right under the introduction, before "Features": the 1100 pixels of screenshots push the feature list down, and the themes become the first picture of a tool whose description is about tables and diagrams.

### README: the section is called `Color themes`

GitHub gives the anchor `#themes` to the first heading with that text. A new `## Themes` above "How it works" would take `#themes`, and the documentation would move to `#themes-1`. Then the flags table and the sentence before the screenshots would point to the screenshots instead of the documentation, and the link baked into the screenshots would be fixed only by making them again. `Color themes` gets `#color-themes`, collides with nothing and is the name the "Features" list and the heading of `themes.md` already use.

### Page: the showcase is the first section of `<main>`

The `<section class="showcase">` block moves in `page.html` before the range over the examples, without changes. Its windows sit in the same `cols-<N>` wrappers, so the width switcher and the per-theme `<style>` rules work as before, and `.showcase` and `.example` have the same margins, so `style.css` does not change.

Alternative: after the first example, the table. The themes are then one example lower, not right under the switchers.

### Page: the page theme switcher is labelled `Page theme`

With the showcase first, the sticky `Theme` switcher (auto, light, dark) and the `termd theme` switcher (dark, light, dracula, ...) are next to each other, and `dark` and `light` mean different things in them. Only the visible label and the `aria-label` of the page theme radiogroup change. The ids `theme-*` and the radio name `theme` stay, so `style.css` keeps working.

Alternatives:
- Leave the label as is and rely on the bar border, the spacing and the `termd theme` label.
- An HTML heading above the showcase: the captions of the examples and of the showcase are part of the documents termd renders, and an HTML heading would break that.
- The showcase switcher under the window: the switcher would be about a screen below the top of the showcase.

### Tests

`TestShowcasePage` checks that the showcase comes before the first `<section class="example">` instead of after the last one. Its loop over the switcher inputs starts at the showcase section and does not change. The same test checks that the page has one radiogroup labelled `Page theme` and one labelled `termd theme`, which is the "Switcher label" scenario of the spec. `TestThemeScreenshots` checks only that the README references every screenshot and needs no change.

## Risks / Trade-offs

- [In the README, "Install" moves down by the height of the screenshot grid, about 1100 pixels on GitHub] → Accepted: the goal is to show the themes almost immediately, and "Install" is still the next section.
- [A later rename of the section to `Themes` would silently move the `#themes` anchor] → The proposal and this design record why the name differs; no test guards it.
