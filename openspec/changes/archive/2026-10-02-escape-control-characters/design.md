# Design

## Context

`render.Render` is the only entry point for both `cmd/termd` and the site generator (`internal/site`). It splits off the frontmatter (`splitFrontmatter`), parses the body with goldmark (`Parse`), builds blocks from goldmark segments that point into the source, lays them out and only then turns each line into bytes with `text.Line.Render`, which adds the SGR and OSC 8 sequences.

The layout measures text with `text.Width` (uniseg) before `Line.Render` runs: paragraph wrapping, table column widths, list and quote prefixes, frames. mermaid-ascii lays out diagrams with its own measurement, from the diagram source. uniseg measures ESC and BEL as 0 columns, a control picture as 1.

Document text reaches the output in two ways:

- Taken from the source without decoding: paragraph and heading text, code spans, code blocks (plain, highlighted by chroma, or passed to mermaid-ascii), HTML blocks and inline HTML, link destinations (termd writes `n.Destination` as it is in the source), autolinks, the invalid YAML frame, framed diagram source.
- Decoded, so the output can hold characters the source does not: character references in text and image alt text, which `unescape` resolves (goldmark turns `&#27;` into ESC; only `&#0;` becomes U+FFFD), and double-quoted YAML strings in the frontmatter (`"\e"` is ESC, `"\x9b"` is U+009B). The current output of termd confirms both: `&#27;]0;x&#7;` and `title: "\e]0;x\a"` write a raw title sequence.

goldmark keeps the `\r` of a CRLF line in code block segments, which is why code blocks leak it. `splitFrontmatter` already trims `\r` from the frontmatter lines itself. yaml.v3 rejects raw control characters, so today a raw ESC in the frontmatter produces the invalid YAML frame with the ESC inside.

The requirements are in `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- Replace the characters before anything is measured, so the layout measures exactly what is printed.
- One mapping from a control character to what is shown, used by every path.
- Documents without control characters render byte for byte as before; the existing golden files do not change.

**Non-Goals:**
- Percent-encoding link destinations for OSC 8. A destination keeps the control pictures as UTF-8 text.
- A control character example on the landing page. The site gets the behavior through `render.Render`.
- Changing `splitFrontmatter`: its BOM and CRLF handling stays for direct callers, although through `Render` it no longer sees a CR.

## Decisions

### Clean the source once, at the start of `render.Render`

`Render` calls `cleanSource(src)` before `splitFrontmatter` and `Parse`. It turns CRLF into LF, then every remaining CR into LF, then replaces the control characters and the invalid UTF-8. Everything that is built from the source afterwards, including goldmark segments, chroma tokens, mermaid-ascii input and output, frames and the frontmatter, sees clean text, and every measurement counts a control picture as the one column it takes. The caller's slice is never modified; the cleaned source is a copy.

The line endings are converted before the parser, as CommonMark asks, so a CR is never shown as `␍` when it ends a line, and the frontmatter delimiters are found as before.

Alternatives:
- Replace the characters in `text.Line.Render`, where span text and link destinations are written. Rejected: the layout has already measured ESC as 0 columns, so every line holding a control picture would come out wider than measured, which breaks wrapping at the output width, table column separators and frame borders. Diagrams would still be laid out by mermaid-ascii from the raw text, and the CR of a CRLF line, which is a line ending and not a character to show, would still have to be removed before parsing.
- Filter the final output in `cmd/termd`. Rejected: at that point the sequences termd emits cannot be told apart from those of the document, and the site generator would not be covered.
- A goldmark option or extension. Rejected: goldmark has no hook for the source text, and the frontmatter never reaches goldmark.

### Clean the text the decoders produce

Two decoders create characters after the source is cleaned, and each gets the same mapping right after it runs:

- `unescape`, which resolves backslash escapes and character references for text nodes and image alt text, maps its result. It is the only place where termd resolves character references; link destinations are written as they are in the source, so `&#27;` in a destination stays the five characters `&#27;`.
- The frontmatter mapping is cleaned in place before the table is built: a walk over the `yaml.Node` tree maps the `Value` of every node. This covers keys, scalar values, list items, nested mappings and the flow-style text of deeper nesting, which yaml.v3 would otherwise print with its own escapes (`"\e"`), with no change to the value formatting code.

