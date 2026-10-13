# Spec Delta

## ADDED Requirements

### Requirement: HTML mode flag
termd SHALL accept `--html=render|source`, defaulting to `render`. In the mode `render` termd SHALL interpret the HTML tags of the document, and in the mode `source` it SHALL show HTML blocks and inline HTML as their source text, in a terminal and in a pipe. Any other value SHALL be a usage error with exit status 2, and the message SHALL name `--html` and list the valid values.

#### Scenario: Default mode
- **WHEN** the user runs `termd README.md | cat` without the flag and the document contains `Press <kbd>Ctrl</kbd>.`
- **THEN** the output is `Press Ctrl.`

#### Scenario: Source mode
- **WHEN** the user runs `termd --html=source README.md | cat` with the same document
- **THEN** the output is `Press <kbd>Ctrl</kbd>.`

#### Scenario: Invalid mode
- **WHEN** the user passes `--html=strip`
- **THEN** termd prints a usage error that names `--html` and lists `render` and `source` to stderr and exits with status 2
