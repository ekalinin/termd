package site

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/termd/internal/theme"
)

var (
	escRE = regexp.MustCompile(`\x1b\[[0-9;]*m|\x1b\]8;;[^\x1b]*\x1b\\`)
	tagRE = regexp.MustCompile(`<[^>]*>`)
)

// TestFragmentsMatchRenderer checks that the text of every fragment is
// termd's styled output without escape sequences.
func TestFragmentsMatchRenderer(t *testing.T) {
	examples, err := Examples()
	if err != nil {
		t.Fatal(err)
	}
	if len(examples) != 5 {
		t.Fatalf("%d examples, want 5", len(examples))
	}
	for _, ex := range examples {
		src, err := examplesFS.ReadFile("examples/" + ex.Name + ".md")
		if err != nil {
			t.Fatal(err)
		}
		if len(ex.Widths) != len(widths) {
			t.Fatalf("%s: %d widths, want %d", ex.Name, len(ex.Widths), len(widths))
		}
		for _, w := range ex.Widths {
			for _, th := range themes {
				t.Run(fmt.Sprintf("%s/%d/%s", ex.Name, w.Cols, th.name), func(t *testing.T) {
					var frag *Fragment
					for i := range w.Fragments {
						if f := &w.Fragments[i]; f.Theme == "" || f.Theme == th.name {
							frag = f
						}
					}
					if frag == nil {
						t.Fatal("no fragment")
					}
					want := escRE.ReplaceAllString(renderANSI(src, w.Cols, th.theme), "")
					got := html.UnescapeString(tagRE.ReplaceAllString(string(frag.HTML), ""))
					if got != want {
						t.Errorf("fragment text differs from the renderer\n--- want\n%s\n--- got\n%s", want, got)
					}
				})
			}
		}
	}
}

// TestThemeFragments checks that only the code example has a rendering per
// theme.
func TestThemeFragments(t *testing.T) {
	examples, err := Examples()
	if err != nil {
		t.Fatal(err)
	}
	for _, ex := range examples {
		want := 1
		if ex.Name == "5-code" {
			want = 2
		}
		for _, w := range ex.Widths {
			if len(w.Fragments) != want {
				t.Errorf("%s at %d: %d fragments, want %d", ex.Name, w.Cols, len(w.Fragments), want)
			}
		}
	}
}

func page(t *testing.T) string {
	t.Helper()
	p, err := Page()
	if err != nil {
		t.Fatal(err)
	}
	return string(p)
}

func TestPageComposition(t *testing.T) {
	p := page(t)
	for _, want := range []string{
		"<h1>termd</h1>",
		`<p class="tagline">` + html.EscapeString(description) + "</p>",
		"<code>" + installCommand + "</code>",
		`<a class="github" href="https://github.com/ekalinin/termd">`,
		`<a href="https://github.com/ekalinin/termd/blob/main/LICENSE">`,
	} {
		if !strings.Contains(p, want) {
			t.Errorf("page has no %q", want)
		}
	}
	if n := strings.Count(p, `<section class="example"`); n != 5 {
		t.Errorf("%d examples on the page, want 5", n)
	}
	hero := p[:strings.Index(p, "</header>")]
	release := `<a href="` + releaseURL + `"><img src="` + badgeURL + `" alt="Latest release"></a>`
	if i := strings.Index(hero, release); i < strings.Index(hero, `<div class="actions">`) {
		t.Errorf("hero has no %q after the install command", release)
	}
	footer := p[strings.Index(p, "<footer>"):]
	for _, want := range []string{repoURL + `"`, licenseURL + `"`} {
		if !strings.Contains(footer, want) {
			t.Errorf("footer has no link to %s", want)
		}
	}
}

// TestInstallCommandMatchesReadme checks the install command against the
// go install command in the README.
func TestInstallCommandMatchesReadme(t *testing.T) {
	f, err := os.Open("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var cmds []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); strings.HasPrefix(line, "go install ") {
			cmds = append(cmds, line)
		}
	}
	if len(cmds) != 1 || cmds[0] != installCommand {
		t.Errorf("README go install commands %q, page shows %q", cmds, installCommand)
	}
}

