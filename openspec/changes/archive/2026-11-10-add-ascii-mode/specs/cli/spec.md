## ADDED Requirements

### Requirement: ASCII mode flag
termd SHALL accept `--ascii`. With it, every character that termd draws itself SHALL be an ASCII character, as described in the "ASCII mode" requirements of the markdown-rendering, table-rendering and diagram-rendering capabilities, in a terminal and in a pipe. The flag SHALL NOT change styles, the theme or hyperlinks. termd SHALL NOT turn the mode on without the flag, whatever the locale. Without the flag, the output SHALL NOT change.

#### Scenario: ASCII output in a pipe
- **WHEN** the user runs `termd --ascii README.md | cat` and the file contains the list item `- item` and the block quote `> quoted`
- **THEN** the output contains `* item` and `| quoted`

#### Scenario: Styles and hyperlinks are kept
- **WHEN** the user runs `termd --ascii README.md` in a terminal and the file contains `> [docs](https://example.com)`
- **THEN** the quote marker is `|` with the marker style of the theme, and `docs` is a terminal hyperlink to `https://example.com`

#### Scenario: Locale without UTF-8
- **WHEN** the user runs `LANG=C termd README.md | cat` without the flag and the file contains `- item`
- **THEN** the output contains `• item`

#### Scenario: Without the flag
- **WHEN** the user runs `termd README.md | cat` without the flag
- **THEN** the output is the same as before this change
