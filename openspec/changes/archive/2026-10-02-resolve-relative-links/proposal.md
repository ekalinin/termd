# Proposal

## Why

A link `[guide](docs/guide.md)` or an image `![arch](img/arch.png)` is written into the OSC 8 hyperlink with its relative destination as is. OSC 8 expects absolute URIs, so a terminal cannot open these links, although they are the most common links of a README.

## What Changes

- When the document is read from a file, the relative destinations of links and images are resolved against the directory of that file and written into the OSC 8 hyperlink as absolute `file://` URLs. A click opens the target with the default application of the system, for example an image in the image viewer.
- Absolute URLs (`https:`, `mailto:` and other schemes) and fragment-only links (`#section`) do not change.
- A document read from stdin has no directory to resolve against; its links do not change.
- When hyperlinks are disabled, a link is still written as `text (url)` with the destination as written in the document, because a path relative to the document is easier to read than an absolute one.

Out of scope:

- A flag that sets the base for stdin, for example `curl ... | termd --base <url>`.
- Opening a linked markdown file in termd.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `markdown-rendering`: the hyperlinks requirement resolves relative destinations against the directory of the document file.

## Impact

- `internal/render`: `Options` gets the directory of the document; links and images (`ast.Link`, `ast.Image`) resolve their destination with it when hyperlinks are enabled.
- `cmd/termd/main.go`: passes the directory of the input file. The site generator passes none, so its output does not change.
- The design decides between `file:///path` and `file://hostname/path`: the OSC 8 spec recommends the host name, so that a terminal can tell that a file lives on a remote machine in an SSH session.
- Tests for relative, absolute, fragment-only and stdin cases; `README.md`, the "Links" section.
