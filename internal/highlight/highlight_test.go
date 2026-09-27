package highlight

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ekalinin/termd/internal/golden"
	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/text"
)

var samples = map[string]string{
	"go":     "package main\n\nimport \"fmt\"\n\n// main prints a greeting.\nfunc main() {\n\tfmt.Println(\"hello\", 42)\n}",
	"python": "def greet(name: str) -> str:\n    \"\"\"Return a greeting.\"\"\"\n    return f\"hello, {name}\"  # comment",
	"sh":     "#!/bin/sh\nfor f in *.md; do\n  echo \"$f\" | tr a-z A-Z   # upper-case\ndone",
}

var sgrRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func render(lines []text.Line, depth style.Depth) string {
	o := style.Options{Styled: true, Depth: depth}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.Render(o))
		b.WriteByte('\n')
	}
	return b.String()
}

func TestHighlightGolden(t *testing.T) {
	themes := map[string]Theme{"dark": Dark, "light": Light}
	depths := map[string]style.Depth{"truecolor": style.TrueColor, "256": style.Color256}
	for lang, code := range samples {
		for themeName, theme := range themes {
			for depthName, depth := range depths {
				t.Run(fmt.Sprintf("%s/%s/%s", lang, themeName, depthName), func(t *testing.T) {
					lines := Highlight(code, lang, theme)
					if lines == nil {
						t.Fatalf("%s is not highlighted", lang)
					}
					got := render(lines, depth)
					if depth == style.Color256 && strings.Contains(got, "38;2;") {
						t.Errorf("256-color output contains 24-bit sequences")
					}
					if depth == style.TrueColor && !strings.Contains(got, "38;2;") {
						t.Errorf("truecolor output has no 24-bit sequences")
					}
					golden.Assert(t, filepath.Join("testdata", fmt.Sprintf("%s.%s.%s.golden", lang, themeName, depthName)), got)
				})
			}
		}
	}
}

func TestHighlightKeepsSource(t *testing.T) {
	for lang, code := range samples {
		for _, theme := range []Theme{Dark, Light} {
			got := sgrRE.ReplaceAllString(render(Highlight(code, lang, theme), style.TrueColor), "")
			if got != code+"\n" {
				t.Errorf("%s: stripped output differs from source\n got %q\nwant %q", lang, got, code+"\n")
			}
		}
	}
}

func TestHighlightUsesDistinctColors(t *testing.T) {
	colors := map[style.Color]bool{}
	for _, l := range Highlight(samples["go"], "go", Dark) {
		for _, sp := range l {
			if sp.Style.FG.Set {
				colors[sp.Style.FG] = true
			}
		}
	}
	if len(colors) < 3 {
		t.Errorf("only %d distinct colors for keywords, strings and comments", len(colors))
	}
}

func TestRecognized(t *testing.T) {
	for _, info := range []string{"go", "Go", "js", "sh", "yml", "python title=x"} {
		if !Recognized(info) {
			t.Errorf("%q is not recognized", info)
		}
	}
	for _, info := range []string{"", "foo", "   "} {
		if Recognized(info) {
			t.Errorf("%q is recognized", info)
		}
		if Highlight("x := 1", info, Dark) != nil {
			t.Errorf("%q is highlighted", info)
		}
	}
}
