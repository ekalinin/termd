# Design

## Context

`renderer.quote` in `internal/render/render.go` renders the block children of a block quote at the output width minus 2 and prefixes every line with `│ ` in `markerStyle` (faint). The quote is used wherever a block quote appears: at the top level, inside another quote and inside a list item.

goldmark has no alert support. A spike on the parser showed how the marker line arrives in the AST:

| Source | Inline nodes of the first paragraph |
|---|---|
| `> [!NOTE]`, `> Useful info.` | `Text "["`, `Text "!NOTE"`, `Text "]"` (soft line break), `Text "Useful"`, `Text " info."` |
| `> [!NOTE]` with two trailing spaces, or with a trailing `\` | the same, but `]` has a hard line break |
| `> \[!NOTE]` | `Text "\[!NOTE"`, `Text "]"` (raw source keeps the backslash) |
| `> [!NOTE] text` | `Text "["`, `Text "!NOTE]"`, `Text " text"` |
| `[!note]: https://x.y` defined, then `> [!NOTE]` | a `Link` node |

The failed link reference leaves the brackets as plain text, and the end of the marker line is the soft or hard line break flag of the last text node on that line. Trailing spaces are not part of the segments.

The requirements are in `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- Alerts are recognized on the AST that `Render` already builds; parsing does not change.
- Quotes that are not alerts render byte for byte as before, so the existing golden files do not change. None of the current fixtures has an alert marker.

**Non-Goals:**
- An alert on the landing page. The site gets the feature through `render.Render`, but its examples stay as they are.
- A new style API: the alert colors are expressed with the existing `style.Style` fields.

## Decisions

### Detection in `renderer.quote`

`quote` asks a helper `alert(n)` whether the quote is an alert. The helper takes the first child of the quote; it must be an `*ast.Paragraph`. It walks the inline children of the paragraph from the start, appending the raw source of each `*ast.Text` node, and stops after the first node with a soft or hard line break, or at the end of the paragraph. Any other node kind on the first line (a link, emphasis, a code span) means a regular quote. The collected text, trimmed of whitespace, is compared with the five markers with `strings.EqualFold`. On a match the helper returns the alert type and the rendered spans of the inline nodes after the marker line.

Matching the raw source, not the unescaped text, keeps `\[!NOTE]` a regular quote: escaping the bracket is the usual way to show the marker as text.

`strings.EqualFold` and not `strings.ToUpper`: `ToUpper` maps the dotless `ı` to `I`, so `[!tıp]` would become a tip; the simple case folding of `EqualFold` has no non-ASCII partners for the letters of the five markers.

Alternatives:
- A goldmark extension with an alert AST node. Rejected by the proposal; it would also need a new node kind and a renderer case for something the renderer can see on its own.
- Matching the raw first line of the paragraph (`Lines().At(0)`). Rejected: the inline nodes of the marker line would still have to be skipped to render the rest, and a backslash hard break keeps the `\` in the raw line, so `[!NOTE]\` would not match.

### The title takes the place of the marker line

The children of the quote are rendered as before. For an alert, the block of the first paragraph is replaced by the wrap of the title span, a `text.Break` and the spans of the rest of the paragraph (no `Break` when the marker is the whole paragraph). Everything else in the quote stays as it is.

So the layout follows the source: `> [!NOTE]`, `> Text.` gives the title and the text on consecutive lines, and a blank `>` line after the marker gives a blank quote line under the title, as between any two blocks of a quote. A quote with only the marker line is the title alone.

Alternatives:
- The title as a separate block before the rest of the quote. Rejected: blocks are joined with a blank line, so the common form `> [!NOTE]`, `> Text.` would get a blank line between the title and the text.
- Always dropping a blank line after the title. Rejected: it needs a special case in the join, and the source layout is a simple rule to explain.

### Colors from the basic palette

The alert types are a list of marker, title and basic ANSI color code:

| Marker | Title | Color | Code |
|---|---|---|---|
| `[!NOTE]` | `Note` | blue | 34 |
| `[!TIP]` | `Tip` | green | 32 |
| `[!IMPORTANT]` | `Important` | magenta (purple) | 35 |
| `[!WARNING]` | `Warning` | yellow | 33 |
| `[!CAUTION]` | `Caution` | red | 31 |

`internal/style` has two ways to express a color: `Style.ANSI`, a basic palette code, and `Style.FG`, a 24-bit color that is emitted as 24-bit or as the nearest 256-color entry depending on `Options.Depth`. termd uses `ANSI` for its own styles (`codeStyle` is 36, links are 34) and `FG` only for chroma themes, which come in a dark and a light variant chosen by the terminal background. The alert colors use `ANSI` like the other styles of the renderer: the terminal maps them through its own palette, so they fit dark and light backgrounds without a theme, do not depend on the color depth, and work in terminals with 16 colors.

Alternatives:
- GitHub's colors as 24-bit values, a pair per type for the dark and the light theme. Rejected: the renderer would need the theme for quotes, so the background query (OSC 11), which today runs only for documents with highlighted code, would also run for documents with alerts; and the colors would clash with the rest of the output, which uses the terminal palette.
- The bright codes 90-97. Rejected: they are harder to read on light backgrounds, and the renderer uses the normal codes elsewhere.

### Marker and title styles

The quote marker of an alert is `style.Style{ANSI: code}`: the color replaces the faint attribute of the regular marker, so the bar on the left shows the type of the alert, as the colored border does on GitHub. Faint would dim the color.

The title is `style.Style{Bold: true, ANSI: code}`. It plays the role of a heading, and termd shows headings in bold; bold also sets the title apart from the text when the palette color is close to the text color. Alternative: the color alone, rejected for this reason.

In plain mode `style.Options.SGR` emits nothing, so the title is the bare word after `│ `, the marker is the usual `│ ` and the output has no escape sequences. No icons, as the proposal says.

### Alerts at any depth

`quote` renders every block quote, so a quote nested in another quote or placed in a list item is an alert under the same rule, and a regular quote nested in an alert keeps its faint marker inside the alert's colored one. GitHub documents that alerts cannot be nested within other elements; termd does not copy this restriction, since the proposal defines an alert by the block quote alone and showing a nested alert as an alert reads better than its marker text.

## Risks / Trade-offs

- [A link reference definition with the label `!note` turns the marker into a link, and the quote stays regular] → Accepted: such a definition is unlikely in a real document.
- [Palette colors depend on the terminal theme, and yellow can be pale on a light background] → The same holds for code spans and links today; the bold title stays readable.
- [termd shows an alert where GitHub shows a nested quote with the marker as text] → Accepted, see "Alerts at any depth".

## Migration Plan

None. Quotes without a marker render as before, so the existing golden files do not change. Rollback is a revert of the change.
