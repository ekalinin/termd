# color-themes Specification

## Purpose

Defines the built-in color themes of termd: which themes exist, which elements of a document a theme styles, how the styles of nested elements combine, and the colors of every theme.

## Requirements

### Requirement: Built-in themes
termd SHALL provide these themes: `dark`, `light`, `dracula`, `nord`, `onedark`, `monokai`, `solarized-dark`, `solarized-light`, `gruvbox`, `gruvbox-light`, `catppuccin-mocha` and `catppuccin-latte`. Every theme SHALL be one fixed set of styles that does not depend on the terminal background. A theme SHALL style the document elements listed in "Themed elements" and SHALL color the code blocks in a recognized language with a highlighting style: `github-dark` for `dark`, `github` for `light`, and the highlighting style of the same name for every other theme. When styling is disabled, the output SHALL be the same with every theme and SHALL contain no escape sequences.

#### Scenario: Theme in a pipe
- **WHEN** the user runs `termd --theme=dracula README.md > out.txt`
- **THEN** `out.txt` is byte for byte the output of `termd --theme=dark README.md > out.txt` and contains no escape sequences

#### Scenario: Code block in a named theme
- **WHEN** a fenced block with the info string `go` is rendered to a terminal with `--theme=nord`
- **THEN** the block is colored with the colors of the `nord` highlighting style, and removing the escape sequences from the output yields the original source

### Requirement: Themed elements
A theme SHALL define one style for each of these elements:
- headings: the `#` markers and the text, the same style for every level;
- inline code;
- links: the visible text of a link, an autolink and the `[image: alt]` label of an image;
- markers: list bullets and numbers, horizontal lines, including the line before the footnote definitions, and the quote marker of a regular block quote;
- each of the five GitHub alert types: the title and the quote marker of the alert;
- table header cells;
- table borders: the column separators and the line under the header, in markdown tables and in the frontmatter table;
- frontmatter keys.

The style of an element SHALL apply on top of the style of the text around it: its attributes (bold, italic, underline, strikethrough, faint) SHALL be added, and its color, when it has one, SHALL replace the color around it. The inner element SHALL win: inline code inside a heading or a link SHALL have the inline code color, and a link inside a heading SHALL have the link color and stay bold. In a table header cell, text without a color of its own SHALL get the header color, and inline code and links SHALL keep their colors.

A theme SHALL NOT color paragraph text, emphasis, strong and strikethrough text (they keep only their attributes), footnote references, task checkboxes, diagrams, framed diagram source and code blocks without a recognized language. No theme SHALL set a background color.

#### Scenario: Link inside a heading
- **WHEN** a document contains `# See [docs](https://example.com)` and is rendered to a terminal with `COLORTERM=truecolor` and `--theme=dracula`
- **THEN** `See` is bold in `#bd93f9`, and `docs` is bold, underlined and in `#8be9fd`

