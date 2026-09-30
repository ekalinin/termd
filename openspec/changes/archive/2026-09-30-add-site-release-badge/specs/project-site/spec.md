# Spec Delta

## MODIFIED Requirements

### Requirement: Page location and composition
The site SHALL be published at `https://ekalinin.github.io/termd/` as a single page. The page SHALL contain the project name, a one-line description, the install command, a link to the GitHub repository, the version of the latest release, the examples and a footer with links to the repository and the license. The page SHALL NOT reproduce the documentation sections of the README.

#### Scenario: Opening the page
- **WHEN** a reader opens `https://ekalinin.github.io/termd/`
- **THEN** the page shows the project name, the one-line description, the install command, a link to `https://github.com/ekalinin/termd`, the version of the latest release, the five examples and a footer linking to the repository and the license

#### Scenario: Install command matches the README
- **WHEN** the page is built
- **THEN** the install command on the page is the same `go install` command as in the Install section of the README

## ADDED Requirements

### Requirement: Latest release
The page SHALL show the version of the latest release under the install command, as a badge image that the reader's browser loads from an external badge service when the page is viewed. The version SHALL NOT be written into the page when the site is built. The badge SHALL link to `https://github.com/ekalinin/termd/releases/latest`. The badge and the link SHALL skip pre-releases. When the badge image cannot be loaded, the link SHALL show the text `Latest release`.

#### Scenario: Latest release on the page
- **WHEN** `v0.1.0` is the latest release and a reader opens the page
- **THEN** the page shows a badge with `v0.1.0` under the install command, and clicking the badge opens the release `v0.1.0`

#### Scenario: New release without a site build
- **WHEN** the release `v0.2.0` is published and no commit is pushed to `main` after it
- **THEN** once the cached badge expires, the page shows `v0.2.0` without a new site build

#### Scenario: Pre-release
- **WHEN** the release before the pre-release `v0.2.0-rc.1` is `v0.1.0`, and `v0.2.0-rc.1` is published
- **THEN** the badge shows `v0.1.0`, and clicking it opens the release `v0.1.0`

#### Scenario: Badge not available
- **WHEN** the badge image cannot be loaded
- **THEN** the page shows the text `Latest release` in place of the badge, and clicking it opens the latest release

#### Scenario: Build does not depend on releases
- **WHEN** the site is built from the same commit before and after a new release is published
- **THEN** both builds write the same page
