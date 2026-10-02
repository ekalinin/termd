# Tasks

## 1. Link resolution

- [x] 1.1 Add `TestResolveLink` to `internal/render/render_test.go` for the directory `/home/u/proj` and the host `box`, one case per row of the design table: `docs/guide.md`, `./a.md`, `../other/b.md`, `my notes.md` and `my%20notes.md` (both `my%20notes.md`), `файл.md` (`%D1%84%D0%B0%D0%B9%D0%BB.md`), `my\_notes.md` (`my_notes.md`), `guide.md#setup` (fragment kept), `logo.png?raw=true` (query dropped), and as written: `https://example.com`, `mailto:me@example.com`, `C:/x`, `#usage`, `?tab=1`, `/docs/x.md`, `//example.com/x`, the empty destination and `%zz`; plus `docs/guide.md` with an empty directory (as written) and with an empty host (`file:///home/u/proj/docs/guide.md`); add `TestFileURL` checking that `fileURL("", "C:/proj/docs/guide.md", "")` is `file:///C:/proj/docs/guide.md`; verify `go test ./internal/render` fails
- [x] 1.2 Create `internal/render/link.go` with `resolveLink`, `fileURL` and the host lookup (`os.Hostname` once via `sync.OnceValue`, empty on Windows and on error); verify the tests of 1.1 pass

## 2. Rendering

- [x] 2.1 Add `TestRelativeLinks` to `internal/render/render_test.go`: `[guide](docs/guide.md)` and `![arch](img/arch.png)` rendered with `Dir` `/home/u/proj` and hyperlinks enabled give the OSC 8 links `file://<os.Hostname()>/home/u/proj/docs/guide.md` and `.../img/arch.png` with the visible text `guide` and `[image: arch]`; `[empty]()` gives `empty` without an OSC 8 link; the same source with `Dir` and hyperlinks disabled gives `guide (docs/guide.md)` and `[image: arch] (img/arch.png)`; without `Dir` and with hyperlinks enabled the OSC 8 link is `docs/guide.md`; verify `go test ./internal/render` fails
- [x] 2.2 Add `Dir` to `render.Options` and resolve the destinations of `ast.Link` and `ast.Image` through it when hyperlinks are enabled; verify the tests of 2.1 pass, `go test ./internal/render ./internal/site` passes and `git status testdata/golden` shows no changes

## 3. Command line

- [x] 3.1 Add `TestRelativeLinks` to `cmd/termd/main_test.go`: a temporary `doc.md` with `[guide](docs/guide.md)` run with `--hyperlinks=always` by absolute path gives `file://<host><dir>/docs/guide.md` in the OSC 8 link; after `t.Chdir` to the parent of the temporary directory, the same file given by a relative path gives the same link; the same document from stdin and with `-` gives `docs/guide.md`; the file with `--hyperlinks=never` gives `guide (docs/guide.md)`; verify `go test ./cmd/termd` fails
- [x] 3.2 In `run` keep the path passed to `env.readFile` and set `Options.Dir` to its absolute directory (empty for stdin and when `filepath.Abs` fails); verify `go test ./cmd/termd` passes

## 4. Documentation

- [x] 4.1 In the "Links" section of `README.md` describe the resolution: a relative link or image of a document read from a file opens as a `file://` URL with the host name, built from the directory of the file; absolute URLs, `#` and `/` destinations, stdin and disabled hyperlinks keep the destination as written; verify every statement matches `specs/markdown-rendering/spec.md`

## 5. Verification

- [x] 5.1 Run `make check` and `openspec validate resolve-relative-links --strict`; verify both pass and `git status testdata/golden` shows no changes
- [x] 5.2 Run `make build`, then `./termd --no-pager --hyperlinks=always testdata/inline.md | cat -v` from the repository root and from `testdata` with the relative path `inline.md`, and the same document through stdin; verify the image link is `file://<host><repo>/testdata/docs/arch.png` in both file runs, the `https://` links are unchanged, and the stdin run keeps `docs/arch.png`