#### Scenario: Inline code inside a link
- **WHEN** a document contains ``[`run`](https://example.com)`` and is rendered to a terminal with `--theme=dark`
- **THEN** `run` is underlined and cyan (SGR `4;36`), the color of inline code

#### Scenario: Same style for every heading level
- **WHEN** a document contains `# Title` and `### Section` and is rendered to a terminal with `--theme=nord`
- **THEN** both headings, including their `#` markers, have the same style

#### Scenario: Table borders
- **WHEN** a table with a header row is rendered to a terminal with `COLORTERM=truecolor` and `--theme=nord`
- **THEN** every column separator `│` and the line under the header are in `#616e87`, and the header cells are bold in `#88c0d0`

#### Scenario: Unthemed elements
- **WHEN** a paragraph with `*italic*`, a footnote reference and a task item `- [ ] todo` is rendered to a terminal with `--theme=gruvbox`
- **THEN** the paragraph text, the footnote reference `[1]` and `[ ]` have no color, and `italic` is italic without a color

### Requirement: Default themes
The `dark` and `light` themes SHALL style the elements of a document outside code blocks in the same way, with the basic palette and the attributes of the terminal:
- headings and table header cells bold;
- inline code cyan (SGR `36`);
- links underlined and blue (SGR `4;34`);
- markers faint (SGR `2`);
- alerts in the basic palette colors given in markdown-rendering;
- table borders without a style;
- frontmatter keys bold.

`dark` and `light` SHALL differ only in the colors of code blocks.

#### Scenario: Document without code
- **WHEN** a document with headings, links, lists, alerts, a table and frontmatter but without code blocks is rendered to a terminal with `--theme=dark` and with `--theme=light`
- **THEN** both outputs are byte for byte the same

#### Scenario: Default element styles
- **WHEN** a document contains `# Title`, `` `x` ``, `[a](https://example.com)` and `- item` and is rendered to a terminal with `--theme=dark`
- **THEN** `# Title` is in SGR `1`, `x` in SGR `36`, `a` in SGR `4;34`, the bullet `•` in SGR `2`, and no 24-bit or 256-color sequence is emitted

### Requirement: Named theme colors
Every theme other than `dark` and `light` SHALL use 24-bit colors from its row of the table below:
- headings, table header cells and frontmatter keys SHALL be bold in the heading color;
- inline code SHALL be in the inline code color;
- links SHALL be underlined and in the link color;
- markers and table borders SHALL be in the marker color and SHALL NOT be faint;
- the quote marker of an alert SHALL be in the color of its type, and the title SHALL be bold in that color.

| Theme | Heading | Inline code | Link | Marker | Note | Tip | Important | Warning | Caution |
|---|---|---|---|---|---|---|---|---|---|
| `dracula` | `#bd93f9` | `#50fa7b` | `#8be9fd` | `#6272a4` | `#bd93f9` | `#50fa7b` | `#ff79c6` | `#f1fa8c` | `#ff5555` |
| `nord` | `#88c0d0` | `#8fbcbb` | `#81a1c1` | `#616e87` | `#81a1c1` | `#a3be8c` | `#b48ead` | `#ebcb8b` | `#bf616a` |
| `onedark` | `#e06c75` | `#98c379` | `#61afef` | `#7f848e` | `#61afef` | `#98c379` | `#c678dd` | `#e5c07b` | `#e06c75` |
| `monokai` | `#a6e22e` | `#fd971f` | `#66d9ef` | `#75715e` | `#66d9ef` | `#a6e22e` | `#ae81ff` | `#e6db74` | `#f92672` |
| `solarized-dark` | `#cb4b16` | `#2aa198` | `#268bd2` | `#586e75` | `#268bd2` | `#859900` | `#d33682` | `#b58900` | `#dc322f` |
| `solarized-light` | `#cb4b16` | `#2aa198` | `#268bd2` | `#93a1a1` | `#268bd2` | `#859900` | `#d33682` | `#b58900` | `#dc322f` |
| `gruvbox` | `#b8bb26` | `#8ec07c` | `#83a598` | `#928374` | `#83a598` | `#b8bb26` | `#d3869b` | `#fabd2f` | `#fb4934` |
| `gruvbox-light` | `#79740e` | `#427b58` | `#076678` | `#928374` | `#076678` | `#79740e` | `#8f3f71` | `#b57614` | `#9d0006` |
| `catppuccin-mocha` | `#fab387` | `#a6e3a1` | `#89b4fa` | `#6c7086` | `#89b4fa` | `#a6e3a1` | `#cba6f7` | `#f9e2af` | `#f38ba8` |
| `catppuccin-latte` | `#fe640b` | `#40a02b` | `#1e66f5` | `#9ca0b0` | `#1e66f5` | `#40a02b` | `#8839ef` | `#df8e1d` | `#d20f39` |

The colors SHALL be emitted as 24-bit colors or as the nearest colors of the 256-color palette, by the same rule as the colors of code blocks.

#### Scenario: Heading in truecolor
- **WHEN** a document contains `# Title` and is rendered to a terminal with `COLORTERM=truecolor` and `--theme=dracula`
- **THEN** the output line is `\033[1;38;2;189;147;249m# Title\033[0m`

#### Scenario: Heading without truecolor
- **WHEN** the same document is rendered to a terminal without `COLORTERM` and with `--theme=dracula`
- **THEN** the output line is `\033[1;38;5;141m# Title\033[0m`

#### Scenario: Marker is not faint
- **WHEN** a document contains `- item` and is rendered to a terminal with `COLORTERM=truecolor` and `--theme=catppuccin-mocha`
- **THEN** the bullet `•` is in `#6c7086` (SGR `38;2;108;112;134`) without SGR `2`

#### Scenario: Alert in a named theme
- **WHEN** a document contains `> [!WARNING]` and `> Careful.` and is rendered to a terminal with `COLORTERM=truecolor` and `--theme=nord`
- **THEN** the title `Warning` is bold in `#ebcb8b` (SGR `1;38;2;235;203;139`), and the quote marker of both lines is in `#ebcb8b` without bold
