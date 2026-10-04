# project-site Specification

## Purpose

Defines the termd landing page on GitHub Pages: what the page contains, how its examples are produced from termd's real output, how the reader switches the example width and the theme, and how the page is built and published.

## Requirements

### Requirement: Page location and composition
The site SHALL be published at `https://ekalinin.github.io/termd/` as a single page. The page SHALL contain the project name, a one-line description, the install command, a link to the GitHub repository, the version of the latest release, the theme showcase, the examples and a footer with links to the repository and the license. The page SHALL NOT reproduce the documentation sections of the README.

#### Scenario: Opening the page
- **WHEN** a reader opens `https://ekalinin.github.io/termd/`
- **THEN** the page shows the project name, the one-line description, the install command, a link to `https://github.com/ekalinin/termd`, the version of the latest release, the theme showcase, the five examples and a footer linking to the repository and the license

#### Scenario: Install command matches the README
- **WHEN** the page is built
- **THEN** the install command on the page is the same `go install` command as in the Install section of the README

### Requirement: Example set
The page SHALL show five examples written for the site:
- a table with emoji, CJK text, column alignment from the markdown source, a long column that wraps while the short columns keep their width, and a link in a cell;
- a mermaid `sequenceDiagram`;
- a mermaid `flowchart`;
- a diagram type termd does not support, shown in the framed-source fallback;
- a fenced code block with syntax highlighting.

The example documents SHALL be separate from the test fixtures. A caption for an example SHALL be part of the example document, so it is rendered by termd with the rest of the example.

#### Scenario: Test fixtures do not affect the page
- **WHEN** a file under `testdata/` is changed and the site is rebuilt
- **THEN** the examples on the page do not change

#### Scenario: Unsupported diagram
- **WHEN** a reader looks at the unsupported diagram example
- **THEN** it shows the diagram source in a frame labelled with the language, the diagram type and the reason, as termd prints it

### Requirement: Examples are termd output
The text of each example on the page SHALL be exactly the output of the termd renderer for that example document at the shown width, with styles and hyperlinks enabled and 24-bit colors, after the escape sequences are removed. Bold, faint, italic, underline, strikethrough and colors in that output SHALL be shown with the same attributes on the page. A terminal hyperlink SHALL be shown as a clickable link with the same text and destination.

#### Scenario: Text matches the renderer
- **WHEN** the site is built
- **THEN** for every example, width and theme, the visible text of the example equals termd's styled output for the same document, width and theme with the escape sequences removed

#### Scenario: Link in a table cell
- **WHEN** a reader looks at the table example
- **THEN** the cell with the link shows only the link text, and clicking it opens the link destination

### Requirement: Alignment of wide characters
Every grapheme cluster that termd counts as two columns wide SHALL occupy exactly two character cells on the page, independent of the width of its glyph in the browser font.

#### Scenario: Table with emoji and CJK
- **WHEN** a reader looks at the table example in a browser whose font draws CJK or emoji glyphs narrower or wider than two cells
- **THEN** the column separators of every row line up vertically, as in the terminal

### Requirement: Unsupported escape sequences fail the build
The site build SHALL fail when termd output for an example contains an escape sequence the page does not know how to show, and the error SHALL name the example. Raw escape sequences SHALL never appear on the page.

#### Scenario: New escape sequence in termd output
- **WHEN** termd output for an example contains an escape sequence the page does not support
- **THEN** the site build exits with an error naming the example, and no page is published from that build

### Requirement: Width switcher
The page SHALL have one width switcher with 40, 60 and 80 columns that applies to all examples at once, with 80 selected when the page opens. Each example SHALL be shown as termd lays it out at the selected width. A block wider than the visible area SHALL scroll horizontally, and its lines SHALL NOT be wrapped by the browser.

#### Scenario: Default width
- **WHEN** a reader opens the page
- **THEN** 80 columns are selected and every example shows termd's 80-column layout

#### Scenario: Table at a narrower width
- **WHEN** the reader switches the width from 80 to 40 columns
- **THEN** every example, including the table, shows termd's 40-column layout

#### Scenario: Flowchart changes direction
- **WHEN** the reader switches the width from 80 to 40 columns
- **THEN** the flowchart that is laid out left-to-right at 80 columns is laid out top-to-bottom at 40 columns

#### Scenario: Narrow screen
- **WHEN** the page is opened on a screen narrower than 80 columns of the page font and 80 columns are selected
- **THEN** each example keeps its lines unwrapped and scrolls horizontally

### Requirement: Theme
The page SHALL follow the reader's system color scheme by default and SHALL offer a manual choice of auto, light and dark. The switcher of this choice SHALL be labelled `Page theme`. With a light page theme, the code example SHALL use termd's light highlighting theme; with a dark page theme, termd's dark highlighting theme. The manual choice SHALL NOT be saved across page loads.

#### Scenario: Switcher label
- **WHEN** a reader opens the page
- **THEN** the switcher with auto, light and dark is labelled `Page theme`, and the switcher of the theme showcase is labelled `termd theme`

#### Scenario: System dark scheme
- **WHEN** the reader's system prefers a dark color scheme and the theme choice is auto
- **THEN** the page uses the dark theme and the code example uses termd's dark highlighting colors

#### Scenario: Manual light theme
- **WHEN** the reader's system prefers a dark color scheme and the reader selects light
- **THEN** the page uses the light theme and the code example uses termd's light highlighting colors

#### Scenario: Reload
- **WHEN** the reader selects dark and reloads the page
- **THEN** the theme choice is auto again

### Requirement: No JavaScript
The page SHALL NOT include JavaScript. The width and theme switchers SHALL work with JavaScript disabled.

