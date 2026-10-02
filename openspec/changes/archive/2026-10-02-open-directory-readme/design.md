# Design

## Context

`run` in `cmd/termd/main.go` picks the input in one `switch`: a single argument other than `-` is read with `env.readFile`, otherwise stdin is read. Any read error is printed as `termd: <error>` with exit status 1; for a directory, `os.ReadFile` fails with `read docs/: is a directory`. The `env` struct holds everything `run` takes from the outside world, and `cmd/termd/main_test.go` builds it with a `fake` that uses real temp files through `os.ReadFile`.

The requirements are in `specs/cli/spec.md`.

## Goals / Non-Goals

**Goals:**
- A file argument, `-` and stdin behave exactly as before, including the text of every error.
- The README path that is found is the path `run` reads, held in one variable, so the parallel change that resolves relative links can take its directory after both are merged.
- The README rendering goes through the same code as a file argument, so width, styles and paging cannot differ.

**Non-Goals:**
- Resolving relative links against the README directory (the resolve-relative-links change).
- Any change in `internal/render`; no golden file changes.

## Decisions

### Resolve the path before reading it

A new function `inputPath(arg, stat, readDir)` returns the file to read: the argument itself, or the README when the argument is a directory. In `run` the `fs.NArg() == 1` case calls it and reads the result with `e.readFile`. The result is stored in `file`, declared next to `src` and `err` and empty for stdin, so the path that was read is available after the `switch`.

Alternative: read the argument first and look for a README when the read fails with "is a directory". Rejected: the error is `EISDIR` on Unix but a different error on Windows, and matching on the error text is fragile.

### `env` gets `stat` and `readDir`

`env` gets two fields, `stat func(string) (os.FileInfo, error)` and `readDir func(string) ([]os.DirEntry, error)`, set to `os.Stat` and `os.ReadDir` in `systemEnv` and in the test fake. They take OS paths, like `readFile`.

Alternative: one `fs.FS` from `os.DirFS`. Rejected: an `fs.FS` takes unrooted slash paths, while the argument can be absolute, relative with `..`, or end with a separator; `readFile` would also have to move to it.

### Directory detection follows symbolic links

The argument is a directory when `stat` (which follows symbolic links) reports a directory, so a symbolic link to a directory is treated as the directory. When `stat` fails or reports anything else, the argument goes to `readFile` unchanged, so a missing or unreadable file gives the same error as before (`open missing.md: no such file or directory`).

### Finding the README

`readDir` lists the directory once. For each name of `README.md`, `README.markdown`, `README`, in this order, the entries are compared with `strings.EqualFold`, and the first matching entry that is a regular file wins. `os.ReadDir` returns the entries sorted by name, so names that differ only in case are tried in byte order: `README.md` before `Readme.md` before `readme.md`. That case only exists on a case-sensitive file system; on the default file systems of macOS and Windows a directory cannot hold two of them.

Alternative: `stat` each candidate path directly. Rejected: on a case-sensitive file system it misses `readme.md` unless every case variant is tried, and on a case-insensitive one it returns the name as typed instead of the name as stored.

### Which entries count

A matching entry counts only when `stat` of its path reports a regular file. `stat` follows symbolic links, so `docs/README.md` linking to `../README.md` counts. These entries are skipped and the search goes on:

- a directory named `README.md`: it is not a document, and failing on it would hide a `README` file next to it;
- a broken symbolic link: there is nothing to render;
- a named pipe or a device: reading it can block or never end.

Alternative: the type from the directory listing (`DirEntry.Type`). Rejected: it does not follow symbolic links, so a README that is a link to a file would need a second check anyway.

A regular file that cannot be read, for example without read permission, is selected and its read fails with an error naming it. termd does not fall back to the next candidate: `stat` cannot tell whether a file is readable, and showing another file would hide the problem.

### The README path keeps the argument

The path is `filepath.Join(arg, name)`, with the argument as given and the name as stored in the directory: `termd docs/` reads `docs/README.md`, and `termd link` reads `link/README.md`, not the target of the link. The operating system follows the links when the file is read, and the path is the same as when the user types it, so `termd docs/` and `termd docs/README.md` behave the same.

### Error for a directory without a README

`inputPath` returns `fmt.Errorf("no README in %s", arg)`, and the existing error path of `run` prints it as `termd: no README in docs/` with exit status 1. The directory is shown as given, with its trailing separator, so the message matches what the user typed. A directory that cannot be listed fails with the error from `os.ReadDir`, which names the directory.

### Usage text

The usage line becomes `Usage: termd [flags] [FILE | DIR]`, and the description says that for a directory its `README.md`, `README.markdown` or `README` is rendered. The error for extra arguments (`expected at most one file`) stays as it is.

### Tests

The tests use real temporary directories, like `writeFile` does for files, through a helper that creates files, directories and symbolic links. The case-variant scenario cannot be created on a case-insensitive file system, which is the default on macOS, so its test fakes `stat`, `readDir` and `readFile` with a `testing/fstest.MapFS`. The symbolic link tests skip when `os.Symlink` fails, as it can on Windows without the privilege.

## Risks / Trade-offs

- [A README that is a symbolic link to a file in another directory is read as `docs/README.md`, so its relative links would resolve against `docs/`, not against the directory of the target] → The same holds for `termd docs/README.md` today; the resolve-relative-links change decides how to treat links.
- [Byte order puts `README.MD` before `README.md` when both exist] → Only possible on a case-sensitive file system with both names; accepted for a rule that is simple to state.
- [The whole directory is listed to find one file] → A single `os.ReadDir` call, small next to parsing and rendering the document.

## Migration Plan

None. File arguments and stdin work as before; only a directory argument, an error today, changes its behavior. Rollback is a revert of the change.
