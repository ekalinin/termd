# Proposal

## Why

termd has releases since `v0.1.0`, but the landing page shows only the `go install ...@latest` command and a link to the repository. A reader of the page cannot see which version is the latest one.

## What Changes

- The hero gets a line under the install command and the GitHub button: the shields.io badge of the latest release, `https://img.shields.io/github/v/release/ekalinin/termd`, with the default style, label and colors. The badge links to `https://github.com/ekalinin/termd/releases/latest`.
- The reader's browser loads the badge; the version number is not part of the built page. A new release appears on the page without a new site build, after the shields.io cache expires (5 minutes).
- Pre-releases are not shown: both the badge and `/releases/latest` skip them.
- The page still has no JavaScript.

Out of scope:

- The version number written into the page at build time, and a Pages deployment started by the release workflow.
- A custom badge style or colors, and separate badges for the light and dark themes.
- A pinned install command (`@vX.Y.Z`): the page and the README keep `@latest`.
- The release date and download links for the platforms.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `project-site`: the page composition includes the latest release badge that links to the latest release.

## Impact

- `internal/site/page.html`: the badge line in the hero. `internal/site/style.css`: the spacing of that line.
- `internal/site/site.go`: the badge and latest release URLs passed to the template, as the repository and license URLs are.
- `internal/site/site_test.go`: `TestPageComposition` checks the link and the badge.
- `pages.yml`, `release.yml`, `cmd/termd-site` and the site build do not change. No new Go dependencies.
- The page gets its first external resource: every page view requests the badge from img.shields.io. When the badge cannot be loaded, the link shows its alt text.