#### Scenario: JavaScript disabled
- **WHEN** a reader opens the page in a browser with JavaScript disabled
- **THEN** the examples are shown and both switchers work

### Requirement: Build and publication
On every push to `main`, the site SHALL be built from that commit and published to GitHub Pages. When the build fails, the previously published site SHALL stay in place. A maintainer SHALL be able to build the same page locally with one command, into a directory that git ignores.

#### Scenario: Change in rendering
- **WHEN** a commit that changes termd's table layout is pushed to `main`
- **THEN** after the deployment the table example on the page shows the new layout

#### Scenario: Failed build
- **WHEN** the site build fails for a push to `main`
- **THEN** the site keeps showing the page from the last successful deployment

#### Scenario: Local build
- **WHEN** a maintainer runs the local build command
- **THEN** the same page is written to a git-ignored directory and `git status` shows no new files

### Requirement: Latest release
The page SHALL show the version of the latest release under the install command, as a badge image that the reader's browser loads from an external badge service when the page is viewed. The version SHALL NOT be written into the page when the site is built. The badge SHALL link to `https://github.com/ekalinin/termd/releases/latest`. The badge and the link SHALL skip pre-releases. When the badge image cannot be loaded, the link SHALL show the text `Latest release`.

#### Scenario: Latest release on the page
- **WHEN** `v0.1.0` is the latest release and a reader opens the page
- **THEN** the page shows a badge with `v0.1.0` under the install command, and clicking the badge opens the release `v0.1.0`

#### Scenario: New release without a site build
- **WHEN** the release `v0.2.0` is published and no commit is pushed to `main` after it
- **THEN** once the cached badge expires, the page shows `v0.2.0` without a new site build

#### Scenario: Pre-release
- **WHEN** the release before the pre-release `v0.2.0-rc.1` is `v0.1.0`, and `v0.2.0-rc.1` is published
- **THEN** the badge shows `v0.1.0`, and clicking it opens the release `v0.1.0`

#### Scenario: Badge not available
- **WHEN** the badge image cannot be loaded
- **THEN** the page shows the text `Latest release` in place of the badge, and clicking it opens the latest release

#### Scenario: Build does not depend on releases
- **WHEN** the site is built from the same commit before and after a new release is published
- **THEN** both builds write the same page

### Requirement: Theme showcase
Before the examples, right under the width and page theme switchers, the page SHALL show a theme showcase: one document written for the site, `themes.md`, with a heading, a paragraph with a link and inline code, a list, alerts, a table and a code block in a recognized language, rendered by termd with each built-in theme. The showcase document SHALL be separate from the test fixtures and from the examples.

The showcase SHALL have its own theme switcher that lists every built-in theme in the order of `termd --help`, with `dark` selected when the page opens, and that works without JavaScript. The showcase SHALL follow the width switcher. The page theme switcher (auto, light, dark) SHALL NOT change the showcase, and the showcase switcher SHALL NOT change the page or the examples.

The window of the showcase SHALL have the background and the text color of the highlighting style of the selected theme. The basic palette colors in the `dark` and `light` windows SHALL be the page's colors for a dark and a light color scheme. The title of the window SHALL be the command that produces the output, for example `termd --width 80 --theme dracula themes.md`. The text and the styles of the showcase SHALL be termd's output as for the examples ("Examples are termd output").

#### Scenario: Opening the page
- **WHEN** a reader opens the page
- **THEN** the showcase is the first section under the switchers and shows `themes.md` with the `dark` theme at 80 columns, under the title `termd --width 80 --theme dark themes.md`, and the examples follow it

#### Scenario: Selecting a theme
- **WHEN** the reader selects `dracula` in the showcase switcher
- **THEN** the showcase shows termd's `dracula` output on the background `#282a36`, and the examples do not change

#### Scenario: Width
- **WHEN** `dracula` is selected and the reader switches the width to 40 columns
- **THEN** the showcase shows termd's 40-column `dracula` output

#### Scenario: Page theme
- **WHEN** `nord` is selected in the showcase and the reader selects the light page theme
- **THEN** the page turns light and the showcase still shows `nord` on its own background

#### Scenario: Showcase text matches the renderer
- **WHEN** the site is built
- **THEN** for every theme and width, the visible text of the showcase equals termd's styled output for `themes.md` at that width with that theme, with the escape sequences removed

#### Scenario: Without JavaScript
- **WHEN** a reader opens the page in a browser with JavaScript disabled and selects `gruvbox-light`
- **THEN** the showcase shows the `gruvbox-light` output

### Requirement: Theme screenshots
The README SHALL show a screenshot of the showcase document for every built-in theme at 60 columns, labelled with the name of the theme. The screenshots SHALL be PNG files `docs/themes/<name>.png` in the repository, made from the same HTML as the showcase window on the page. A maintainer SHALL be able to make all of them again with one command, `make screenshots`, which needs Google Chrome and ImageMagick on the machine. Building the page and the checks in CI SHALL NOT need Google Chrome or ImageMagick.

#### Scenario: A screenshot for every theme
- **WHEN** a reader opens the README on GitHub
- **THEN** it shows twelve screenshots, one for each built-in theme, each labelled with the theme name

#### Scenario: Missing screenshot
- **WHEN** a theme has no `docs/themes/<name>.png` or the README does not show it
- **THEN** `make check` fails and names the theme

#### Scenario: Making the screenshots
- **WHEN** a maintainer with Google Chrome and ImageMagick runs `make screenshots`
- **THEN** `docs/themes/<name>.png` is written for every built-in theme, each showing the window of the showcase at 60 columns with that theme

#### Scenario: Building the page without Chrome
- **WHEN** the page is built on a machine without Google Chrome and ImageMagick
- **THEN** the build succeeds
