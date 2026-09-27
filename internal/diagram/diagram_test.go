package diagram

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mdiagram "github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	mrender "github.com/AlexanderGrooff/mermaid-ascii/pkg/render"

	"github.com/ekalinin/termd/internal/golden"
	"github.com/ekalinin/termd/internal/text"
)

func lines(res Result) string {
	var b strings.Builder
	for _, l := range res.Lines {
		b.WriteString(l.String())
		b.WriteByte('\n')
	}
	return b.String()
}

func readFixture(t *testing.T, name string) string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

func TestLanguage(t *testing.T) {
	for _, info := range []string{"mermaid", "Mermaid", "plantuml", "puml", "mermaid title=x"} {
		if _, ok := Language(info); !ok {
			t.Errorf("%q is not a diagram", info)
		}
	}
	for _, info := range []string{"", "go", "mermaidjs"} {
		if _, ok := Language(info); ok {
			t.Errorf("%q is a diagram", info)
		}
	}
}

func TestMermaidType(t *testing.T) {
	tests := []struct{ src, want string }{
		{"sequenceDiagram\n  A->>B: x", "sequenceDiagram"},
		{"---\ntitle: Login\n---\nsequenceDiagram\n  A->>B: x", "sequenceDiagram"},
		{"\n%% comment\n%%{init: {}}%%\nflowchart LR\n  A-->B", "flowchart"},
		{"  graph TD\n  A-->B", "graph"},
		{"stateDiagram-v2\n  [*] --> A", "stateDiagram-v2"},
		{"%% only a comment", ""},
	}
	for _, tt := range tests {
		if got := MermaidType(tt.src); got != tt.want {
			t.Errorf("MermaidType(%q) = %q, want %q", tt.src, got, tt.want)
		}
	}
}

// TestGolden renders every fixture at a width large enough not to trigger
// fitting, so the goldens show the library output for each construct.
func TestGolden(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		ext := filepath.Ext(file)
		if ext == ".golden" {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(file), ext)
		t.Run(name, func(t *testing.T) {
			lang := "mermaid"
			if ext == ".puml" {
				lang = "plantuml"
			}
			res := Render(lang, readFixture(t, filepath.Base(file)), 200)
			golden.Assert(t, filepath.Join("testdata", name+".golden"), lines(res))
		})
	}
}

func TestSupportedTypesAreRendered(t *testing.T) {
	for _, name := range []string{"seq-notes.mmd", "seq-blocks.mmd", "seq-activation.mmd", "flow-lr.mmd", "flow-td.mmd", "er.mmd", "title.mmd"} {
		res := Render("mermaid", readFixture(t, name), 200)
		if res.Fallback {
			t.Errorf("%s fell back to source:\n%s", name, lines(res))
		}
	}
}

func TestTitleIsPrintedAbove(t *testing.T) {
	res := Render("mermaid", readFixture(t, "title.mmd"), 200)
	if res.Lines[0].String() != "Login flow" {
		t.Errorf("first line %q, want the title", res.Lines[0].String())
	}
}

func TestFallbackLabels(t *testing.T) {
	tests := []struct {
		lang, file string
		want       []string
	}{
		{"mermaid", "state.mmd", []string{"mermaid", "stateDiagram-v2", "not supported"}},
		{"plantuml", "plantuml.puml", []string{"plantuml"}},
		{"mermaid", "seq-invalid.mmd", []string{"mermaid", "sequenceDiagram", "invalid syntax"}},
	}
	for _, tt := range tests {
		src := readFixture(t, tt.file)
		res := Render(tt.lang, src, 200)
		if !res.Fallback {
			t.Errorf("%s did not fall back", tt.file)
			continue
		}
		label := res.Lines[0].String()
		for _, w := range tt.want {
			if !strings.Contains(label, w) {
				t.Errorf("%s: label %q lacks %q", tt.file, label, w)
			}
		}
		body := lines(res)
		for l := range strings.SplitSeq(strings.TrimSpace(src), "\n") {
			if !strings.Contains(body, l) {
				t.Errorf("%s: frame lacks source line %q", tt.file, l)
			}
		}
	}
}

func TestFrameIsRectangular(t *testing.T) {
	lines := Frame("mermaid - x", "short\na much longer line with 日本語\n\ttab")
	w := text.Width(lines[0])
	for _, l := range lines {
		if text.Width(l) != w {
			t.Errorf("frame line %q is %d wide, want %d", l, text.Width(l), w)
		}
	}
}

func TestPanicBecomesFallback(t *testing.T) {
	orig := renderDiagram
	defer func() { renderDiagram = orig }()
	renderDiagram = func(string, *mdiagram.Config) (string, mrender.WidthStatus, error) {
		panic("boom")
	}
	res := Render("mermaid", "sequenceDiagram\n  A->>B: hi", 80)
	if !res.Fallback || !strings.Contains(res.Lines[0].String(), "renderer failed: boom") {
		t.Errorf("panic not turned into framed source:\n%s", lines(res))
	}
}

func width(res Result) int {
	w := 0
	for _, l := range res.Lines {
		w = max(w, l.Width())
	}
	return w
}

// renderAt renders src with the library directly, bypassing fitting.
func renderAt(t *testing.T, src string, maxWidth int) string {
	t.Helper()
	cfg := mdiagram.DefaultConfig()
	cfg.MaxWidth = maxWidth
	out, err := safeRender(src, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestFlowchartWidthFitting(t *testing.T) {
	src := readFixture(t, "flow-lr.mmd")
	full := blockWidth(renderAt(t, src, 0))
	compact := blockWidth(renderAt(t, src, 1))
	vertical, _ := toTopDown(src)
	if !(compact < full) {
		t.Fatalf("fixture does not compact: full %d, compact %d", full, compact)
	}

	t.Run("fits with compact spacing", func(t *testing.T) {
		w := compact
		res := Render("mermaid", src, w)
		if res.Wide || width(res) > w {
			t.Errorf("width %d: got %d wide, wide=%v", w, width(res), res.Wide)
		}
		if lines(res) != renderAt(t, src, w)+"\n" {
			t.Errorf("width %d: layout is not the compacted LR one", w)
		}
	})
	t.Run("re-laid out top-to-bottom", func(t *testing.T) {
		w := compact - 1
		res := Render("mermaid", src, w)
		if lines(res) != renderAt(t, vertical, w)+"\n" {
			t.Errorf("width %d: expected the TD layout, got\n%s", w, lines(res))
		}
	})
	t.Run("still too wide", func(t *testing.T) {
		res := Render("mermaid", src, 5)
		if !res.Wide || res.Fallback {
			t.Errorf("wide=%v fallback=%v, want a wide diagram", res.Wide, res.Fallback)
		}
	})
	t.Run("vertical flowchart is not rewritten", func(t *testing.T) {
		if _, ok := toTopDown(readFixture(t, "flow-td.mmd")); ok {
			t.Error("TD flowchart rewritten")
		}
	})
}

func TestWideSequenceDiagram(t *testing.T) {
	var b strings.Builder
	b.WriteString("sequenceDiagram\n")
	for i := range 8 {
		fmt.Fprintf(&b, "    participant P%c as Participant%c\n", 'A'+i, 'A'+i)
	}
	b.WriteString("    PA->>PH: hello\n")
	res := Render("mermaid", b.String(), 80)
	if !res.Wide || res.Fallback || width(res) <= 80 {
		t.Errorf("wide=%v fallback=%v width=%d", res.Wide, res.Fallback, width(res))
	}
}
