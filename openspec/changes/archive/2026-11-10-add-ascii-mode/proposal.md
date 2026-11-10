# Proposal

## Why

termd draws tables, quote markers, horizontal lines, list bullets, diagram frames and diagrams with Unicode characters (`│`, `─`, `┼`, `•`, `┌`), also in plain mode. Some fonts, terminals and destinations of the output, such as logs or email, do not show them well. glamour has an `ascii` style that draws tables, quotes and lines with ASCII characters.

## What Changes

- A new flag `--ascii`. With it, every character termd draws itself is ASCII:

  | Element | Unicode | ASCII |
  |---|---|---|
  | table separators and rule | `│`, `─`, `┼` | `\|`, `-`, `+` |
  | quote and alert marker | `│` | `\|` |
  | thematic break and footnote line | `─` | `-` |
  | list bullet | `•` | `*` |
  | frame of diagram source and invalid frontmatter | `┌`, `─`, `┐`, `│`, `└`, `┘` | `+`, `-`, `\|` |
  | `details` marker | `▾` | `v` |
  | mindmap branches | `├──`, `│`, `└──` | `\|--`, `\|`, `` `-- `` |
  | gantt bars: task, active task, done task, milestone | `█`, `▓`, `░`, `◆` | `#`, `=`, `-`, `*` |
  | state diagram start, end, choice, fork and join | `●`, `◉`, `◇`, `━━━` | `*`, `(*)`, `<>`, `===` |
  | line between concurrent regions of a state | `┈` | `.` |

- Mermaid diagrams are drawn with ASCII characters: mermaid-ascii has a `UseAscii` setting, supported by flowcharts, sequence diagrams and ER diagrams. termd draws state diagrams and PlantUML sequence and state diagrams with mermaid-ascii too, so they get the setting as well.
- In the text of a state, termd today replaces a `-` after a `-` and a `>` after a `-` or a `=` with look-alikes (`‑`, `›`), so that mermaid-ascii does not read the text as a link. In ASCII mode these characters get a space before them instead: `a --> b` is shown as `a - - > b`.
- The flag works in a terminal and in a pipe and does not change colors, themes or hyperlinks.
- Without the flag, the output does not change.

Out of scope:

- Changing the text of the document: non-ASCII characters in the document stay as they are, and so do emoji of shortcodes such as `:tada:` and characters of references such as `&copy;`.
- Turning the mode on by itself, for example when the locale (`LC_ALL`, `LC_CTYPE`, `LANG`) is not UTF-8: in docker, cron and ssh the locale is often not set while the terminal shows Unicode.
- Control pictures (`␛`) and the replacement character (`�`) that termd shows for control characters and invalid bytes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `cli`: a new requirement for the `--ascii` flag.
- `markdown-rendering`: list bullets, quote markers, thematic breaks, the footnote line and the `details` marker in ASCII mode.
- `table-rendering`: column separators and the header rule in ASCII mode.
- `diagram-rendering`: diagrams, including mindmaps, gantt charts and state diagrams, and the frame of diagram source in ASCII mode; the "Supported mermaid types" requirement names Unicode box-drawing characters only.

## Impact

- `cmd/termd/main.go`: the flag and the usage text.
- `internal/render`: `Options` gets the mode; bullets, quote markers, rules and the `details` marker take their characters from it.
- `internal/table`: `Table` gets the mode for its separator characters.
- `internal/diagram`: `Render` and `Frame` get the mode; `Render` sets `UseAscii` in the mermaid-ascii config and passes the mode to the mindmap, gantt and state renderers.
- Golden files: a set for ASCII mode; the existing golden files must not change.
- `README.md`: the flags table and "How it works".
