# Design

## Context

The styles of termd's elements are package variables spread over four packages:

| Element | Where | Today |
|---|---|---|
| inline code, markers | `internal/render/render.go:44` `codeStyle`, `markerStyle` | SGR 36, faint |
| heading | `internal/render/render.go` `block`, `hs` | bold |
| alerts | `internal/render/render.go:298` `alertTypes` | SGR 34, 32, 35, 33, 31 |
| frontmatter keys | `internal/render/frontmatter.go:105` `keyStyle` | bold |
| links | `internal/text/text.go:300` `linkStyle`, applied in `LinkSpans` to spans without a color | underline, SGR 34 |
| table header | `internal/table/table.go:40` `headerStyle`, `withStyle` adds bold only | bold |
| table borders | `internal/table/table.go` `row`, `Render` | unstyled |

`style.Style` already holds a basic palette color (`ANSI`) and a 24-bit color (`FG`), and `style.Options.SGR` converts `FG` to the 256-color palette when `COLORTERM` does not advertise truecolor. Code colors come from chroma: `highlight.Theme` is an `int` enum, `Dark` and `Light`, mapped to the `github-dark` and `github` styles. `render.Options.Theme` is a lazy `func() highlight.Theme`, called at most once and only when a code block is highlighted, so `--theme=auto` queries the terminal only then. `cmd/termd` maps the `--theme` value to that function in `themeFunc` and reads the environment through `env.getenv`, which the tests fake.

The landing page (`internal/site`) renders its examples with `render.Options` without anything theme related other than `Theme`, and its ANSI parser accepts SGR 30-37, 90-97 and `38;2`.

The requirements are in `specs/color-themes/spec.md`, `specs/cli/spec.md` and `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- One place that defines every theme, with the named palettes written as rows that read like the table in `specs/color-themes/spec.md`.
- Output with `dark` and `light` is byte for byte today's output: every existing golden file stays unchanged, in `internal/render`, `internal/highlight`, `internal/diagram` and `internal/site`.
- Callers that do not know about themes (the site, existing tests) keep working without changes.

**Non-Goals:**
- A theme file format or a public Go API for themes.
- Changing the background query (`internal/termbg`) or its timeout.
- Styling diagrams; `internal/diagram` is not touched.

## Decisions

### `internal/theme` holds the themes

A new package:

```go
// Palette is the style of every element a theme styles outside code blocks.
type Palette struct {
	Heading, InlineCode, Link, Marker      style.Style
	Note, Tip, Important, Warning, Caution style.Style
	TableHeader, TableBorder, FrontmatterKey style.Style
}

