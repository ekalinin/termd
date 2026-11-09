## ADDED Requirements

### Requirement: Base URL flag
termd SHALL accept `--base-url URL`. When it is given and hyperlinks are enabled, termd SHALL resolve the destinations of links and images against `URL` as described in the "Hyperlinks" requirement of the markdown-rendering capability, for a document read from a file and from stdin. The value SHALL be an absolute URL with a scheme and a host, such as `https://github.com/o/r/blob/main/`, or a `file:` URL with an absolute path, such as `file:///home/me/proj/`. Any other value, including an empty one, SHALL be a usage error with exit status 2, and the message SHALL name `--base-url`. When hyperlinks are disabled, a valid value SHALL NOT change the output.

#### Scenario: Document from stdin
- **WHEN** the user runs `curl -s https://raw.githubusercontent.com/o/r/main/README.md | termd --hyperlinks=always --base-url https://github.com/o/r/blob/main/` and the document contains `[guide](docs/guide.md)`
- **THEN** the hyperlink of `guide` is `https://github.com/o/r/blob/main/docs/guide.md`

#### Scenario: Document from a file
- **WHEN** the user runs `termd --hyperlinks=always --base-url https://example.com/docs/ README.md` in the directory `/home/u/proj` and the file contains `[a](a.md)`
- **THEN** the hyperlink of `a` is `https://example.com/docs/a.md`

#### Scenario: File URL
- **WHEN** the user runs `cat README.md | termd --hyperlinks=always --base-url file:///home/me/proj/` and the document contains `[guide](docs/guide.md)`
- **THEN** the hyperlink of `guide` is `file:///home/me/proj/docs/guide.md`

#### Scenario: Hyperlinks disabled
- **WHEN** the user runs `termd --base-url https://example.com/ README.md | cat` and the file contains `[guide](docs/guide.md)`
- **THEN** the output contains `guide (docs/guide.md)`

#### Scenario: Value without a scheme
- **WHEN** the user passes `--base-url github.com/o/r/blob/main/`
- **THEN** termd prints a usage error that names `--base-url` to stderr and exits with status 2

#### Scenario: Value without a host
- **WHEN** the user passes `--base-url localhost:8080/docs/` or `--base-url https:///docs/`
- **THEN** termd prints a usage error that names `--base-url` to stderr and exits with status 2

#### Scenario: Empty value
- **WHEN** the user passes `--base-url=`
- **THEN** termd prints a usage error that names `--base-url` to stderr and exits with status 2
