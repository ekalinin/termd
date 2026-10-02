# Proposal

## Why

GitHub renders a block quote that starts with `[!NOTE]`, `[!TIP]`, `[!IMPORTANT]`, `[!WARNING]` or `[!CAUTION]` as an alert with a title and a color, and such alerts are common in READMEs. termd renders them as a plain quote with the marker as text: `│ [!NOTE] Useful info.`.

## What Changes

- A block quote whose first line is one of the five markers, alone on the line and compared case-insensitively, is rendered as an alert: the marker line becomes the title `Note`, `Tip`, `Important`, `Warning` or `Caution`, and the rest of the quote is rendered as a block quote under it.
- When styling is enabled, the quote marker and the title use the color of the alert type, as on GitHub: blue for note, green for tip, purple for important, yellow for warning, red for caution. When styling is disabled, the title is plain text after the usual quote marker.
- A quote with another marker (for example `[!FOO]`) or with text after the marker on the first line is rendered as a regular block quote.
- Alerts have no icons: terminals disagree on the width of emoji, which would break the layout.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `markdown-rendering`: a new requirement for GitHub alerts.

## Impact

- `internal/render`: the block quote rendering detects the marker in the first paragraph and renders the title; no new goldmark extension is needed.
- New golden fixture with the five alert types and the fallback cases.
- `README.md`: alerts in "How it works".