A CR that comes from a decoder (`&#13;`, `"\r"`) is a character of the text, not a line ending of the source, so it is shown as `␍` like the other control characters.

Alternative: clean only the source. Rejected: `&#27;]0;pwned&#7;` and `title: "\e]0;pwned\a"` would still change the window title, which the requirement forbids ("control characters that the document writes as character references or YAML escapes").

### One rune mapping

A single function `controlPicture(r rune) rune` decides what a character becomes:

| Character | Shown as |
|---|---|
| tab, line feed | unchanged |
| other U+0000 to U+001F | U+2400 + code point (`␀` to `␟`) |
| DEL U+007F | `␡` U+2421 |
| U+0080 to U+009F | `�` U+FFFD |
| anything else | unchanged |

It is applied with `bytes.Map` to the source and to the result of `unescape`, and with `strings.Map` to YAML values. Both functions decode invalid UTF-8 as U+FFFD, one per invalid byte, so the invalid bytes need no separate rule.

Alternative: `bytes.ToValidUTF8` for the invalid bytes, which writes one U+FFFD per run of invalid bytes. Rejected: it adds a second pass and a second rule for no visible benefit; one U+FFFD per byte is what Go's UTF-8 decoding gives everywhere else.

NUL follows the proposal: a raw NUL becomes `␀`, because the source is cleaned before goldmark sees it. `&#0;` is already turned into U+FFFD by goldmark, as CommonMark requires, and stays `�`.

### Code layout

`cleanSource` and `controlPicture` live in a new file `internal/render/control.go`. The YAML walk lives in `internal/render/frontmatter.go`, next to the other YAML code, so `control.go` does not import yaml. `internal/text` does not change: it measures and wraps what it is given, and only `render` knows where text comes from.

### Tests and fixture

Unit tests in `internal/render/render_test.go` build their input in Go with escaped bytes (`"\x1b"`, `"\r\n"`), one test per path of the spec, and assert the exact output. A `cmd/termd` test checks that pipe output contains no ESC when the document does, as the `cli` spec promises.

The golden fixture `testdata/control.md` holds the raw bytes, because it must show the behavior end to end at all widths and in both modes. To keep it readable, every line says in words which bytes it contains (for example "a color sequence, ESC [31m"), and the golden files show the bytes as control pictures, so they read as plain text. Its lines end with CRLF, with a lone CR in a paragraph and in a code block. It contains no NUL, so git does not show it as a binary file in diffs; NUL is covered by a unit test.

## Risks / Trade-offs

- [Form feed and vertical tab become pictures, so they no longer act as whitespace in markdown syntax] → Accepted: the proposal keeps only tab and line feed, and these characters are rare in markdown.
- [A frontmatter block with a raw control character was shown in the invalid YAML frame and now becomes a table] → Intended: the control pictures are valid YAML, and the value is shown as text like any other.
- [A link destination with a control picture is not a valid URL for OSC 8, which expects ASCII] → The terminal gets UTF-8 text inside the OSC 8 sequence and can no longer be controlled by it; percent-encoding is a non-goal.
- [The source is copied up to three times] → Documents are small and the copies are linear; parsing and layout dominate the run time.
- [Text decoded by future code that bypasses `unescape` or the YAML walk would not be cleaned] → New paths for decoded text call `controlPicture`; the scenarios of the requirement are unit tests that cover each current path.
- [git end-of-line conversion (`core.autocrlf`) could rewrite the CRLF lines of the fixture] → git treats a file with a lone CR as binary for this conversion and leaves it as it is; the unit tests build CRLF input in Go as well.

## Migration Plan

None. Documents without control characters render as before, so the existing golden files do not change. Rollback is a revert of the change.
