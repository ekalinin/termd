# Design

## Context

`render.Render` turns `ast.Link` and `ast.Image` into spans with `text.LinkSpans(label, string(n.Destination), hyperlinks)`. With hyperlinks enabled the destination becomes the `Link` of the spans and `style.Options.LinkOpen` writes it into the OSC 8 sequence; without hyperlinks it is appended as ` (url)`. goldmark gives the destination as written in the source: backslash escapes and character references are not resolved, and the angle brackets of `<my notes.md>` are removed.

`render.Render` does not know where the document comes from. `run` in `cmd/termd/main.go` reads the document either with `env.readFile(path)` or from stdin. The site generator (`internal/site`) and the golden tests call `Render` with their own `Options`.

The requirements are in `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- Rendering without a directory is byte for byte the same as before: stdin, the site, the golden files.
- The resolution is a pure function of the destination, the directory and the host name, so it is tested without a file system and without the real host name.

**Non-Goals:**
- Checking that the target of a link exists.
- Links in raw HTML (`<a href>`, `<img src>`), which are shown as source text.
- Autolinks: a CommonMark autolink always has a scheme, and a GFM `www.` autolink gets `http://`, so they are never relative.

## Decisions

### `Options.Dir`, set by `run` only

`render.Options` gets a `Dir string` field: the absolute directory of the document file, or `""` when there is none. The renderer resolves destinations only when `Dir` is set and hyperlinks are enabled. `internal/site` and the golden tests leave it empty, so their output does not change.

`run` keeps the path it passes to `env.readFile` in a variable and sets `Dir` from it with `filepath.Abs` and `filepath.Dir`. Taking the path at that point means that after the parallel change that lets `termd DIR` read `DIR/README.md` the base is the directory of the file actually read, without further changes here. The path is not passed through `filepath.EvalSymlinks`: a document opened through a symlink resolves its links next to the symlink, the way the user named it, and the operating system follows the symlinks of the target when it opens it. When `filepath.Abs` fails (the working directory is gone), `Dir` stays empty and links stay as written.

Alternative: pass the document path instead of the directory. Rejected: the renderer needs only the directory, and the proposal names the directory.

### Host form: `file://HOSTNAME/path`

The URL carries the host name of the machine, as returned by `os.Hostname`, for example `file://box/home/u/proj/docs/guide.md`.

- The OSC 8 specification asks for the host name in `file:` URLs, so that a terminal can tell that a file lives on another machine, which happens whenever termd runs in an SSH session. kitty, for example, offers to fetch such a file over SSH instead of opening a local file with the same path.
- The tools that already write `file:` hyperlinks do the same: GNU coreutils `ls --hyperlink` writes `file://HOSTNAME/absolute/path`, and ripgrep's default hyperlink format is `file://{host}{path}`. `os.Hostname` returns the same name as `gethostname`, which they use, so a terminal compares termd's links the same way.
- macOS ignores the host when it turns a `file:` URL into a path: `NSURL` parses `file://<this host>/tmp/a%20b.png` as a file URL with the path `/tmp/a b.png` (checked on this machine), so terminals that hand the URL to the system open the local file.

The host is looked up once per process, with `sync.OnceValue`, and only when a document with a directory has a link and hyperlinks are enabled.

When `os.Hostname` fails, the host is empty (`file:///home/u/proj/docs/guide.md`). On Windows the host is always empty (`file:///C:/proj/docs/guide.md`): there the host of a `file:` URL names a network server, `file://server/share/x` is `\\server\share\x`, so a host name would point to a share instead of the local disk. ripgrep makes the same split: its default format is `file://{host}{path}` everywhere except Windows, where it is `file://{path}`.

Alternatives:
- `file:///path` everywhere: accepted by every terminal, but in an SSH session a click opens a local file with the same path, or fails, with no way for the terminal to know better.
- `file://localhost/path`: the same problem as the empty host.

### Which destinations are resolved

A destination is resolved when all of the following hold; otherwise it is written as in the document:

| Destination | Example | Result |
|---|---|---|
| empty | `[x]()` | as written: no hyperlink, as today |
| starts with `/` | `/docs/x.md`, `//example.com/x` | as written |
| not a valid URL reference after unescaping | `%zz`, `1:x` | as written |
| has a scheme | `https://example.com`, `mailto:me@example.com`, `C:/x` | as written |
| no path (fragment or query only) | `#usage`, `?tab=1` | as written |
| a relative path | `docs/guide.md`, `./a.md`, `../b.md`, `guide.md#setup` | resolved |