// Theme is a built-in theme: its name, its palette and the highlighting
// style of code blocks.
type Theme struct {
	Name    string
	Palette Palette
	Code    highlight.Theme
}
```

`Default` is the palette of `dark` and `light` (today's styles). The named themes are built by one helper from a row of nine hex colors in the column order of the spec table (heading, inline code, link, marker, note, tip, important, warning, caution). The helper applies the rules of "Named theme colors": bold heading color for headings, table headers and frontmatter keys; underline for links; the marker color for markers and table borders, without faint. A bad hex string panics at package initialization, and a test loads every theme, so a typo fails `go test`, not the user.

The package exports the list of themes in the order of the spec, a lookup by name, and the list of names for the help text and the usage error. `auto` is not a theme; `cmd/termd` handles it.

Alternatives:
- Themes in `internal/render`. Rejected: `cmd/termd` needs the names and the lookup, and render would grow a data table that has nothing to do with layout.
- A map of styles keyed by element name. Rejected: a struct makes a missing field a compile-time zero that a test can catch, and keeps the call sites typed.

### `highlight.Theme` becomes the chroma style name

`type Theme string`, with `Dark = "github-dark"` and `Light = "github"`; `styleName` returns the string. The names `Dark` and `Light` keep their meaning, so `internal/site`, the highlight golden tests and their file names stay as they are. A theme's `Code` is the chroma style of the same name. chroma's `styles.Get` silently falls back to its default style for an unknown name, so a test checks that every theme's `Code` is in `styles.Names()`.

Alternative: keep the enum and add a constant per theme. Rejected: twelve constants that only map to strings, and a second list to keep in sync with `internal/theme`.

### Styles layer from the outside in, the inner element wins

A new function in `internal/style` layers a style over a base: the attributes of both are combined, and the color of the top style replaces the color of the base when the top style has one (`ANSI` or `FG`; setting one clears the other). This one rule gives all the nesting behavior of "Themed elements":

- heading text renders on top of `Palette.Heading`;
- inline code: `Layer(base, Palette.InlineCode)`, replacing `cs.ANSI = codeStyle.ANSI`;
- links: render lays out the link children on top of `Layer(base, Palette.Link)`, so text inside a link gets the link color and underline, and inline code inside it gets the inline code color again; the same base is used for the label of an autolink and of an image;
- table header cells: each span becomes `Layer(TableHeader, span)`, so text without a color gets the header color and inline code and links keep theirs; `withStyle`, which only added bold, does this with `style.Layer`.

Today `LinkSpans` colors only spans that have no color yet. That rule cannot tell "inline code inside a link" from "link inside a colored heading": both arrive with a color. With the link color applied by render before the children are rendered, `LinkSpans` no longer touches styles; `linkStyle` is removed, and `LinkSpans` keeps its signature and only attaches the destination or appends ` (url)`.

With the default palette the result is today's output: no element around a link has a color today, so a link gets SGR 34 and underline as before, inline code inside a link keeps SGR 36, and a header cell gets bold as before.

Alternative: pass the link style into `LinkSpans` and keep the "only spans without a color" rule. Rejected: a link inside a colored heading would keep the heading color, against the spec.

### `render.Options.Palette`, the zero value is the default

`render.Options` gets a `Palette theme.Palette` field. A zero `Palette` means `theme.Default`, so `internal/site`, the golden tests (`goldenOptions`) and every other existing caller render as today without a change. `Options.Theme` stays the lazy highlighting theme.

The package variables `codeStyle`, `markerStyle` and `keyStyle`, and the color field of `alertTypes`, are replaced by lookups in the palette. `alertTypes` keeps its marker and title and gets the palette field of its type instead of a basic color code; the title is the alert style with bold added. `frontmatterBlock` receives the palette for the key style and the table border.

Alternative: a `*theme.Palette` where nil means default. Rejected: a value with a meaningful zero is the more common Go idiom here, and no theme needs an all-empty palette (plain mode already turns all styles off).

### Table styles are fields of `table.Table`

`Table` gets `HeaderStyle` and `BorderStyle`. `Render` styles the line under the header with `BorderStyle`, `row` styles each ` │ ` separator with it, and header cells are layered on `HeaderStyle`. `headerStyle` is removed; render sets both fields from the palette for markdown tables and the frontmatter table. A zero `Table` renders without styles; no existing `internal/table` test checks styles, so none of them changes.

### Theme resolution in `cmd/termd`

`run` already records which flags were given with `fs.Visit` for `--width`; the same visit records `--theme`. The value is taken in this order: `--theme` when given, `TERMD_THEME` when not empty, `auto`. Only the value that is used is checked, so an invalid `TERMD_THEME` next to a valid `--theme` is ignored. `--version` and `--help` are handled before the check, as today. The usage error reads `<source> must be one of auto, dark, light, ..., got "<value>"`, where `<source>` is `--theme` or `TERMD_THEME`, and the list comes from `internal/theme`, as does the `--help` text of the flag. It becomes `color theme: auto, dark, light, dracula, ...`.

`themeFunc` keeps its job of returning the lazy highlighting theme: for `auto`, the query as today, choosing between the `Code` of `dark` and `light`; for a theme name, a function that returns the theme's `Code` without a query. The palette passed in `render.Options.Palette` is the theme's palette, and for `auto` it is `theme.Default`. Since `dark` and `light` share `Default`, the palette does not depend on the query, and the rule "query only when a code block is highlighted" holds without a change in `render`.

The local variable `theme` in `run` collides with the package name and is renamed.

Alternatives:
- Resolve `auto` eagerly to a full theme. Rejected: it would query the terminal for every styled document, breaking "Nothing to highlight".
- Accept an invalid `TERMD_THEME` with a warning. Rejected by the proposal: it is a usage error, like the flag.

### The palettes

The colors in the spec table come from the published palettes of the themes and follow the chroma style of the same name where chroma defines the role:

- the marker color is the comment color of the chroma style, for example `#7f848e` for onedark and `#616e87` for nord;
- the heading color is the `GenericHeading` color of the chroma style where it has a color of its own (nord, solarized-dark, gruvbox, gruvbox-light and both catppuccin styles); dracula's `GenericHeading` is the plain text color, and onedark and monokai define none, so their heading colors follow the schemes' markdown conventions: purple in dracula, red in onedark, green in monokai;
- the link color is the blue or cyan of the scheme, and the inline code color one of its green, aqua or orange accents;
- the alert colors map the five GitHub types to the blue, green, magenta or purple, yellow and red of the scheme.

`solarized-light` shares the accents of `solarized-dark` and differs in the marker color, as the scheme does.

### Tests and golden files

- `internal/style`: the layering rule, including `ANSI` replacing `FG` and the other way round.
- `internal/theme`: the names in the order of the spec; every `Code` is a chroma style; `dark` and `light` have `Default` as palette and differ only in `Code`; every named palette has a color in each colored field, a bold heading, an underlined link and no faint marker.
- `internal/table`: border style on separators and the header line, header layering with a code span, zero styles render as today.
- `internal/render`: one new fixture `testdata/themes/sample.md`, outside the `testdata/*.md` glob of `TestGolden`, with a heading holding a link and inline code, a paragraph with emphasis, a footnote and a link, a list with a task item, a regular quote, the five alerts, a table, frontmatter, a `go` block and a horizontal line. `TestThemeGolden` renders it at width 60, styled, truecolor, with every theme into `testdata/golden/theme.<name>.golden`. The existing goldens are rendered with a zero `Palette` and must not change.
- `cmd/termd`: the fake environment gets a `TERMD_THEME` value; one test case per scenario of `specs/cli/spec.md` that the fake can check.

## Risks / Trade-offs

- [A named theme meant for a dark background is used on a light terminal, or the other way round, and markers or code become hard to read] → termd sets no background by design; the README says which themes are light (`solarized-light`, `gruvbox-light`, `catppuccin-latte`) and that a named theme does not adapt.
- [On a 256-color terminal, close palette colors map to the same index, for example two alert colors] → Same limitation as code colors today; truecolor terminals get the exact colors.
- [A chroma upgrade renames or removes a style] → The `internal/theme` test fails on the upgrade, before a release.
- [The `--help` line of `--theme` is long] → `flag` wraps nothing, so it is one long line; it stays readable and is generated from the same list as the error, so it cannot drift.
- [`TERMD_THEME` with an invalid value breaks every run, also in scripts that pipe termd] → Chosen in the proposal; the message names the variable so the cause is clear.

## Migration Plan

None. Default output does not change, `--theme=dark|light|auto` keep their behavior, and the new values and `TERMD_THEME` are additions. Rollback is a revert of the change.
