# Proposal

## Why

`termd docs/` fails with `termd: read docs/: is a directory`. A repository or a documentation folder is often opened by its directory, and glow renders the README of a directory given as its argument.

## What Changes

- When the argument is a directory, termd renders the README in that directory: `README.md`, `README.markdown` or `README`, with the names compared case-insensitively. When more than one exists, the first one in this order is used.
- Only the directory itself is searched. glow walks the whole tree and can pick `Docs/README.md` before `./README.md`; termd does not look into subdirectories.
- When the directory has no README, termd prints `termd: no README in <dir>` and exits with status 1, as for a file that cannot be read.
- Everything else (width, styles, paging) works as when the README path is given directly.

Out of scope:

- `termd` without arguments opening `./README.md`; it keeps printing the usage.
- Listing or choosing between other markdown files of the directory.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `cli`: the input requirement accepts a directory and renders its README.

## Impact

- `cmd/termd/main.go`: the input selection (`run`, the `fs.NArg() == 1` case) checks whether the argument is a directory; `env` gets a way to stat and list it, so tests can fake it.
- `cmd/termd/main_test.go`: a directory with a README, with several candidates, with a lowercase name, and without a README.
- `README.md`: the Usage examples.