func TestWidthSwitcher(t *testing.T) {
	p := page(t)
	inputs := regexp.MustCompile(`<input type="radio" name="width"[^>]*>`).FindAllString(p, -1)
	if len(inputs) != 3 {
		t.Fatalf("%d width inputs, want 3", len(inputs))
	}
	for i, in := range inputs {
		checked := strings.Contains(in, " checked")
		if want := fmt.Sprintf(`id="width-%d"`, widths[i]); !strings.Contains(in, want) {
			t.Errorf("input %q, want %s", in, want)
		}
		if checked != (widths[i] == 80) {
			t.Errorf("input %q: checked = %v", in, checked)
		}
	}
	if n := strings.Count(p, `role="radiogroup" aria-label="Width"`); n != 1 {
		t.Errorf("%d width switchers, want 1", n)
	}
	css := string(styleCSS)
	for _, w := range widths {
		if w != defaultWidth && !strings.Contains(css, fmt.Sprintf("#width-%d:checked) .cols-%d", w, w)) {
			t.Errorf("style.css does not show .cols-%d for #width-%d", w, w)
		}
	}
}

func TestRuler(t *testing.T) {
	for cols, want := range map[int]string{
		10: "----+----<b>1</b>",
		40: "----+----<b>1</b>----+----<b>2</b>----+----<b>3</b>----+----<b>4</b>",
	} {
		if got := string(ruler(cols)); got != want {
			t.Errorf("ruler(%d) = %q, want %q", cols, got, want)
		}
		if got := tagRE.ReplaceAllString(string(ruler(cols)), ""); len(got) != cols {
			t.Errorf("ruler(%d) is %d columns wide", cols, len(got))
		}
	}
}

// TestWindowCommands checks the command above each rendering: it names the
// width and, for the code example, the theme.
func TestWindowCommands(t *testing.T) {
	p := page(t)
	for _, want := range []string{
		"termd --width 80 table.md</figcaption>",
		"termd --width 40 flowchart.md</figcaption>",
		"termd --width 60 --theme light code.md</figcaption>",
		"termd --width 60 --theme dark code.md</figcaption>",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("page has no %q", want)
		}
	}
}

func TestNoJavaScript(t *testing.T) {
	if p := strings.ToLower(page(t)); strings.Contains(p, "<script") || strings.Contains(p, "javascript:") {
		t.Error("page contains JavaScript")
	}
}

func TestBuild(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "site")
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.html", "style.css"} {
		if fi, err := os.Stat(filepath.Join(dir, name)); err != nil || fi.Size() == 0 {
			t.Errorf("%s not written: %v", name, err)
		}
	}
}

// lightThemes are the themes whose showcase window has a light background.
var lightThemes = []string{"light", "solarized-light", "gruvbox-light", "catppuccin-latte"}

func TestShowcaseMatchesRenderer(t *testing.T) {
	showcase, err := Showcase()
	if err != nil {
		t.Fatal(err)
	}
	if len(showcase) != len(widths) {
		t.Fatalf("%d widths, want %d", len(showcase), len(widths))
	}
	for i, w := range showcase {
		if w.Cols != widths[i] {
			t.Errorf("width %d is %d columns, want %d", i, w.Cols, widths[i])
		}
		var names []string
		for _, win := range w.Windows {
			names = append(names, win.Theme)
		}
		if !slices.Equal(names, theme.Names()) {
			t.Fatalf("width %d: themes %q, want %q", w.Cols, names, theme.Names())
		}
		for _, win := range w.Windows {
			th, _ := theme.Get(win.Theme)
			want := escRE.ReplaceAllString(renderTheme(themesMD, w.Cols, th), "")
			got := html.UnescapeString(tagRE.ReplaceAllString(string(win.HTML), ""))
			if got != want {
				t.Errorf("%s at %d: text differs from the renderer\n--- want\n%s\n--- got\n%s", win.Theme, w.Cols, want, got)
			}
		}
	}
}

