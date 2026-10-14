# Spec Delta

## MODIFIED Requirements

### Requirement: GFM parsing
termd SHALL parse documents as CommonMark with the GitHub Flavored Markdown extensions for tables, task lists, strikethrough, autolinks and footnotes, with definition lists and with GitHub emoji shortcodes.

#### Scenario: GFM extensions are recognized
- **WHEN** a document contains a pipe table, a `- [x]` task item, `~~text~~`, a bare `https://` URL and a footnote reference `[^1]` with the definition `[^1]: Note.`
- **THEN** they are rendered as a table, a checked task item, struck-through text, a link and a footnote respectively

## ADDED Requirements

### Requirement: Emoji shortcodes
termd SHALL replace a GitHub emoji shortcode, the name of a GitHub emoji between two colons such as `:rocket:`, with the Unicode characters of the emoji. This SHALL apply to the text of paragraphs, headings, list items, block quotes and alerts, table cells, link text, footnotes and definition lists, including the text between inline HTML tags such as `<kbd>:zap:</kbd>`. Replacement SHALL always be on.

A shortcode SHALL stay as text when the character right before its first colon is a letter or a digit, so `a:b:c` and `foo:rocket:bar` do not change. A name that is not the name of a GitHub emoji with a Unicode character, such as `:octocat:`, SHALL stay as text. Shortcodes in inline code, code blocks, diagrams, link destinations, bare URLs, image alt text, HTML blocks and frontmatter SHALL NOT be replaced.

The emoji SHALL take the style of the surrounding text, SHALL be shown when styling is disabled too, and SHALL be measured by its display width.

#### Scenario: Shortcodes in a paragraph
- **WHEN** a paragraph contains `Fast :rocket: and tested :white_check_mark:` and styling is disabled
- **THEN** the output is `Fast 🚀 and tested ✅`

#### Scenario: Shortcode in a heading
- **WHEN** a document contains `# Features :sparkles:` and styling is disabled
- **THEN** the output is `# Features ✨`

#### Scenario: Shortcode in a table cell
- **WHEN** a table has the header `Feature | Status` and the rows `Tables | :white_check_mark:` and `Diagrams | :warning:`
- **THEN** the cells read `✅` and `⚠️`, each is measured as 2 columns, and the column separator is at the same column on every line of the table

#### Scenario: Shortcode in an alert
- **WHEN** a document contains `> [!NOTE]` followed by `> :warning: Read this.` and styling is disabled
- **THEN** the output is `│ Note` followed by `│ ⚠️ Read this.`

#### Scenario: Link text and destination
- **WHEN** a document contains `[docs :book:](https://example.com/:book:)` and hyperlinks are disabled
- **THEN** the output is `docs 📖 (https://example.com/:book:)`

#### Scenario: Shortcode between inline HTML tags
- **WHEN** a paragraph contains `Press <kbd>:arrow_up:</kbd> to scroll.`
- **THEN** the output is `Press ⬆️ to scroll.`

#### Scenario: Shortcode after a letter or a digit
- **WHEN** a paragraph contains `a:b:c, foo:rocket:bar, 1:zap: and x :b: y`
- **THEN** the output is `a:b:c, foo:rocket:bar, 1:zap: and x 🅱️ y`

#### Scenario: Adjacent shortcodes
- **WHEN** a paragraph contains `:rocket::zap:` and `**:tada:**`
- **THEN** the output contains `🚀⚡`, and `🎉` in bold when styling is enabled

#### Scenario: Names that are not emoji
- **WHEN** a paragraph contains `:octocat: and :not_an_emoji:`
- **THEN** the output is `:octocat: and :not_an_emoji:`

#### Scenario: Code and bare URLs
- **WHEN** a document contains the paragraph `` `:zap:` and https://example.com/:smile:/x `` and a fenced code block with the line `echo :zap:`, and hyperlinks are disabled
- **THEN** the output contains `:zap: and https://example.com/:smile:/x` and the code line `echo :zap:`

#### Scenario: Image alt text
- **WHEN** a document contains `![logo :star:](logo.png)` and hyperlinks are disabled
- **THEN** the output is `[image: logo :star:] (logo.png)`

#### Scenario: HTML block
- **WHEN** a document contains `<p align="center">:rocket: Fast</p>`
- **THEN** the output is `:rocket: Fast`

#### Scenario: Frontmatter
- **WHEN** a document starts with the frontmatter `---`, `title: ":rocket: Launch"`, `---`
- **THEN** the value cell of `title` reads `:rocket: Launch`
