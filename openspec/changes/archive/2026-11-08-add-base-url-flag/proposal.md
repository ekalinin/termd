# Proposal

## Why

termd resolves relative links and images only against the directory of the document file. A document read from stdin, for example `curl -s https://raw.githubusercontent.com/o/r/main/README.md | termd`, keeps its relative links as written, and a terminal cannot open them. The "resolve-relative-links" change left a flag for this out of scope. glamour has the `WithBaseURL` option.

## What Changes

- A new flag `--base-url URL`. With it and with hyperlinks enabled, the relative destinations of links and images are resolved against `URL` and written into the hyperlink as absolute URLs, for a document from stdin and from a file.
- The flag wins over the directory of the file: with `--base-url`, no `file:` URLs are made.
- Destinations that start with `/` are resolved as glamour does: the leading `/` is removed and the rest is resolved against the base, so `/docs/x.md` with the base `https://github.com/o/r/blob/main/` becomes `https://github.com/o/r/blob/main/docs/x.md`. This matches GitHub when the base is the root of the repository. A destination that starts with `//` gets the scheme of the base: `//example.com/x` becomes `https://example.com/x`.
- Destinations with a scheme and fragment-only destinations do not change, as today.
- When hyperlinks are disabled, destinations are shown as written, as today.
- A value that is not an absolute URL with a scheme and a host, or a `file:` URL with an absolute path, is a usage error with exit status 2, and the message names `--base-url`. So `--base-url file:///home/me/proj/` is valid.

Out of scope:

- An environment variable or a configuration file for the base URL.
- Fetching a document from a URL.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `cli`: a new requirement for the `--base-url` flag, its validation and its usage error.
- `markdown-rendering`: the "Hyperlinks" requirement resolves relative destinations against the base URL when it is given.

## Impact

- `cmd/termd/main.go`: the flag, its validation and the usage text.
- `internal/render`: `Options` gets the base URL; `destination` and `htmlDestination` resolve against it before the directory of the file.
- Tests for stdin, a file, a destination with a scheme, a fragment, a leading `/`, disabled hyperlinks and an invalid value.
- `README.md`: the flags table and the "Links" section.
