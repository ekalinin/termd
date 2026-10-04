# Spec Delta

## MODIFIED Requirements

### Requirement: Page location and composition
The site SHALL be published at `https://ekalinin.github.io/termd/` as a single page. The page SHALL contain the project name, a one-line description, the install command, a link to the GitHub repository, the version of the latest release, the examples, the theme showcase and a footer with links to the repository and the license. The page SHALL NOT reproduce the documentation sections of the README.

#### Scenario: Opening the page
- **WHEN** a reader opens `https://ekalinin.github.io/termd/`
- **THEN** the page shows the project name, the one-line description, the install command, a link to `https://github.com/ekalinin/termd`, the version of the latest release, the five examples, the theme showcase and a footer linking to the repository and the license

#### Scenario: Install command matches the README
- **WHEN** the page is built
- **THEN** the install command on the page is the same `go install` command as in the Install section of the README

## ADDED Requirements

### Requirement: Theme showcase
After the examples, the page SHALL show a theme showcase: one document written for the site, `themes.md`, with a heading, a paragraph with a link and inline code, a list, alerts, a table and a code block in a recognized language, rendered by termd with each built-in theme. The showcase document SHALL be separate from the test fixtures and from the examples.

The showcase SHALL have its own theme switcher that lists every built-in theme in the order of `termd --help`, with `dark` selected when the page opens, and that works without JavaScript. The showcase SHALL follow the width switcher. The page theme switcher (auto, light, dark) SHALL NOT change the showcase, and the showcase switcher SHALL NOT change the page or the examples.

The window of the showcase SHALL have the background and the text color of the highlighting style of the selected theme. The basic palette colors in the `dark` and `light` windows SHALL be the page's colors for a dark and a light color scheme. The title of the window SHALL be the command that produces the output, for example `termd --width 80 --theme dracula themes.md`. The text and the styles of the showcase SHALL be termd's output as for the examples ("Examples are termd output").

#### Scenario: Opening the page
- **WHEN** a reader opens the page
- **THEN** the showcase shows `themes.md` with the `dark` theme at 80 columns, under the title `termd --width 80 --theme dark themes.md`

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
