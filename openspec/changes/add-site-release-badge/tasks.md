# Tasks

## 1. Page

- [x] 1.1 Extend `TestPageComposition` in `internal/site/site_test.go`: the hero (the part of the page before `</header>`) contains a link to `releaseURL` with `<img src="` + `badgeURL` + `" alt="Latest release">` inside, after the `.actions` row; verify `go test ./internal/site` fails because the constants and the markup do not exist yet
- [x] 1.2 In `internal/site/site.go` add the constants `releaseURL = repoURL + "/releases/latest"` and `badgeURL = "https://img.shields.io/github/v/release/ekalinin/termd"` next to `repoURL` and `licenseURL`, and the fields `Release` and `Badge` in the page data; in `internal/site/page.html` add `<p class="release"><a href="{{.Release}}"><img src="{{.Badge}}" alt="Latest release"></a></p>` after the `.actions` block inside `<header class="hero">`, with no `width` or `height` on the image; verify `go test ./internal/site` passes
- [x] 1.3 In `internal/site/style.css` add the `.release` rule: a top margin equal to the `.actions` gap (`.75rem`), `min-height: 20px`, no extra space under the image; verify with `make site` and `_site/index.html` in a browser: the badge shows `release | v0.1.0` under the install row, the page does not shift when the badge loads, and the line looks the same with the light and dark themes and in a window narrower than the install row
- [x] 1.4 Verify the fallback: block `img.shields.io` in the browser dev tools (request blocking) and reload `_site/index.html`; the page shows the text `Latest release` in place of the badge, and clicking it opens `https://github.com/ekalinin/termd/releases/latest`
- [x] 1.5 Verify the build does not depend on releases: `grep -E 'v[0-9]+\.[0-9]+\.[0-9]+' _site/index.html` finds nothing, and `make check` passes

## 2. After the merge

- [ ] 2.1 After the `Pages` run for the merge commit passes, open `https://ekalinin.github.io/termd/`; verify the badge shows the same tag as `gh release view --json tagName --jq .tagName`, and clicking it opens that release
