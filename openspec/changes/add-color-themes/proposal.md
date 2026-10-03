# Proposal

## Why

termd's colors are fixed: outside code blocks it uses the basic palette of the terminal (bold headings, cyan inline code, blue links, faint markers, colored alerts), and `--theme` only switches code highlighting between github-dark and github. glow offers whole-document styles such as dracula or tokyo-night; termd has no way to give a document the look of a popular color scheme.

## What Changes

- A theme styles the whole document: headings, inline code, links, list markers, horizontal lines and quote markers, the five GitHub alert types, table headers, table borders (new, unstyled today), frontmatter keys, and code blocks through the chroma style of the same name.
- Built-in themes: `dark`, `light`, `dracula`, `nord`, `onedark`, `monokai`, `solarized-dark`, `solarized-light`, `gruvbox`, `gruvbox-light`, `catppuccin-mocha` and `catppuccin-latte`. Every theme is one fixed palette; there are no theme families with dark and light variants.
- `dark` and `light` produce exactly today's output: the basic palette of the terminal outside code, github-dark and github for code. They differ only in code highlighting, so `--theme=auto` keeps querying the terminal background only when a code block is highlighted.
- The other themes use 24-bit colors from their own palettes, written by hand for each theme, converted to the 256-color palette when the terminal does not advertise truecolor, as code colors are today.
- `--theme` accepts `auto` and every theme name. The `TERMD_THEME` environment variable selects the theme when the flag is not given; the flag wins over the variable, and `auto` is used when neither is set. An invalid value of the flag or of the variable is a usage error with exit status 2, and the message names the flag or the variable.
- The theme names are listed in the `--help` text of `--theme`, in the usage error and in the README. There is no `--list-themes` flag.

Out of scope:

- Theme files defined by the user and a configuration file; a separate change.
- A showcase of the themes on the landing page; the page keeps its dark and light renderings, also a separate change.
- Background colors, colors for emphasis, strong and strikethrough text, footnote references and task checkboxes, a style per heading level, and colors in diagrams.
- Theme families that pick a dark or light variant by the terminal background, and themes made from all chroma styles automatically.

## Capabilities

### New Capabilities
- `color-themes`: the built-in themes, the document elements a theme styles, how nested elements combine their styles, and the colors of the default and the named themes.

### Modified Capabilities
- `cli`: the highlighting theme selection becomes theme selection: the `--theme` values, the `TERMD_THEME` environment variable, its precedence and its usage error.
- `markdown-rendering`: GitHub alert colors come from the theme; the basic palette colors stay for the `dark` and `light` themes.

## Impact

- `internal/theme` (new): the theme definitions, the palette of every theme and the list of names.
- `internal/highlight`: a theme becomes the name of a chroma style; `Dark` and `Light` keep their meaning.
- `internal/render`: headings, inline code, links, markers, alerts, frontmatter keys and tables take their styles from the theme instead of package variables.
- `internal/text`: `LinkSpans` no longer picks the link color.
- `internal/table`: header and border styles become fields of `Table`.
- `cmd/termd`: theme resolution from the flag, `TERMD_THEME` and `auto`; usage text and errors.
- Tests: golden output of every named theme; the existing golden files must not change.
- `README.md`: the flags table, the theme section of "How it works".
- `internal/site`: no change in behavior.
