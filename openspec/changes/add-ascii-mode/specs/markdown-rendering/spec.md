## ADDED Requirements

### Requirement: ASCII mode
In ASCII mode, termd SHALL draw these elements with ASCII characters instead of Unicode ones, at the same columns and with the same styles:

- the bullet of an unordered list item: `*` instead of `•`;
- the marker of a block quote and of a GitHub alert: `|` instead of `│`;
- the line of a thematic break and the line before the footnotes: `-` instead of `─`;
- the marker of a `details`: `v` instead of `▾`;
- the frame of a frontmatter block that is not valid YAML: the frame of diagram source in ASCII mode, as described in the "ASCII mode" requirement of the diagram-rendering capability.

The key-value table of frontmatter SHALL be drawn as a table in ASCII mode, as described in the table-rendering capability. ASCII mode SHALL NOT change the text of the document: non-ASCII characters of the document, the emoji of shortcodes such as `:tada:`, the characters of character references such as `&copy;`, the control pictures of control characters and the replacement character SHALL be shown as without ASCII mode.

#### Scenario: Lists and quotes
- **WHEN** a document contains `- a`, `  - b` and, after a blank line, `> quoted`, and it is rendered in ASCII mode with styling disabled
- **THEN** the output is `* a`, `  * b`, a blank line and `| quoted`

#### Scenario: Alert
- **WHEN** a document contains `> [!NOTE]` followed by `> Useful info.` and it is rendered in ASCII mode with styling disabled
- **THEN** the output is `| Note` followed by `| Useful info.`

#### Scenario: Thematic break and footnotes
- **WHEN** the document `Text[^1].`, a blank line, `---`, a blank line, `[^1]: Note.` is rendered in ASCII mode at width 20 with styling and hyperlinks disabled
- **THEN** the thematic break and the line before the footnotes are both 20 `-`, and the footnote is `1. Note.`

#### Scenario: Details
- **WHEN** a document contains `<details><summary>More</summary>text</details>` and it is rendered in ASCII mode with styling disabled
- **THEN** the output is `v More` followed by `  text`

#### Scenario: Styles in ASCII mode
- **WHEN** a document contains `> quoted` and `- item` and it is rendered in ASCII mode with styling enabled
- **THEN** `|` has the marker style of the theme, and `*` has the style that `•` has without ASCII mode

#### Scenario: Invalid YAML frontmatter
- **WHEN** a document starts with `---`, `title: [unclosed`, `---` and it is rendered in ASCII mode
- **THEN** `title: [unclosed` is shown in a frame of `+`, `-` and `|` labelled `frontmatter - invalid YAML`

#### Scenario: Text of the document
- **WHEN** a document contains `Café :tada: &copy;` and a control character ESC, and it is rendered in ASCII mode
- **THEN** the output contains `Café 🎉 ©` and `␛`
