# Design

## Context

`render.Parse` uses `extension.GFM`, so a bare URL is found by the linkify parser and a URL in `<` and `>` by the CommonMark parser. A spike with goldmark v1.8.6 showed that both give an `ast.AutoLink` of the type `ast.AutoLinkURL`, also in a table cell, and that `URL()` equals `Label()` for them. Linkify leaves a trailing `.`, `,`, `!` or `)` out of the URL. The exception is `www.github.com/...`, whose `URL()` gets the prefix `http://`. A link with its own text, also `[url](url)`, is an `ast.Link`.

`renderer.inline` renders an `ast.AutoLink` as one span with the label in the link style and passes it to `text.LinkSpans` with the URL. With hyperlinks, `LinkSpans` sets the URL as the hyperlink of the label. Without hyperlinks, it writes the label, and the URL after it only when it differs from the label. Table cells are measured by the visible text of their spans, so a shorter label makes a column narrower without a change in `internal/table`.

glamour (`internal/autolink`) has seven anchored regular expressions for the same forms and uses them only for links in tables. They accept the host `github.com` in lower case only, `issue`, `pulls` and `discussion` next to the plural forms, no `.` in repository names, `[A-z0-9]` for names and SHAs, which also matches `[`, `\`, `]`, `^` and `` ` ``, and any fragment after a commit.

In 3924 READMEs of the Go module cache, there are 4901 links to issues, pull requests, discussions and commits on GitHub. 2816 of them have their own text, and 2085 are bare URLs or URLs in `<` and `>`. 1977 of the latter end with `/pull/N` and 97 with `/issues/N`. None of them starts with `http://`, `www.` or a host with upper-case letters. The README of `github.com/klauspost/compress@v1.20.0` alone has 186 of the latter, in the nested lists of its changelog.

The requirements are in `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- Exactly the forms of the spec are shortened, and the hyperlink always opens the URL as written.
- Documents without GitHub autolinks render byte for byte as before, and plain output never changes. No fixture and no example of the site contains a GitHub autolink, so the existing golden files do not change.

**Non-Goals:**
- `#N` without `owner/repo` for links to the repository of the document, as GitHub shows them. termd does not know the repository of a document.
- `www.github.com`, `/pull/N/files` and other forms of GitHub URLs. They are shown as their URL, as today, and none of them is in the sample.

## Decisions

### Anchored regular expressions in `internal/render/github.go`

`internal/render/github.go` holds a list of patterns, each an anchored regular expression and a function that builds the text from the groups, and `githubText(url string) (string, bool)`, which returns the text of the first pattern that matches. Every pattern starts with `(?i:https?://github\.com)/([A-Za-z0-9_-]+)/([A-Za-z0-9_.-]+)`, so only the scheme and the host are case-insensitive, and ends with `$`. The endings are:

- `/(?:issues|pull|discussions)/([0-9]+)$` → `owner/repo#N`
- `/(?:issues|pull|discussions)/([0-9]+)#(?:issuecomment-|discussion_r|discussioncomment-)[0-9]+$` → `owner/repo#N (comment)`
- `/pull/([0-9]+)#pullrequestreview-[0-9]+$` → `owner/repo#N (review)`
- `/(?:commit|pull/[0-9]+/commits)/([0-9a-fA-F]{7,40})$` → `owner/repo@` and the first 7 characters of the SHA

The patterns are compiled once, at package initialization.

Alternatives:
- The package of glamour. Rejected: it is internal to glamour and differs from the spec in the points of Context.
- `net/url.Parse` and a split of the path. Rejected: it decodes the path and accepts forms the spec does not name, such as a port or user information, and it needs more code for the same rules. The regular expressions state the forms the way the spec does.

The comment fragments are accepted after all three kinds, as the proposal lists them. GitHub gives `#discussion_r` only to pull requests and `#discussioncomment-` only to discussions, so a stricter rule would only reject URLs that do not occur.

A commit URL with a fragment, such as `#diff-...`, is not shortened, unlike in glamour. Such a fragment points to a file or a line of the commit, and the text `owner/repo@SHA` would hide it.

### Shortening in the `ast.AutoLink` case

In `renderer.inline`, the `ast.AutoLink` case replaces the label with `githubText(url)` when hyperlinks are enabled, the type is `ast.AutoLinkURL` and the URL matches. The URL that goes to `text.LinkSpans` does not change. Since `URL()` equals `Label()` for every URL that can match, it does not matter which of the two is matched.

Alternatives:
- In `text.LinkSpans`. Rejected: it does not know whether a label is an autolink, so `[url](url)` would be shortened too.
- Also without hyperlinks, as `o/r#12 (url)`. Rejected by the proposal: the short text would add a second copy of the URL.

### Tests and fixture

`TestGitHubLinks` in `internal/render/render_test.go` renders each scenario with `Options{Width: 80, Style: style.Options{Hyperlinks: true}}` and compares the whole output, so it checks the text and the hyperlink at once. A new fixture `testdata/autolinks.md` adds six golden files. The existing fixtures do not change, so their golden files stay as they are.

## Risks / Trade-offs

- [The short text hides the URL in a terminal] → The hyperlink opens it, and `--hyperlinks=never` and a pipe show the full URL, as today.
- [A terminal ignores OSC 8 although hyperlinks are enabled, so `o/r#12` cannot be opened] → The same holds today for every link with its own text. `--hyperlinks=never` shows the URL.
- [GitHub adds URL forms or changes the text it shows for them] → Each form is one pattern, so a form can be added without changes elsewhere.

## Migration Plan

None. Plain output and documents without GitHub autolinks do not change, so the existing golden files stay as they are. Rollback is a revert of the change.
