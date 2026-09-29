# Proposal

## Why

termd has no project page: GitHub Pages is not enabled and the repository's Homepage field is empty. A landing page on github.io should show what termd does with its real output. The page cannot just paste that output into `<pre>`: a browser does not guarantee that a CJK character takes exactly two columns or that an emoji has a fixed width, so the page would show the same misaligned tables that termd fixes.

## What Changes

- New GitHub Pages site at `ekalinin.github.io/termd`: one page with the project name, a one-line description, the install command, a link to the GitHub repository, the examples and a footer. Documentation stays in the README; the page does not copy it.
- Five examples written for the site (independent of `testdata/`):
  - a table with emoji, CJK text, column alignment, wrapping of only the long column and a link in a cell;
  - a mermaid `sequenceDiagram`;
  - a mermaid `flowchart`;
  - an unsupported diagram shown in the framed-source fallback;
  - a syntax-highlighted code block.
- The examples are rendered by termd's own renderer at build time and converted from ANSI to HTML. Wide grapheme clusters are pinned to two columns, so alignment does not depend on the browser font.
- Width switcher with 40, 60 and 80 columns, CSS only.
- Theme follows `prefers-color-scheme`, with a manual auto/light/dark switch. The code example is rendered with both termd highlighting themes.
- The page has no JavaScript.
- New GitHub Actions workflow that builds the site and deploys it to Pages on every push to `main`. It is the first workflow in the repository.

Out of scope:

- An in-browser playground (termd compiled to WebAssembly) - a separate change later.
- A `--color=always` CLI flag.
- Deploying on tags instead of pushes to `main` - belongs to the release flow.
- Remembering the manual theme choice across page loads.

## Capabilities

### New Capabilities
- `project-site`: content of the landing page, generation of the examples from termd output, the width and theme switchers, and deployment to GitHub Pages.

### Modified Capabilities

None - the rendering and CLI behavior of termd do not change.

## Impact

- New Go generator program in the same module. It imports `internal/render`, `internal/text`, `internal/highlight` and `internal/style`; these packages do not change.
- New directory with the page template, CSS and example documents.
- New workflow under `.github/workflows/`.
- No new Go dependencies: the page template uses `html/template` from the standard library.
- `make check` must cover the generator; `fmt-check` currently checks only `cmd` and `internal`.
- Manual steps in the repository settings: set the Pages source to "GitHub Actions" and fill in the Homepage field after the first deploy.
- The install section can show only `go install` until the release flow exists.
