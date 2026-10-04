# Spec Delta

## MODIFIED Requirements

### Requirement: Page location and composition
The site SHALL be published at `https://ekalinin.github.io/termd/` as a single page. The page SHALL contain the project name, a one-line description, the install command, a link to the GitHub repository, the version of the latest release, the theme showcase, the examples and a footer with links to the repository and the license. The page SHALL NOT reproduce the documentation sections of the README.

#### Scenario: Opening the page
- **WHEN** a reader opens `https://ekalinin.github.io/termd/`
- **THEN** the page shows the project name, the one-line description, the install command, a link to `https://github.com/ekalinin/termd`, the version of the latest release, the theme showcase, the five examples and a footer linking to the repository and the license

#### Scenario: Install command matches the README
- **WHEN** the page is built
- **THEN** the install command on the page is the same `go install` command as in the Install section of the README

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
