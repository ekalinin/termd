# Proposal

## Why

A bare URL of a GitHub issue, pull request or commit is shown in full, for example `https://github.com/ekalinin/termd/pull/19`. With hyperlinks enabled, the URL is already in the hyperlink, so the long text only takes space, and in a table it makes the column wide. glamour shortens such URLs to the form GitHub uses itself, `ekalinin/termd#19`, but only for links in tables.

## What Changes

- When hyperlinks are enabled, the visible text of an autolink (a bare URL or `<url>`) to GitHub is shortened, and the hyperlink still opens the full URL:

  | URL | Text |
  |---|---|
  | `https://github.com/o/r/issues/12`, `/pull/12`, `/discussions/12` | `o/r#12` |
  | the same with `#issuecomment-...`, `#discussion_r...`, `#discussioncomment-...` | `o/r#12 (comment)` |
  | `https://github.com/o/r/pull/12#pullrequestreview-...` | `o/r#12 (review)` |
  | `https://github.com/o/r/commit/<sha>`, `/pull/12/commits/<sha>` | `o/r@<first 7 characters of sha>` |

- This applies everywhere, including table cells, where it makes columns narrower.
- Repository names with dots, such as `vercel/next.js`, are shortened too.
- When hyperlinks are disabled, the URL is shown in full, as today, because the shortened text would hide it.
- Links with their own text (`[text](url)`) and URLs of other hosts do not change.

Out of scope:

- Other hosts, such as GitLab or Codeberg.
- Links to repositories, files and other GitHub pages.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `markdown-rendering`: the "Hyperlinks" requirement shortens the visible text of GitHub autolinks when hyperlinks are enabled.

## Impact

- `internal/render`: the label of `ast.AutoLink` goes through the shortening when hyperlinks are enabled.
- A small matcher of GitHub URLs, in `internal/render` or its own package.
- Tests for every pattern, for an unknown GitHub URL, for a URL with its own text, and for disabled hyperlinks.
- `README.md`: the "Links" section.
