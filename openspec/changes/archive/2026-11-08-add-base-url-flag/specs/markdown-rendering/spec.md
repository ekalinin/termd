## MODIFIED Requirements

### Requirement: Hyperlinks
When hyperlinks are enabled, termd SHALL render each link as its visible text wrapped in an OSC 8 terminal hyperlink to the link destination, and SHALL measure only the visible text. When hyperlinks are disabled, termd SHALL render a link as `text (url)`, or as just the URL when the text equals the URL, except in a table cell, where the URL is listed below the table as described in "Links in cells without hyperlinks" of the table-rendering capability. termd SHALL NOT truncate URLs.

When the document is read from a file, hyperlinks are enabled and no base URL is given, termd SHALL resolve a relative destination of a link or an image against the directory of that file and write it into the hyperlink as an absolute `file:` URL. A relative destination is a destination without a scheme that has a path and does not start with `/`, for example `docs/guide.md`, `./a.md` or `../b.md`. The URL SHALL consist of `file://`, the host name of the machine, the absolute path of the target with the `.` and `..` segments resolved, and the fragment of the destination, if any; the query of the destination SHALL be dropped. Backslash escapes, character references and percent-encoded characters of the destination SHALL be decoded, and every character that a URL does not allow in a path, such as a space or a non-ASCII character, SHALL be percent-encoded, so that no character is encoded twice. When the host name of the machine cannot be determined, and on Windows, the host SHALL be empty.

When a base URL is given with `--base-url` and hyperlinks are enabled, termd SHALL resolve destinations of links and images against the base URL instead of the directory of the document file, for a document from a file and from stdin, and write them into the hyperlink as absolute URLs; termd SHALL NOT make `file:` URLs from the directory of the document file then:

- a relative destination SHALL be resolved by RFC 3986: when the path of the base URL does not end with `/`, its last segment SHALL be replaced, the `.` and `..` segments SHALL be resolved, and the query and the fragment of the destination SHALL be kept;
- a destination that starts with a single `/` SHALL lose the `/` and be resolved as a relative destination;
- a destination that starts with `//` SHALL get the scheme of the base URL.

Backslash escapes and character references of such a destination SHALL be decoded, and the path and the fragment SHALL be percent-encoded as for `file:` URLs, so that no character is encoded twice; the query SHALL keep its percent-encoded characters as written.

termd SHALL write a destination into the hyperlink as written in the document when it has a scheme (`https:`, `mailto:` and any other), when it is fragment-only (`#section`), when it is empty, and when it is not a valid URL reference, and, without a base URL, when it starts with `/` (including `//host/path`). When the document is read from stdin without a base URL, or when hyperlinks are disabled, termd SHALL NOT resolve any destination.

When hyperlinks are enabled, termd SHALL show an autolink, a bare URL or a URL in `<` and `>`, to an issue, a pull request, a discussion or a commit on GitHub with a short text, and its hyperlink SHALL be the full URL. Such a URL SHALL consist of `https://github.com/` or `http://github.com/`, with the scheme and the host in any case, `owner/repo` and one of these endings, where `N` is a number and `SHA` is 7 to 40 hexadecimal digits:

- `/issues/N`, `/pull/N` and `/discussions/N`: the text is `owner/repo#N`;
- the same endings followed by `#issuecomment-`, `#discussion_r` or `#discussioncomment-` and digits: the text is `owner/repo#N (comment)`;
- `/pull/N#pullrequestreview-` followed by digits: the text is `owner/repo#N (review)`;
- `/commit/SHA` and `/pull/N/commits/SHA`: the text is `owner/repo@` followed by the first 7 characters of `SHA`.

The owner SHALL consist of ASCII letters, digits, `-` and `_`, the repository of ASCII letters, digits, `-`, `_` and `.`, and both SHALL be shown as written. termd SHALL show any other autolink as its URL, including a URL of another host or of `www.github.com`, a URL with a query, a trailing `/` or another fragment, and a URL of another GitHub page. A link with its own text, such as `[text](url)`, SHALL keep its text. When hyperlinks are disabled, termd SHALL show the full URL of every autolink outside table cells; in a table cell, the full URL SHALL be listed below the table.

#### Scenario: Link in a terminal
- **WHEN** a document contains `[docs](https://example.com/very/long/path)` and hyperlinks are enabled
- **THEN** the visible output is `docs`, it opens `https://example.com/very/long/path` when clicked, and it occupies 4 columns

#### Scenario: Link in a pipe
- **WHEN** the same document is rendered with hyperlinks disabled
- **THEN** the output contains `docs (https://example.com/very/long/path)` with the full URL

#### Scenario: Autolink
- **WHEN** a document contains the bare URL `https://example.com` and hyperlinks are disabled
- **THEN** the output contains `https://example.com` once

#### Scenario: Relative link in a file
- **WHEN** the file `/home/u/proj/README.md` contains `[guide](docs/guide.md)`, it is rendered with hyperlinks enabled, and the host name of the machine is `box`
- **THEN** the visible output is `guide` and its hyperlink is `file://box/home/u/proj/docs/guide.md`

