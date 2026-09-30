package site

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/render"
	"github.com/ekalinin/termd/internal/style"
)

var (
	//go:embed examples/*.md
	examplesFS embed.FS
	//go:embed page.html
	pageHTML string
	//go:embed style.css
	styleCSS []byte

	pageTmpl = template.Must(template.New("page").Funcs(template.FuncMap{"ruler": ruler}).Parse(pageHTML))
)

const (
	repoURL        = "https://github.com/ekalinin/termd"
	licenseURL     = repoURL + "/blob/main/LICENSE"
	releaseURL     = repoURL + "/releases/latest"
	installCommand = "go install github.com/ekalinin/termd/cmd/termd@latest"
	description    = "A terminal markdown viewer that renders tables and diagrams correctly."
	// badgeURL shows the tag of the release that releaseURL opens; the
	// reader's browser loads it, so the page needs no rebuild on a release.
	badgeURL = "https://img.shields.io/github/v/release/ekalinin/termd"
	// defaultWidth is the width selected when the page opens: the width
	// termd uses when it does not write to a terminal.
	defaultWidth = 80
)

// widths are the example widths the reader can choose.
var widths = []int{40, 60, 80}

// themes are the highlighting themes an example is rendered with.
var themes = []struct {
	name  string
	theme highlight.Theme
}{{"dark", highlight.Dark}, {"light", highlight.Light}}

// Example is an example document rendered at every width. File is the
// document name shown in the command above the output.
type Example struct {
	Name   string
	File   string
	Widths []Width
}

// Width holds the renderings of an example at one width: one fragment when
// the output does not depend on the highlighting theme, otherwise one per
// theme.
type Width struct {
	Cols      int
	Fragments []Fragment
}

// Fragment is the HTML of one rendering; Theme is "dark", "light" or "" when
// the rendering is the same for both themes.
type Fragment struct {
	Theme string
	HTML  template.HTML
}

// renderANSI returns termd's styled output for src.
func renderANSI(src []byte, width int, theme highlight.Theme) string {
	return render.Render(src, render.Options{
		Width: width,
		Style: style.Options{Styled: true, Hyperlinks: true, Depth: style.TrueColor},
		Theme: func() highlight.Theme { return theme },
	})
}

// Examples renders every example document at every width and theme.
func Examples() ([]Example, error) {
	files, err := fs.Glob(examplesFS, "examples/*.md")
	if err != nil {
		return nil, err
	}
	var out []Example
	for _, file := range files {
		src, err := examplesFS.ReadFile(file)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(path.Base(file), ".md")
		// A numeric prefix only orders the examples on the page.
		ex := Example{Name: name, File: strings.TrimLeft(name, "0123456789-") + ".md"}
		for _, cols := range widths {
			outputs := make([]string, len(themes))
			for i, t := range themes {
				outputs[i] = renderANSI(src, cols, t.theme)
			}
			w := Width{Cols: cols}
			same := outputs[0] == outputs[1]
			for i, t := range themes {
				h, err := HTML(outputs[i])
				if err != nil {
					return nil, fmt.Errorf("example %s at width %d (%s theme): %w", ex.Name, cols, t.name, err)
				}
				if same {
					w.Fragments = append(w.Fragments, Fragment{HTML: template.HTML(h)})
					break
				}
				w.Fragments = append(w.Fragments, Fragment{Theme: t.name, HTML: template.HTML(h)})
			}
			ex.Widths = append(ex.Widths, w)
		}
		out = append(out, ex)
	}
	return out, nil
}

// ruler returns a column ruler n columns wide, as terminal editors draw it:
// "----+----1----+----2", with the tens digits in <b>.
func ruler(n int) template.HTML {
	var b strings.Builder
	for col := 1; col <= n; col++ {
		switch {
		case col%10 == 0:
			fmt.Fprintf(&b, "<b>%d</b>", col/10%10)
		case col%5 == 0:
			b.WriteByte('+')
		default:
			b.WriteByte('-')
		}
	}
	return template.HTML(b.String())
}

// Page returns the HTML of the landing page.
func Page() ([]byte, error) {
	examples, err := Examples()
	if err != nil {
		return nil, err
	}
	data := struct {
		Description, Install, Repo, License string
		Release, Badge                      string
		Widths                              []int
		DefaultWidth                        int
		Examples                            []Example
	}{description, installCommand, repoURL, licenseURL, releaseURL, badgeURL, widths, defaultWidth, examples}
	var b bytes.Buffer
	if err := pageTmpl.Execute(&b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Build writes index.html and style.css into dir, creating it when needed.
func Build(dir string) error {
	page, err := Page()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), page, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "style.css"), styleCSS, 0o644)
}