A root-relative destination such as `/docs/x.md` stays as written. GitHub reads it as relative to the repository root, a file system as an absolute path, and the two disagree for nearly every README that uses it: resolving it against `/` would produce a confident link to a file that does not exist, and finding the repository root would mean walking the file system for a `.git` directory, which is a guess outside a repository and is not "resolving against the directory of the file". Both readings agree that it is not relative to the directory of the document, so it is left alone. A protocol-relative destination `//host/x` names a host, not a file next to the document; resolving it against `file:` would give `file://host/x`, which is almost never meant.

A Windows drive path written as a destination, `C:/docs/x.md`, parses as a URL with the scheme `c` and stays as written. It is an absolute path, not a relative one, and it does not work on GitHub either.

### Decoding and encoding

The destination is first unescaped with the existing `unescape` (backslash escapes and character references, as the goldmark HTML renderer does), then parsed with `net/url`, which decodes the percent-encoded characters of the path. The decoded path is converted with `filepath.FromSlash` and joined to `Dir` with `filepath.Join`, which also resolves `.` and `..`. The absolute path is converted back with `filepath.ToSlash`, gets a leading `/` when it starts with a drive letter, and is written as a `url.URL` with the scheme `file`, the host and the fragment. `url.URL.String` percent-encodes spaces, non-ASCII bytes and every other character that a path does not allow, so `<my notes.md>` and `my%20notes.md` both become `my%20notes.md` and nothing is encoded twice. The result contains only printable ASCII, as OSC 8 requires.

`filepath.Join` drops a trailing slash, so a link to `docs/` becomes `file://box/home/u/proj/docs`; a directory opens the same way with or without it.

### The query is dropped, the fragment is kept

`guide.md#setup` becomes `file://box/home/u/proj/guide.md#setup`, and `logo.png?raw=true` becomes `file://box/home/u/proj/logo.png`.

- A query is a request parameter for a web server and means nothing for a local file. The common case is GitHub's `?raw=true` on images in a README. An opener that took the URL literally would look for a file named `logo.png?raw=true`; dropping the query cannot break anything.
- A fragment names a place inside the target, which a browser honors when it opens an HTML file. The openers that turn the URL into a path ignore it: GLib's `g_filename_from_uri`, used by GTK terminals, cuts the URI at `?` and at `#`, and `NSURL` keeps the fragment out of the path.

Alternatives: keep both, as RFC 3986 reference resolution does (rejected for the query, see above), or drop both (rejected: it loses the anchor for no gain).

### Plain mode and stdin keep the destination as written

The resolution happens only when `Style.Hyperlinks` is set. In plain mode the destination is visible text in `text (url)`, and a path relative to the document reads better than an absolute one; the "text equals URL" check of `text.LinkSpans` keeps comparing with the destination as written. A document from stdin has no `Dir`.

### Code layout

The resolution lives in a new file `internal/render/link.go`: `resolveLink(dest, dir, host string) string` with the rules above, `fileURL(host, path, fragment string) string` that builds the URL from an absolute path, the host lookup, and the renderer method `destination` that the `ast.Link` and `ast.Image` cases call instead of `string(n.Destination)`; it returns the destination as written unless hyperlinks are enabled and `Dir` is set. `text.LinkSpans` and `style.Options.LinkOpen` do not change.

### Tests and golden files

`resolveLink` and `fileURL` are tested with a fixed directory and the host `box`, including the Windows drive path, which `fileURL` handles in its slash form on any platform. `Render` is tested with `Dir` set and the real host name. `run` is tested with a temporary file given by absolute and by relative path, with stdin and with `--hyperlinks=never`.

No golden fixture is added: a resolved URL contains the absolute path of the checkout and the host name, which differ between machines. The golden tests render without `Dir`, so the existing golden files, including `inline.md` with its image `docs/arch.png`, do not change.

## Risks / Trade-offs

- [A terminal that does not accept a host name in `file:` URLs cannot open the links] → The same form is written by `ls --hyperlink` and ripgrep, and the OSC 8 specification asks for it; accepted.
- [A link to a file that does not exist gives a hyperlink that opens nothing] → The same as a broken link in a browser; checking every target would make rendering depend on the file system.
- [Root-relative links such as `/docs/x.md` are still not clickable] → Documented in the README.
- [The host name and the absolute path of the document appear in the output, also when `--hyperlinks=always` writes into a file] → They are only inside the OSC 8 sequences, which are off in plain mode by default.
- [The Windows behavior is not run on Windows] → `GOOS=windows go vet` builds the code, and the drive path is covered by the `fileURL` test; Windows binaries are best-effort, as the README says.

## Migration Plan

None. Documents from stdin, the site and the golden files render as before. Rollback is a revert of the change.
