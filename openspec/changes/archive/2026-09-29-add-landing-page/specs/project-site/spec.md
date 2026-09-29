# Spec Delta

## Purpose

Defines the termd landing page on GitHub Pages: what the page contains, how its examples are produced from termd's real output, how the reader switches the example width and the theme, and how the page is built and published.

## ADDED Requirements

### Requirement: Page location and composition
The site SHALL be published at `https://ekalinin.github.io/termd/` as a single page. The page SHALL contain the project name, a one-line description, the install command, a link to the GitHub repository, the examples and a footer with links to the repository and the license. The page SHALL NOT reproduce the documentation sections of the README.

#### Scenario: Opening the page
- **WHEN** a reader opens `https://ekalinin.github.io/termd/`
- **THEN** the page shows the project name, the one-line description, the install command, a link to `https://github.com/ekalinin/termd`, the five examples and a footer linking to the repository and the license

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
The page SHALL follow the reader's system color scheme by default and SHALL offer a manual choice of auto, light and dark. With a light page theme, the code example SHALL use termd's light highlighting theme; with a dark page theme, termd's dark highlighting theme. The manual choice SHALL NOT be saved across page loads.

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
