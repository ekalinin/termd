# Spec Delta

## MODIFIED Requirements

### Requirement: Input source
termd SHALL read the markdown document from the path given as its single positional argument. When the path is a directory, termd SHALL read the README of that directory and render it as when the path of the README is given as the argument. When the argument is `-`, or when no argument is given and stdin is not a terminal, termd SHALL read the document from stdin.

The README of a directory SHALL be the first of `README.md`, `README.markdown` and `README`, in this order, that is an entry of the directory itself, with the names compared case-insensitively. When several entries differ from a name only in case, the first of them in byte order of their names SHALL be used, so `README.md` comes before `readme.md`. An entry SHALL count only when it is a regular file, directly or through symbolic links; any other entry, such as a directory named `README.md` or a broken symbolic link, SHALL be skipped. termd SHALL NOT look into subdirectories. A symbolic link to a directory given as the argument SHALL be treated as that directory. When the directory has no README, termd SHALL print `termd: no README in <dir>` to stderr, where `<dir>` is the argument as given, and exit with status 1. When the directory or its README cannot be read, termd SHALL print an error naming the directory or the README to stderr and exit with status 1.

#### Scenario: Render a file
- **WHEN** the user runs `termd README.md` and the file exists
- **THEN** termd renders the contents of `README.md` and exits with status 0

#### Scenario: Render the README of a directory
- **WHEN** the user runs `termd docs/` and `docs/` contains `README.md` and `guide.md`
- **THEN** termd renders the contents of `docs/README.md` and exits with status 0

#### Scenario: Several README candidates
- **WHEN** the user runs `termd docs/` and `docs/` contains `README`, `README.markdown` and `README.md`
- **THEN** termd renders the contents of `docs/README.md`

#### Scenario: README.markdown before README
- **WHEN** the user runs `termd docs/` and `docs/` contains `README` and `README.markdown`
- **THEN** termd renders the contents of `docs/README.markdown`

#### Scenario: Lowercase name
- **WHEN** the user runs `termd docs/` and the only README in `docs/` is `readme.md`
- **THEN** termd renders the contents of `docs/readme.md`

#### Scenario: Names that differ only in case
- **WHEN** the user runs `termd docs/` on a case-sensitive file system and `docs/` contains `readme.md` and `README.md`
- **THEN** termd renders the contents of `docs/README.md`

#### Scenario: README entry that is a directory
- **WHEN** the user runs `termd docs/` and `docs/` contains a directory `README.md` and a file `README`
- **THEN** termd renders the contents of `docs/README`

#### Scenario: Symbolic link to a directory
- **WHEN** `link` is a symbolic link to the directory `docs/`, which contains `README.md`, and the user runs `termd link`
- **THEN** termd renders the contents of `docs/README.md`

#### Scenario: README that is a symbolic link
- **WHEN** the user runs `termd docs/` and `docs/README.md` is a symbolic link to `../README.md`
- **THEN** termd renders the contents of `README.md`

#### Scenario: README only in a subdirectory
- **WHEN** the user runs `termd docs/` and `docs/` contains `guide/README.md` but no README of its own
- **THEN** termd prints `termd: no README in docs/` to stderr and exits with status 1

#### Scenario: Directory without a README
- **WHEN** the user runs `termd docs/` and `docs/` contains only `guide.md`
- **THEN** termd prints `termd: no README in docs/` to stderr, writes nothing to stdout and exits with status 1

#### Scenario: Unreadable README
- **WHEN** the user runs `termd docs/` and `docs/README.md` exists but cannot be read
- **THEN** termd prints an error naming `docs/README.md` to stderr and exits with status 1

#### Scenario: Same output as for the README path
- **WHEN** the user runs `termd --width 60 docs/` in a terminal and `docs/README.md` is the README
- **THEN** the output, and whether it is shown in the pager, is the same as for `termd --width 60 docs/README.md`

#### Scenario: Render piped input
- **WHEN** the user runs `cat README.md | termd`
- **THEN** termd renders the piped document and exits with status 0

#### Scenario: Explicit stdin
- **WHEN** the user runs `termd -` with a document piped to stdin
- **THEN** termd renders the piped document

#### Scenario: No input available
- **WHEN** the user runs `termd` with no argument and stdin is a terminal
- **THEN** termd prints usage information to stderr and exits with status 2

#### Scenario: Unreadable file
- **WHEN** the user runs `termd missing.md` and the file does not exist or cannot be read
- **THEN** termd prints an error naming the file to stderr and exits with status 1
