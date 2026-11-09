# Tasks

## 1. Resolution against the base URL

- [x] 1.1 Add `TestResolveBase` to `internal/render/render_test.go` for the base `https://github.com/o/r/blob/main/`, one case per row of the design table: `docs/guide.md`, `../other/b.md`, `guide.md#setup` (fragment kept), `logo.png?raw=true` (query kept), `my notes.md` and `my%20notes.md` (both `my%20notes.md`), `файл.md` (`%D1%84%D0%B0%D0%B9%D0%BB.md`), `/docs/x.md` (`.../blob/main/docs/x.md`), `/a%2Fb` (`%2F` kept), `/` (the base), `//example.com/x` (`https://example.com/x`), and as written: `https://example.com`, `mailto:me@example.com`, `C:/x`, `#usage`, `?tab=1`, the empty destination and `%zz`; plus `docs/guide.md` with the base `https://raw.githubusercontent.com/o/r/main/README.md` (`.../main/docs/guide.md`), with `https://github.com/o/r/blob/main` (`.../blob/docs/guide.md`) and with `file:///home/me/proj/` (`file:///home/me/proj/docs/guide.md`); verify `go test ./internal/render` fails
- [x] 1.2 Add `resolveBase(dest, decoded string, base *url.URL) string` to `internal/render/link.go` with the rules of the design (as written on a parse error, a scheme, or no host and no path; one leading `/` removed from `Path` and `RawPath` when there is no host; `base.ResolveReference`); verify the tests of 1.1 pass

## 2. Rendering

- [x] 2.1 Add `TestBaseURL` to `internal/render/render_test.go`: with hyperlinks enabled and `BaseURL` `https://github.com/o/r/blob/main/`, `[guide](docs/guide.md)`, `![logo](img/logo.png?raw=true)`, `<img src="img/arch.png" alt="arch">` and `<a href="docs/a.md">a</a>` give OSC 8 links to `https://github.com/o/r/blob/main/...` with the visible text unchanged; `[empty]()` gives `empty` without an OSC 8 link; with `Dir` `/home/u/proj` set as well, `docs/guide.md` still resolves against the base URL and no `file:` URL appears; with hyperlinks disabled the output is `guide (docs/guide.md)`; verify `go test ./internal/render` fails
- [x] 2.2 Add `BaseURL *url.URL` to `render.Options` and make `destination` and `htmlDestination` call `resolveBase` before the `Dir` check when hyperlinks are enabled and `BaseURL` is set; verify the tests of 2.1 pass, `go test ./internal/render ./internal/site` passes and `git status testdata/golden` shows no changes

## 3. Command line

- [x] 3.1 Add `TestBaseURL` to `cmd/termd/main_test.go`: `[guide](docs/guide.md)` from stdin with `--hyperlinks=always --base-url https://github.com/o/r/blob/main/` gives `https://github.com/o/r/blob/main/docs/guide.md` in the OSC 8 link; a temporary file with `[a](a.md)` and `--base-url https://example.com/docs/` gives `https://example.com/docs/a.md` and no `file:` URL; stdin with `--base-url file:///home/me/proj/` gives `file:///home/me/proj/docs/guide.md`; the file with `--hyperlinks=never` and a valid base gives `guide (docs/guide.md)`; `--base-url github.com/o/r/blob/main/`, `localhost:8080/docs/`, `https:///docs/` and `--base-url=` each exit with status 2 and print a message that contains `--base-url` to stderr; verify `go test ./cmd/termd` fails
- [x] 3.2 In `run` add the flag `base-url` with the usage text "base URL for relative links and images, such as https://github.com/o/r/blob/main/", record with `fs.Visit` whether it was given, validate it as the design says with the message `--base-url must be an absolute URL, such as https://example.com/docs/, got %q`, and pass the parsed URL as `Options.BaseURL`; verify `go test ./cmd/termd` passes and `go run ./cmd/termd -h` lists `-base-url`

## 4. Documentation

- [x] 4.1 In `README.md` add a `--base-url` row to the flags table, and in the "Links" section describe the flag: with hyperlinks enabled, relative destinations of a document from a file or stdin are resolved against the URL instead of the directory of the file, `/docs/x.md` loses the leading `/`, `//host/x` gets the scheme of the base, the query and the fragment are kept, a directory base must end with `/` while the URL of the document works as is, and disabled hyperlinks keep the destination as written; update the sentences that say `/` destinations and stdin links are left as written so they hold only without `--base-url`; verify every statement matches `specs/markdown-rendering/spec.md` and `specs/cli/spec.md`

## 5. Verification

- [x] 5.1 Run `make check` and `openspec validate add-base-url-flag --strict`; verify both pass and `git status testdata/golden` shows no changes
- [x] 5.2 Run `make build`, then `./termd --no-pager --hyperlinks=always --base-url https://github.com/ekalinin/termd/blob/main/ README.md | cat -v` and the same with `README.md` piped through stdin; verify relative links such as `docs/...` become `https://github.com/ekalinin/termd/blob/main/...` in both runs and `https://` links are unchanged; run `./termd --base-url example.com README.md` and verify exit status 2 and the `--base-url` message
