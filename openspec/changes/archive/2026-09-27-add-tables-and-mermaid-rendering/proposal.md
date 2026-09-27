# Proposal

## Why

Existing terminal markdown viewers render tables badly and do not render diagrams at all. Rendering a sample document with glow 2.1.2 showed truncated column headers ("По умолчанию" became "По …"), short words split mid-word ("string" became "strin" + "g"), multi-codepoint emoji measured with the wrong width, link URLs cut off with an ellipsis, and mermaid blocks shown as raw source. termd is a new console application that renders markdown documents correctly, starting with the two things that are broken today: tables and mermaid diagrams.

## What Changes

- New Go command-line application `termd` that reads a markdown document from a file or stdin and renders it to the terminal.
- GFM parsing and rendering of common block and inline elements (headings, paragraphs, emphasis, lists, task lists, block quotes, code blocks, thematic breaks).
- Syntax highlighting of fenced code blocks whose info string names a known language, when writing to a terminal, with truecolor light and dark themes chosen by the terminal's background color (`--theme=auto|dark|light`).
- Display width measured by grapheme clusters, so emoji sequences (ZWJ, variation selectors) and CJK characters do not break alignment.
- Table layout that never splits words shorter than the column minimum, never truncates headers, distributes free width to the columns with long text and keeps column alignment from the markdown source.
- Tables that do not fit even at their minimum widths are emitted at their natural width and stay horizontally scrollable in the pager.
- Clickable links through OSC 8 escape sequences, with a plain-text fallback and a `--hyperlinks=auto|always|never` flag.
- Mermaid `sequenceDiagram`, `flowchart`/`graph` and `erDiagram` blocks rendered as text diagrams using the `AlexanderGrooff/mermaid-ascii` library.
- Any other diagram (other mermaid types, PlantUML, a diagram that fails to parse) shown as its source in a labelled frame instead of failing.
- Output streamed to stdout; when stdout is a terminal and the output is taller than the screen, it is shown through `less -RS`.

Out of scope for this change (planned as later changes): mermaid `stateDiagram` and `mindmap` rendering (v2), PlantUML sequence rendering (v3), inline images and a built-in TUI (later).

## Capabilities

### New Capabilities
- `cli`: invocation, input sources, flags, output width detection, terminal vs pipe behavior and pager integration.
- `markdown-rendering`: rendering of GFM block and inline elements, display width measurement, wide block handling and hyperlinks.
- `table-rendering`: layout of GFM tables - column width allocation, wrapping, alignment and overflow.
- `diagram-rendering`: detection of diagram code blocks, text rendering of supported mermaid types and the framed-source fallback.

### Modified Capabilities

None - this is the first change in the project.

## Impact

- New Go module (the project currently contains only OpenSpec scaffolding).
- New dependencies: a CommonMark/GFM parser for Go, a grapheme-aware display width library, `github.com/AlexanderGrooff/mermaid-ascii` (MIT, no tagged releases, pinned by commit), `github.com/alecthomas/chroma/v2` (syntax highlighting).
- Runtime dependency on `less` (version 566 or later for OSC 8 pass-through) for paging; termd works without it by printing directly.
- Target environment: modern local terminals (iTerm2 in particular), also tmux, screen and SSH sessions.
