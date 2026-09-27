package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/termd/internal/golden"
	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/style"
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
// plain and styled modes and compares the output with testdata/golden.
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
	}
}
