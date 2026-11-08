# Design

## Context

`internal/render/link.go` resolves destinations in two places: `destination`, called by the `ast.Link` and `ast.Image` cases, and `htmlDestination`, called for `href` and `src` of HTML tags. Both return the destination as written unless hyperlinks are enabled and `Options.Dir` is set, and then call `resolveDecoded(dest, decoded, dir, host)`, which builds a `file:` URL. `destination` decodes backslash escapes and character references with `unescape` first; the HTML tokenizer has already decoded attribute values.

`run` in `cmd/termd/main.go` validates every flag before it reads the input and reports a bad value with `usageError`, which prints `termd: <message>` and the usage text and returns 2. `Options.Dir` comes from `docDir(file)` and is empty for stdin.

Autolinks always have a scheme, so they are never resolved, and the GitHub short text does not change. Without hyperlinks nothing is resolved, so plain mode and the reference markers of table cells do not change.

The requirements are in `specs/cli/spec.md` and `specs/markdown-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- Without `--base-url`, the output is byte for byte the same as before: files, stdin, the site and the golden files.
- The resolution is a pure function of the destination and the base URL, tested without a file system.

**Non-Goals:**
- Normalizing the base URL, for example adding a trailing `/`.
- Checking that the targets exist.

## Decisions

### `Options.BaseURL *url.URL`, parsed once in `run`

`render.Options` gets a field `BaseURL *url.URL`; `nil` means no base URL. `run` parses and validates the flag value, so the renderer gets a URL that is known to be good and does not parse it again for every link. `internal/site` and the golden tests leave it `nil`.

Alternative: a `string` field parsed by the renderer. Rejected: the renderer would have to handle a parse error that `run` has already reported, and it would parse the same value for every link.

### The base URL wins inside the renderer

`destination` and `htmlDestination` check `BaseURL` before `Dir`: with hyperlinks enabled and `BaseURL` set, they call the new `resolveBase(dest, decoded string, base *url.URL) string`; otherwise they behave as today. `run` still sets `Dir` from the file. Keeping the precedence in the renderer means that any caller that sets both gets the same result as the command line.

Alternative: `run` clears `Dir` when `--base-url` is given. Rejected: the precedence would then be a rule of the command line only, and the renderer would silently prefer `Dir` for other callers.

### Which destinations are resolved against the base URL

`resolveBase` parses the decoded destination with `net/url` and returns the destination as written when parsing fails, when it has a scheme, and when it has neither a host nor a path. Otherwise it removes one leading `/` from the path of a destination without a host, as glamour does, and returns `base.ResolveReference(ref).String()`:

| Destination | Example | Result with the base `https://github.com/o/r/blob/main/` |
|---|---|---|
| empty | `[x]()` | as written: no hyperlink |
| not a valid URL reference | `%zz` | as written |
| has a scheme | `https://example.com`, `mailto:me@example.com`, `C:/x` | as written |
| no host and no path | `#usage`, `?tab=1` | as written |
| a relative path | `docs/guide.md`, `../b.md`, `guide.md#setup` | `.../blob/main/docs/guide.md`, `.../blob/b.md`, `.../blob/main/guide.md#setup` |
| starts with a single `/` | `/docs/x.md` | `.../blob/main/docs/x.md` |
| only `/` | `/` | the base URL itself |
| starts with `//` | `//example.com/x` | `https://example.com/x` |

The leading `/` is removed from both `Path` and `RawPath`: `RawPath` is either empty or the escaped form of `Path`, so it stays a valid encoding, and `/a%2Fb` keeps its `%2F`.

`ResolveReference` follows RFC 3986: a base that does not end with `/` loses its last segment. This is what makes the URL of the document itself a good base, `curl -s URL | termd --base-url URL`, and it is how a browser treats `<base href>`. A directory given as the base must end with `/`; the README says so.

Alternative: add a missing trailing `/` to the base. Rejected: the URL of the document would then resolve `docs/guide.md` to `.../README.md/docs/guide.md`.

### Decoding and encoding

The input to `resolveBase` is the same as for `resolveDecoded`: `destination` passes the destination as written and its `unescape`d form, and `htmlDestination` passes the decoded attribute value twice. `url.Parse` decodes the percent-encoded characters of the path, and `url.URL.String` encodes the path and the fragment again, so `<my notes.md>` and `my%20notes.md` both become `my%20notes.md`. Unlike a `file:` URL, the query is kept, since it means something to a web server (GitHub's `?raw=true`); `url.URL.String` writes `RawQuery` as it was parsed.

### Validation of `--base-url`

`run` gets a flag `base-url` with an empty default and records with `fs.Visit` whether it was given, as it does for `--width` and `--theme`. When it was given, the value is parsed with `url.Parse`, which lowercases the scheme, and accepted when it has a scheme and a host, or when the scheme is `file` and the path starts with `/`. Anything else, including an empty value, `github.com/o/r` (no scheme), `localhost:8080/docs/` (the scheme `localhost` with an opaque part) and `https:///docs/` (no host), gives:

```
termd: --base-url must be an absolute URL, such as https://example.com/docs/, got "github.com/o/r"
```

The value is checked also when hyperlinks are disabled, so a script fails the same way in a terminal and in a pipe.

### Tests and golden files

`resolveBase` is tested with a table for the base `https://github.com/o/r/blob/main/`, plus the base of a document, a base without a trailing `/` and a `file:` base. `Render` is tested with `BaseURL` set for links, images, HTML `img` and `a`, with and without `Dir`, and with hyperlinks disabled. `run` is tested with stdin, a file, a `file:` base, `--hyperlinks=never` and the invalid values of the spec.

No golden fixture is added: the golden tests render without `BaseURL`, so the existing golden files do not change.

## Risks / Trade-offs

- [A base that names a directory without a trailing `/` resolves one level too high] → RFC 3986 behavior, which also makes the URL of the document work as a base; documented in the README with both forms.
- [`/docs/x.md` is right only when the base is the root of the repository or a file in it] → Matches glamour and GitHub for the common case of a README at the root of the repository; a document in a subdirectory still gets a clickable URL, the same one glamour gives.
- [A space or a non-ASCII character in the query is written into the hyperlink as is] → The same as for absolute URLs written in the document today; a markdown destination can have a space only inside `<` and `>`.
- [`//host/x` with a `file:` base becomes `file://host/x`] → The RFC result; such links are rare in documents read with a `file:` base.

## Migration Plan

None. Without the flag the output does not change. Rollback is a revert of the change.
