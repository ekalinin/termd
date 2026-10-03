# Spec Delta

## MODIFIED Requirements

### Requirement: GitHub alerts
When a block quote starts with a paragraph whose first line is `[!NOTE]`, `[!TIP]`, `[!IMPORTANT]`, `[!WARNING]` or `[!CAUTION]`, alone on the line and compared case-insensitively, termd SHALL render the block quote as an alert: the marker line SHALL be replaced by the title `Note`, `Tip`, `Important`, `Warning` or `Caution`, and the rest of the quote SHALL follow under the title, laid out by the block quote rules. A block quote that holds only the marker line SHALL be rendered as the title alone. Block quotes nested in other block quotes or in list items SHALL be detected as alerts the same way.

When styling is enabled, the title and the quote marker on every line of the alert SHALL be shown in the color the theme gives the alert type. With the `dark` and `light` themes the color SHALL be taken from the basic palette of the terminal: blue for note, green for tip, purple (magenta) for important, yellow for warning and red for caution. The title SHALL also be bold. When styling is disabled, the title SHALL be plain text after the usual quote marker. Alerts SHALL NOT have icons.

A block quote whose first line holds another marker, for example `[!FOO]`, or text after the marker SHALL be rendered as a regular block quote, with the marker as text.

#### Scenario: Note alert
- **WHEN** a document contains `> [!NOTE]` and `> Useful info.` and is rendered to a pipe
- **THEN** the output is `│ Note` followed by `│ Useful info.`, with no escape sequences

#### Scenario: Alert colors
- **WHEN** a note, a tip, an important, a warning and a caution alert are rendered to a terminal with the `dark` or the `light` theme
- **THEN** their quote markers are blue, green, magenta, yellow and red (SGR `34`, `32`, `35`, `33` and `31`) and not faint, and each title is bold in the color of its alert, for example `Note` in SGR `1;34`

#### Scenario: Alert colors of a named theme
- **WHEN** a warning alert is rendered to a terminal with `COLORTERM=truecolor` and `--theme=nord`
- **THEN** its quote markers are in `#ebcb8b` and its title is bold in `#ebcb8b`

#### Scenario: Case-insensitive marker
- **WHEN** a document contains `> [!warning]` and `> Careful.`
- **THEN** the output is `│ Warning` followed by `│ Careful.`

#### Scenario: Marker only
- **WHEN** a document contains only `> [!TIP]`
- **THEN** the output is the single line `│ Tip`

#### Scenario: Several blocks
- **WHEN** a document contains `> [!IMPORTANT]`, `> First.`, `>` and `> Second.`
- **THEN** the output is `│ Important`, `│ First.`, `│` and `│ Second.`

#### Scenario: Long alert
- **WHEN** a note alert holds a paragraph longer than the output width of 30
- **THEN** every output line of the alert starts with the quote marker and no line is wider than 30 columns

#### Scenario: Nested quote inside an alert
- **WHEN** a document contains `> [!NOTE]`, `> Text.`, `>` and `> > Quoted.`
- **THEN** the output is `│ Note`, `│ Text.`, `│` and `│ │ Quoted.`, and in a terminal the inner quote marker has the style of the quote marker of a regular quote, which is faint with the `dark` and `light` themes

#### Scenario: Alert in a list item
- **WHEN** a list item `- item` continues with the indented lines `> [!TIP]` and `> Hint.`
- **THEN** the output is `• item`, `  │ Tip` and `  │ Hint.`

#### Scenario: Text after the marker
- **WHEN** a document contains `> [!NOTE] Useful info.`
- **THEN** it is rendered as a regular block quote: `│ [!NOTE] Useful info.`

#### Scenario: Unknown marker
- **WHEN** a document contains `> [!FOO]` and `> Text.`
- **THEN** it is rendered as a regular block quote: `│ [!FOO] Text.`