#### Scenario: Relative image in a file
- **WHEN** the same file contains `![arch](img/arch.png)`
- **THEN** the visible output is `[image: arch]` and its hyperlink is `file://box/home/u/proj/img/arch.png`

#### Scenario: Dot segments
- **WHEN** the same file contains `[a](./a.md)` and `[b](../other/b.md)`
- **THEN** their hyperlinks are `file://box/home/u/proj/a.md` and `file://box/home/u/other/b.md`

#### Scenario: Spaces and non-ASCII characters
- **WHEN** the same file contains `[n](<my notes.md>)`, `[n](my%20notes.md)` and `[f](файл.md)`
- **THEN** the first two hyperlinks are both `file://box/home/u/proj/my%20notes.md` and the third is `file://box/home/u/proj/%D1%84%D0%B0%D0%B9%D0%BB.md`

#### Scenario: Fragment and query
- **WHEN** the same file contains `[setup](guide.md#setup)` and `![logo](logo.png?raw=true)`
- **THEN** their hyperlinks are `file://box/home/u/proj/guide.md#setup` and `file://box/home/u/proj/logo.png`

#### Scenario: Destinations that are not resolved
- **WHEN** the same file contains links to `https://example.com`, `mailto:me@example.com`, `#usage`, `/docs/x.md` and `//example.com/x`
- **THEN** the hyperlink of each link is its destination exactly as written

#### Scenario: Empty destination
- **WHEN** the same file contains `[empty]()`
- **THEN** the output contains `empty` without a hyperlink

#### Scenario: Relative file argument
- **WHEN** the user runs `termd --hyperlinks=always docs/README.md` in the directory `/home/u/proj` and the file contains `[a](a.md)`
- **THEN** the hyperlink of `a` is `file://box/home/u/proj/docs/a.md`

#### Scenario: Document from stdin
- **WHEN** the user runs `cat README.md | termd --hyperlinks=always` and the document contains `[guide](docs/guide.md)`
- **THEN** the hyperlink of `guide` is `docs/guide.md`

#### Scenario: Hyperlinks disabled
- **WHEN** the file `/home/u/proj/README.md` contains `[guide](docs/guide.md)` and it is rendered with hyperlinks disabled
- **THEN** the output contains `guide (docs/guide.md)`

#### Scenario: Unknown host name
- **WHEN** the host name of the machine cannot be determined and the file `/home/u/proj/README.md` contains `[guide](docs/guide.md)`
- **THEN** the hyperlink of `guide` is `file:///home/u/proj/docs/guide.md`

#### Scenario: Windows
- **WHEN** termd runs on Windows and the file `C:\proj\README.md` contains `[guide](docs/guide.md)`
- **THEN** the hyperlink of `guide` is `file:///C:/proj/docs/guide.md`

#### Scenario: Base URL
- **WHEN** a document from stdin contains `[guide](docs/guide.md)` and `![logo](img/logo.png?raw=true)`, and it is rendered with hyperlinks enabled and the base URL `https://github.com/o/r/blob/main/`
- **THEN** their hyperlinks are `https://github.com/o/r/blob/main/docs/guide.md` and `https://github.com/o/r/blob/main/img/logo.png?raw=true`

#### Scenario: Base URL of a document
- **WHEN** a document contains `[guide](docs/guide.md)` and it is rendered with hyperlinks enabled and the base URL `https://raw.githubusercontent.com/o/r/main/README.md`, and then with the base URL `https://github.com/o/r/blob/main`
- **THEN** the hyperlink of `guide` is `https://raw.githubusercontent.com/o/r/main/docs/guide.md` and then `https://github.com/o/r/blob/docs/guide.md`

#### Scenario: Base URL with dot segments, fragments and spaces
- **WHEN** a document contains `[b](../other/b.md)`, `[setup](guide.md#setup)` and `[n](<my notes.md>)`, and it is rendered with hyperlinks enabled and the base URL `https://example.com/docs/`
- **THEN** their hyperlinks are `https://example.com/other/b.md`, `https://example.com/docs/guide.md#setup` and `https://example.com/docs/my%20notes.md`

#### Scenario: Base URL and a leading slash
- **WHEN** a document contains `[x](/docs/x.md)` and `[e](//example.com/x)`, and it is rendered with hyperlinks enabled and the base URL `https://github.com/o/r/blob/main/`
- **THEN** their hyperlinks are `https://github.com/o/r/blob/main/docs/x.md` and `https://example.com/x`

#### Scenario: Base URL over the directory of the file
- **WHEN** the file `/home/u/proj/README.md` contains `[a](a.md)` and it is rendered with hyperlinks enabled and the base URL `https://example.com/docs/`
- **THEN** the hyperlink of `a` is `https://example.com/docs/a.md`