func TestShowcasePage(t *testing.T) {
	p := page(t)
	section := strings.Index(p, `<section class="showcase"`)
	if section < 0 || section < strings.LastIndex(p, `<section class="example"`) {
		t.Fatal("no showcase section after the examples")
	}
	last := section
	for _, name := range theme.Names() {
		input := `<input type="radio" name="termd-theme" id="termd-theme-` + name + `"`
		i := strings.Index(p, input)
		if i < last {
			t.Errorf("no switcher input for %s after the previous one", name)
		}
		last = i
		checked := strings.HasPrefix(p[i+len(input):], " checked")
		if checked != (name == "dark") {
			t.Errorf("%s checked = %v", name, checked)
		}
		if rule := ":root:has(#termd-theme-" + name + ":checked) .showcase .theme-" + name + " "; !strings.Contains(p, rule) {
			t.Errorf("no rule %q", rule)
		}
		scheme := "color-scheme:dark"
		if slices.Contains(lightThemes, name) {
			scheme = "color-scheme:light"
		}
		for _, cols := range widths {
			title := fmt.Sprintf("termd --width %d --theme %s themes.md</figcaption>", cols, name)
			if !strings.Contains(p, title) {
				t.Errorf("no window titled %q", title)
			}
		}
		window := `<figure class="window theme-` + name + `" style="` + scheme
		if n := strings.Count(p, window); n != len(widths) {
			t.Errorf("%d windows starting with %q, want %d", n, window, len(widths))
		}
	}
	if !strings.Contains(p, `<figure class="window theme-dracula" style="color-scheme:dark;--win-bg:#282a36;--win-fg:#f8f8f2">`) {
		t.Error("dracula window does not have the colors of the dracula style")
	}
}

func TestBuildShots(t *testing.T) {
	dir := t.TempDir()
	if err := BuildShots(dir); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(theme.Names())+1 {
		t.Errorf("%d files written, want a page per theme and style.css", len(entries))
	}
	if fi, err := os.Stat(filepath.Join(dir, "style.css")); err != nil || fi.Size() == 0 {
		t.Errorf("style.css not written: %v", err)
	}
	preRE := regexp.MustCompile(`(?s)<pre class="term">(.*)</pre>`)
	for _, name := range theme.Names() {
		b, err := os.ReadFile(filepath.Join(dir, name+".html"))
		if err != nil {
			t.Fatal(err)
		}
		p := string(b)
		if n := strings.Count(p, "<figure"); n != 1 || !strings.Contains(p, `<figure class="window theme-`+name+`" style="`) {
			t.Errorf("%s: %d windows, want one of the theme", name, n)
		}
		if title := "termd --width 60 --theme " + name + " themes.md</figcaption>"; !strings.Contains(p, title) {
			t.Errorf("%s: no title %q", name, title)
		}
		if strings.Contains(p, `class="ruler"`) {
			t.Errorf("%s: page has a ruler", name)
		}
		m := preRE.FindStringSubmatch(p)
		if m == nil {
			t.Fatalf("%s: no output", name)
		}
		th, _ := theme.Get(name)
		want := escRE.ReplaceAllString(renderTheme(themesMD, 60, th), "")
		if got := html.UnescapeString(tagRE.ReplaceAllString(m[1], "")); got != want {
			t.Errorf("%s: text differs from the renderer\n--- want\n%s\n--- got\n%s", name, want, got)
		}
	}
}

// TestThemeScreenshots checks that every theme has a screenshot in
// docs/themes and that the README shows it.
func TestThemeScreenshots(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range theme.Names() {
		file := "docs/themes/" + name + ".png"
		if fi, err := os.Stat("../../" + file); err != nil || fi.Size() == 0 {
			t.Errorf("%s: no screenshot %s (run make screenshots): %v", name, file, err)
		}
		if !strings.Contains(string(readme), "("+file+")") {
			t.Errorf("%s: README does not show %s", name, file)
		}
	}
}
