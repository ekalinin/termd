# Design

## Context

See proposal.md - Why for the motivation, and `specs/project-site/spec.md` for the requirements.

State of the code and the services that shapes the approach:

- `internal/site/site.go` passes the page data to `page.html` in an anonymous struct: `Description`, `Install`, `Repo`, `License`, the widths and the examples. The URLs are constants (`repoURL`, `licenseURL`), and `TestPageComposition` checks the page against the same constants.
- The hero in `page.html` is the `<h1>`, the tagline and the `.actions` row with the install command and the GitHub button (`.actions` is a wrapping flex row with a gap of `.75rem`).
- The page is published by `pages.yml` on every push to `main`. A release is published by `release.yml` after the merge; GitHub creates the tag with `GITHUB_TOKEN`, so the release starts no other workflow and the page is not rebuilt.
- `https://img.shields.io/github/v/release/ekalinin/termd` returns an SVG of 94x20 px with `release | v0.1.0` and `cache-control: max-age=300, s-maxage=300`. With the default parameters (`sort=date`, no `include_prereleases`, no `filter`), shields.io asks GitHub for `/repos/ekalinin/termd/releases/latest` and shows the tag of that release (`display_name` defaults to `tag`). GitHub's latest release skips drafts and pre-releases.
- The page has no external resources yet.

## Goals / Non-Goals

**Goals:**
- The number on the page and the release the link opens are the same release.
- The site build does not depend on releases or the network.

**Non-Goals:**
- Matching the badge to the page palette or to the light and dark themes.
- Hiding the badge line when shields.io is unavailable.

## Decisions

### A shields.io badge instead of a number written at build time
The page shows `<img src="https://img.shields.io/github/v/release/ekalinin/termd">` inside a link to `https://github.com/ekalinin/termd/releases/latest`. The reader's browser loads the number when the page is viewed, so a release shows up on the page without a Pages deployment.
Alternatives:
- The latest tag written into the page by `termd-site`, with `pages.yml` turned into a reusable workflow that `release.yml` calls after publishing the draft (as `ci.yml` is called). The page stays free of external resources, but it changes both workflows, `pages.yml` needs the tags (`fetch-depth: 0`), and a local build without fetched tags differs from the published page.
- The same with the GitHub API (`gh release view`) at build time: the site build needs the network and a token.
- A link "Latest release" without the number: always correct, but shows no version.
- JavaScript that fetches the release from the GitHub API: the page has no JavaScript.

### Default badge parameters
No query parameters: flat style, label `release`, the shields.io colors (orange for `0.x`, blue from `1.0.0`). The defaults also give the release selection that matches the link: any of `sort=semver`, `include_prereleases` or `filter` makes shields.io pick the release itself instead of asking GitHub for the latest one.
Alternatives: `flat-square` with the page's muted and accent colors (one image for both themes); two images for the light and dark themes, switched by the same CSS rules as the `.window.light` and `.window.dark` fragments (two requests per view); `<picture>` with `prefers-color-scheme` (does not follow the manual theme switch).

### Markup
A block after `.actions` inside `<header class="hero">`:

```html
<p class="release"><a href="{{.Release}}"><img src="{{.Badge}}" alt="Latest release"></a></p>
```

- `alt="Latest release"`: the number is unknown at build time. When the image fails, the browser shows the alt text inside the link, so the link still works.
- No `width` or `height` on `<img>`: the width depends on the length of the version, and a fixed box can hide the alt text of a failed image. The `.release` block gets `min-height: 20px`, so the page does not shift when the badge arrives, and a top margin equal to the `.actions` gap. While the badge is loading, Chrome shows the alt text in its place; `line-height: 20px` on `.release` keeps that line 20px high too, otherwise it is 24px and the page shifts up by 4px when the badge arrives.
- `site.go` gets two constants next to `repoURL` and `licenseURL`: `releaseURL = repoURL + "/releases/latest"` and `badgeURL`, and two fields `Release` and `Badge` in the page data. `TestPageComposition` checks the link and the `src` against these constants.

## Risks / Trade-offs

- [Every page view sends a request to img.shields.io, a third party] → accepted; the badge is the only external resource, and the page does not depend on it to work.
- [shields.io is down or rate-limited by the GitHub API] → the badge shows an error text or the image fails and the alt text `Latest release` is shown; the link still opens the latest release.
- [The number lags behind a new release by the cache time] → at most a few minutes (`max-age=300`); the link is never stale.
- [The examples are rendered from the head of `main`, which can be ahead of the latest release] → accepted; the page does not claim that the examples come from the shown version.
- [The default colors change from orange to blue at `v1.0.0`] → accepted with the defaults.

## Migration Plan

1. Merge the change; `pages.yml` deploys the page.
2. Open `https://ekalinin.github.io/termd/` and check the badge and the link.

Rollback: remove the `.release` block, its CSS, the two constants and the check in `TestPageComposition`.
