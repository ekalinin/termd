package render

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/termd/internal/golden"
	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/theme"
)

const fixtures = "../../testdata"

// goldenOptions returns the options of a golden mode. Styled output uses an
// explicit theme and color depth so it never depends on the terminal.
func goldenOptions(width int, styled bool) Options {
	return Options{
		Width: width,
		Style: style.Options{Styled: styled, Hyperlinks: styled, Depth: style.TrueColor},
		Theme: func() highlight.Theme { return highlight.Dark },
	}
}

// TestGolden renders every testdata/*.md fixture at widths 40, 60 and 80 in
// plain and styled modes, and at width 80 in plain ASCII mode, and compares
// the output with testdata/golden.
func TestGolden(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(fixtures, "*.md"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no fixtures found: %v", err)
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimSuffix(filepath.Base(file), ".md")
		for _, width := range []int{40, 60, 80} {
			for _, mode := range []string{"plain", "styled"} {
				t.Run(fmt.Sprintf("%s/%d/%s", name, width, mode), func(t *testing.T) {
					got := Render(src, goldenOptions(width, mode == "styled"))
					path := filepath.Join(fixtures, "golden", fmt.Sprintf("%s.w%d.%s.golden", name, width, mode))
					golden.Assert(t, path, got)
				})
			}
		}
		t.Run(name+"/80/ascii", func(t *testing.T) {
			opts := goldenOptions(80, false)
			opts.ASCII = true
			golden.Assert(t, filepath.Join(fixtures, "golden", name+".w80.ascii.golden"), Render(src, opts))
		})
	}
}

// TestASCIIFixtures renders every testdata/*.md fixture in ASCII mode and
// checks that the output has no box-drawing, block, geometric-shape or arrow
// character and no other Unicode mark of termd that the fixture itself does
// not contain.
func TestASCIIFixtures(t *testing.T) {
	drawn := func(r rune) bool {
		return r >= 0x2190 && r <= 0x21FF || r >= 0x2500 && r <= 0x25FF || r == '•' || r == '‑' || r == '›'
	}
	files, err := filepath.Glob(filepath.Join(fixtures, "*.md"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no fixtures found: %v", err)
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimSuffix(filepath.Base(file), ".md")
		for _, width := range []int{40, 60, 80} {
			for _, mode := range []string{"plain", "styled"} {
				opts := goldenOptions(width, mode == "styled")
				opts.ASCII = true
				var found []rune
				for _, r := range stripEscapes(Render(src, opts)) {
					if drawn(r) && !strings.ContainsRune(string(src), r) && !slices.Contains(found, r) {
						found = append(found, r)
					}
				}
				if len(found) > 0 {
					t.Errorf("%s/%d/%s: %q", name, width, mode, string(found))
				}
			}
		}
	}
}

// TestCodeFileGolden renders every testdata/codefiles fixture as a code file.
// The plain output must be the fixture itself; the styled output is compared
// with testdata/golden/<file>.styled.golden. The width does not matter, since
// code is never wrapped.
func TestCodeFileGolden(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(fixtures, "codefiles", "*"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no code file fixtures found: %v", err)
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			if got := Code(src, name, goldenOptions(80, false)); got != string(src) {
				t.Errorf("plain output differs from the file\n--- want\n%s\n--- got\n%s", src, got)
			}
			got := Code(src, name, goldenOptions(80, true))
			if stripEscapes(got) != string(src) {
				t.Errorf("styled output without escapes differs from the file\n%s", got)
			}
			golden.Assert(t, filepath.Join(fixtures, "golden", name+".styled.golden"), got)
		})
	}
}

// TestThemeGolden renders testdata/themes/sample.md at width 60 with every
// built-in theme and compares the output with
// testdata/golden/theme.<name>.golden.
func TestThemeGolden(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(fixtures, "themes", "sample.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range theme.Names() {
		t.Run(name, func(t *testing.T) {
			th, _ := theme.Get(name)
			opts := goldenOptions(60, true)
			opts.Palette = th.Palette
			opts.Theme = func() highlight.Theme { return th.Code }
			golden.Assert(t, filepath.Join(fixtures, "golden", "theme."+name+".golden"), Render(src, opts))
		})
	}
}