#### Scenario: File base URL
- **WHEN** a document from stdin contains `[guide](docs/guide.md)` and it is rendered with hyperlinks enabled and the base URL `file:///home/me/proj/`
- **THEN** the hyperlink of `guide` is `file:///home/me/proj/docs/guide.md`

#### Scenario: Destinations that are not resolved against the base URL
- **WHEN** a document contains links to `https://example.com`, `mailto:me@example.com` and `#usage`, and `[empty]()`, and it is rendered with hyperlinks enabled and the base URL `https://github.com/o/r/blob/main/`
- **THEN** the hyperlink of each of the first three links is its destination exactly as written, and `empty` has no hyperlink

#### Scenario: HTML with a base URL
- **WHEN** a document contains `<img src="img/arch.png" alt="arch">` and `<a href="docs/guide.md">guide</a>`, and it is rendered with hyperlinks enabled and the base URL `https://example.com/`
- **THEN** their hyperlinks are `https://example.com/img/arch.png` and `https://example.com/docs/guide.md`

#### Scenario: Base URL with hyperlinks disabled
- **WHEN** a document contains `[guide](docs/guide.md)` and it is rendered with hyperlinks disabled and the base URL `https://example.com/`
- **THEN** the output contains `guide (docs/guide.md)`

#### Scenario: GitHub pull request
- **WHEN** a document contains the bare URL `https://github.com/ekalinin/termd/pull/19` and hyperlinks are enabled
- **THEN** the visible output is `ekalinin/termd#19`, it opens `https://github.com/ekalinin/termd/pull/19` when clicked, and it occupies 17 columns

#### Scenario: GitHub issue and discussion
- **WHEN** a document contains `<https://github.com/o/r/issues/3>` and `https://github.com/o/r/discussions/7` and hyperlinks are enabled
- **THEN** the visible output is `o/r#3` and `o/r#7`, and they open their full URLs

#### Scenario: GitHub comment and review
- **WHEN** a document contains the bare URLs `https://github.com/o/r/issues/12#issuecomment-123`, `https://github.com/o/r/pull/12#discussion_r456`, `https://github.com/o/r/discussions/12#discussioncomment-789` and `https://github.com/o/r/pull/12#pullrequestreview-5` and hyperlinks are enabled
- **THEN** the first three are shown as `o/r#12 (comment)` and the last one as `o/r#12 (review)`

#### Scenario: GitHub commit
- **WHEN** a document contains the bare URLs `https://github.com/o/r/commit/0123456789abcdef0123456789abcdef01234567` and `https://github.com/o/r/pull/12/commits/abcdef1234567` and hyperlinks are enabled
- **THEN** they are shown as `o/r@0123456` and `o/r@abcdef1`

#### Scenario: Repository with a dot
- **WHEN** a document contains the bare URL `https://github.com/vercel/next.js/pull/12` and hyperlinks are enabled
- **THEN** it is shown as `vercel/next.js#12`

#### Scenario: Scheme and host case
- **WHEN** a document contains the bare URL `http://GitHub.com/o/r/issues/2` and hyperlinks are enabled
- **THEN** it is shown as `o/r#2` and opens `http://GitHub.com/o/r/issues/2`

#### Scenario: Autolinks that are not shortened
- **WHEN** a document contains the bare URLs `https://github.com/o/r`, `https://github.com/o/r/blob/main/README.md`, `https://github.com/o/r/pull/12/files`, `https://github.com/o/r/pull/12/`, `https://github.com/o/r/issues/12?x=1`, `https://github.com/o/r/issues/12#top`, `https://github.com/o/r/commit/abc12`, `https://www.github.com/o/r/issues/1` and `https://gitlab.com/o/r/-/issues/1` and hyperlinks are enabled
- **THEN** each of them is shown as its URL

#### Scenario: GitHub link with its own text
- **WHEN** a document contains `[PR](https://github.com/o/r/pull/1)` and `[https://github.com/o/r/pull/1](https://github.com/o/r/pull/1)` and hyperlinks are enabled
- **THEN** the visible output is `PR` and `https://github.com/o/r/pull/1`

#### Scenario: GitHub autolink in a table
- **WHEN** a table has the header `PR` and the cell `https://github.com/o/r/pull/12` in one column and hyperlinks are enabled
- **THEN** the cell shows `o/r#12` and the column is 6 columns wide

#### Scenario: GitHub autolink in a pipe
- **WHEN** a document contains the bare URL `https://github.com/ekalinin/termd/pull/19` and hyperlinks are disabled
- **THEN** the output contains `https://github.com/ekalinin/termd/pull/19` once and does not contain `ekalinin/termd#19`

#### Scenario: Link in a table in a pipe
- **WHEN** a table cell contains `[docs](https://example.com/very/long/path)`, a paragraph contains the same link, and hyperlinks are disabled
- **THEN** the cell shows `docs [1]`, the line `[1] https://example.com/very/long/path` follows the table, and the paragraph shows `docs (https://example.com/very/long/path)`

