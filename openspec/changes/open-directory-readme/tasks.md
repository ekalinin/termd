# Tasks

## 1. Environment

- [x] 1.1 In `cmd/termd/main.go` add the fields `stat func(string) (os.FileInfo, error)` and `readDir func(string) ([]os.DirEntry, error)` to `env`, set to `os.Stat` and `os.ReadDir` in `systemEnv`, and set the same functions in the `fake` of `cmd/termd/main_test.go`; verify `go vet ./cmd/termd` and `go test ./cmd/termd` pass

## 2. Directory README

- [x] 2.1 Add a `writeDir` helper to `cmd/termd/main_test.go` that creates files, and directories for names ending in `/`, in a temp directory, each file holding `# <name>`; add `TestDirectoryInput` with one case per scenario of `specs/cli/spec.md`: `README.md` and `guide.md` (renders `# README.md`), `README`, `README.markdown` and `README.md` (`# README.md`), `README` and `README.markdown` (`# README.markdown`), only `readme.md` (`# readme.md`), a directory `README.md/` and a file `README` (`# README`), only `guide/README.md`, and only `guide.md`; the last two pass the directory with a trailing separator and expect exit 1, empty stdout and stderr exactly `termd: no README in <dir>/` and a newline, the others expect exit 0 and the heading in stdout; verify `go test ./cmd/termd` fails with `is a directory`
- [x] 2.2 Add `TestDirectorySymlinks`: `termd <link>`, where `<link>` is a symbolic link to a directory with `README.md`, renders `# README.md`, and `termd docs`, where `docs/README.md` is a symbolic link to `../README.md`, renders the content of the target, and a `README.md` linking to a missing file next to a file `README` renders `# README`; all three skip when `os.Symlink` fails; verify `go test ./cmd/termd` fails
- [x] 2.3 Add a case-variant test that fakes `stat`, `readDir` and `readFile` with a `testing/fstest.MapFS` holding `docs/readme.md` and `docs/README.md` with different content, and expects the content of `docs/README.md`; add error tests: `readFile` faked to fail for `<dir>/README.md` gives exit 1 and stderr naming `<dir>/README.md`, and `readDir` faked to fail gives exit 1 and its error on stderr; verify `go test ./cmd/termd` fails
- [x] 2.4 Add a test that renders a `README.md` of five long paragraphs with bold text with `--width 60` through a terminal `fake` (width 80, height 24, `less` available) once as `<dir>` and once as `<dir>/README.md`, and expects the same pager content and no direct output in both runs; verify `go test ./cmd/termd` fails
- [x] 2.5 Add `inputPath(arg, stat, readDir)` to `cmd/termd/main.go` as in the design (directory detection through `stat`, one `readDir`, the names in order compared with `strings.EqualFold`, only regular files count, the path joined from the argument and the stored name, `no README in <arg>` otherwise), and in `run` resolve the argument into `file`, declared next to `src` and `err`, before `e.readFile(file)`; verify the tests of 2.1-2.4 pass and the existing tests of `go test ./cmd/termd` pass unchanged

## 3. Usage text

- [x] 3.1 Change the usage text in `cmd/termd/main.go` to `Usage: termd [flags] [FILE | DIR]` with a sentence that a directory renders its `README.md`, `README.markdown` or `README`; verify `go run ./cmd/termd --help` shows it, every line of it fits in 80 columns, and `TestInvalidFlags` passes

## 4. Documentation

- [x] 4.1 In the Usage section of `README.md` add the example `termd docs/` that renders the README of a directory, without changing the other lines; verify the statement matches `specs/cli/spec.md`

## 5. Verification

- [x] 5.1 Run `make check` and `openspec validate open-directory-readme --strict`; verify both pass and `git status testdata/golden` shows no change
- [x] 5.2 Run `make build` and try `./termd --no-pager` on `.`, on `openspec/` and on `README.md`; verify `.` renders the README of the repository with the same output as `README.md` (compared with `diff`), and `openspec/` prints `termd: no README in openspec/` with exit status 1
