# Design

## Context

`render.Parse` builds the goldmark parser with `extension.GFM` and `extension.Footnote`. `renderer.block` lays out the known block nodes; a block it does not know but that has children goes through `renderer.blocks` again, and inline children of such a block are dropped, because `renderer.block` returns nothing for an inline node. List items are laid out by `renderer.items`, which prefixes the first line of an item with its marker and the further non-empty lines with spaces.

goldmark v1.8.6 ships `extension.DefinitionList` with the nodes `DefinitionList`, `DefinitionTerm` and `DefinitionDescription` in `extension/ast`. A spike with the cases of the spec showed how it builds the AST:

- A line that starts with `:` and at least one space, after a paragraph, opens a `DefinitionList`. The parser takes the paragraph out of the document and makes every line of it a `DefinitionTerm` with inline children, so `Apple`, `Pomme`, `: A fruit` gives two terms. The `:` line can interrupt a paragraph.
- Each `:` line opens a `DefinitionDescription`, a container of block children. Indented lines after it, two columns or more for `: `, continue it, so `    Second paragraph.` after a blank line becomes a second block of the same description. Lazy continuation lines join its first paragraph.
- `DefinitionDescription.IsTight` is true when the `:` line has no blank line before it. A tight description gets its paragraphs as `TextBlock` nodes, a loose one as `Paragraph` nodes; `renderer.block` already treats both alike.
- All entries of a list are children of one `DefinitionList`, in source order: term, term, description, description, term, description. A blank line between a description and the next term does not end the list and is not recorded anywhere in the AST.
- A `:` line after a heading or another block that is not a paragraph or a definition list stays text. `:not` without a space stays text.
- Definition lists work inside list items and block quotes like any other block.

The requirements are in `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- The renderer only lays out the definition list nodes; what is a term and where a definition ends comes from goldmark.
- Documents without definition lists render byte for byte as before. No current fixture, no example of the site and no README contains a line that starts with `: `, so the existing golden files do not change.

**Non-Goals:**
- Changing how goldmark turns the lines of a paragraph into terms.
- A definition list example on the landing page. The site gets the feature through `render.Render`, its examples stay as they are.

## Decisions

### Enable `extension.DefinitionList` in `Parse`

`Parse` becomes `goldmark.WithExtensions(extension.GFM, extension.Footnote, extension.DefinitionList)`. The parser and the rendering have to land in the same step of the implementation: with the extension and without a `DefinitionList` case, `renderer.block` would drop the terms, whose children are inline nodes, and show the descriptions as plain paragraphs.

Alternative: a parser of our own. Rejected: goldmark already parses the syntax, including descriptions with several blocks and nesting in containers.

### One block per definition list

A `DefinitionList` case in `renderer.block` returns one block and walks the children in order:

- A `DefinitionTerm` is wrapped like a paragraph, `text.Wrap(r.inlines(term, style.Style{Bold: true}), width, true)`, so wrapped lines of a long term start at column 0.
- A `DefinitionDescription` is laid out with `r.blocks(desc, width-4)`, its blocks joined with a blank line, and every non-empty line prefixed with 4 spaces, as `renderer.items` does with the further lines of an item. Empty lines stay empty, so the output has no trailing spaces.
- `Wide` is set when a block of a description is wide, for example a code block wider than the width minus 4.

Alternative: reusing `renderer.items`. Rejected: it treats every child as an item with a marker, while terms have no marker and no indentation; making it skip terms adds parameters to the list layout for one caller. The 4-space prefix is a loop of a few lines.

### Indentation by 4 columns, without a marker

A definition is indented by 4 columns and has no marker, as browsers show `<dd>` and as man pages show tagged paragraphs.

Alternatives:
- The `: ` of the source as a marker before each definition. Rejected: the output reads as unrendered source. Headings keep `#` because it shows the level; the colon carries no information.
- 2 columns. Rejected: in plain mode, without bold terms, a 2-column indent barely separates a definition from the next term.
- An arrow such as `🠶`, as glamour does. Rejected: terminals disagree on the width of such symbols, the same reason why alerts have no icons.

### Term style

The term is the base style `Bold` for its inline children. Inline code and links layer their palette styles on top of it, so they keep their colors and become bold, the same as in headings. A term takes no style from the palette, so no theme colors it and `theme.Palette` does not change.

Alternative: a palette field for terms. Rejected: the proposal leaves a theme color for terms out of scope.

### Tight or loose list

The list is loose when one of its descriptions is not tight (`IsTight` is false) or has more than one block, and tight otherwise. In a loose list a blank line comes before every description except the first one of a term, and before every term that follows a description. A term is never separated from its first description or from the next term. This is the rule of the footnote definitions and of CommonMark lists, applied to the information goldmark keeps.

Alternatives:
- A blank line between entries when the source has one. Rejected: goldmark does not record it, and finding it would mean scanning the source between nodes; HTML does not show it either.
- Always loose. Rejected: a blank line between short entries doubles the height of a glossary.
- Always tight. Rejected: the blocks of a definition with several paragraphs would run into the next term.

## Risks / Trade-offs

- [GitHub does not render definition lists. A document written for GitHub with a paragraph followed by a line that starts with `: ` renders differently in termd] → Accepted with the change, see proposal.md. A line of prose rarely starts with a colon and a space.
- [Every line of the paragraph before a definition becomes a term: a paragraph of two lines followed by `: Definition` gives two terms] → goldmark behavior, which follows PHP Markdown Extra, where several terms share their definitions. It is a scenario of the spec, so the behavior is explicit.
- [In a tight list, two definitions of one term look like one definition that wraps] → Accepted, as in HTML. A blank line before the second definition makes the list loose and separates them.
- [A blank line before a definition list inside a list item does not make the list loose: goldmark decides from the blank-line flag of the children of an item, and the `DefinitionList` that replaces the paragraph does not carry it, so `- item`, a blank line, `  Term`, `  : Def` renders without a blank line between `item` and `Term`] → Accepted: the columns are as the spec requires, and `renderer.items` takes the tightness of a list from goldmark, as for every other list.

## Migration Plan

None. Documents without definition lists render as before, so the existing golden files do not change. Rollback is a revert of the change.
