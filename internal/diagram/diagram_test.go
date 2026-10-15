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
	for _, name := range []string{"seq-notes.mmd", "seq-blocks.mmd", "seq-activation.mmd", "flow-lr.mmd", "flow-td.mmd", "er.mmd", "title.mmd", "mindmap.mmd", "plantuml.puml", "plantuml-seq.puml", "plantuml-mindmap.puml"} {
		lang := "mermaid"
		if filepath.Ext(name) == ".puml" {
			lang = "plantuml"
		}
		res := Render(lang, readFixture(t, name), 200)
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
		{"plantuml", "plantuml-class.puml", []string{"plantuml", "class", "not supported"}},
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

func TestMindmap(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			"tree",
			"mindmap\n  root((Plan))\n    Goals\n      Faster\n      Simpler\n    Risks",
			[]string{"Plan", "├── Goals", "│   ├── Faster", "│   └── Simpler", "└── Risks"},
		},
		{
			"shapes",
			"mindmap\n  root((A))\n    b[Square]\n    c(Rounded)\n    d))Bang((\n    e)Cloud(\n    f{{Hexagon}}",
			[]string{"A", "├── Square", "├── Rounded", "├── Bang", "├── Cloud", "└── Hexagon"},
		},
		{
			"icons and classes",
			"mindmap\n  R\n    Child\n    ::icon(fa fa-book)\n    :::urgent large",
			[]string{"R", "└── Child"},
		},
		{"markdown string", "mindmap\n  root[\"`**Root** text`\"]", []string{"Root text"}},
		{"markdown string with a line break", "mindmap\n  root[\"`First line\nsecond line`\"]", []string{"First line", "second line"}},
		{
			"line break in a node",
			"mindmap\n  R\n    On effectiveness<br/>and features\n    Next",
			[]string{"R", "├── On effectiveness", "│   and features", "└── Next"},
		},
		{
			"nearest node with less indentation",
			"mindmap\n  R\n    A\n        B\n      C",
			[]string{"R", "└── A", "    ├── B", "    └── C"},
		},
		{"comments", "mindmap\n  %% a comment\n  R\n    A", []string{"R", "└── A"}},
		{"title", "---\ntitle: Ideas\n---\nmindmap\n  R\n    A", []string{"Ideas", "", "R", "└── A"}},
		// A tab counts as 4 spaces, so A and B are siblings.
		{"tabs", "mindmap\n\tR\n\t\tA\n        B", []string{"R", "├── A", "└── B"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Render("mermaid", tt.src, 80)
			if res.Fallback {
				t.Fatalf("fell back:\n%s", lines(res))
			}
			if got, want := lines(res), strings.Join(tt.want, "\n")+"\n"; got != want {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
	errs := []struct{ name, src, reason string }{
		{"more than one root", "mindmap\n  A\n  B", "more than one root"},
		{"unclosed shape", "mindmap\n  root((Plan)", "unclosed shape"},
		{"unclosed markdown string", "mindmap\n  root[\"`Plan\n    A", "unclosed markdown string"},
		{"no nodes", "mindmap", "no nodes"},
	}
	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			res := Render("mermaid", tt.src, 80)
			if !res.Fallback {
				t.Fatalf("did not fall back:\n%s", lines(res))
			}
			if label, want := res.Lines[0].String(), "mermaid - mindmap - "+tt.reason; !strings.Contains(label, want) {
				t.Errorf("label %q lacks %q", label, want)
			}
		})
	}
}

func TestMindmapWidth(t *testing.T) {
	t.Run("long text", func(t *testing.T) {
		res := Render("mermaid", "mindmap\n  R\n    Short\n    "+strings.Repeat("word ", 8), 30)
		got := strings.Split(strings.TrimSuffix(lines(res), "\n"), "\n")
		if len(got) < 4 || got[1] != "├── Short" || !strings.HasPrefix(got[2], "└── word") {
			t.Fatalf("text did not wrap: %q", got)
		}
		for _, l := range got[3:] {
			if !strings.HasPrefix(l, "    word") {
				t.Errorf("continuation line %q", l)
			}
		}
		if res.Wide || width(res) > 30 {
			t.Errorf("wider than 30: %q", got)
		}
	})
	t.Run("deep tree", func(t *testing.T) {
		src := "mindmap\n  R\n    A\n      B\n        C\n          D\n            " + strings.Repeat("abc ", 7)
		res := Render("mermaid", src, 20)
		got := strings.Split(strings.TrimSuffix(lines(res), "\n"), "\n")
		node := got[5:]
		if len(node) < 3 || !strings.HasSuffix(got[4], "└── D") {
			t.Fatalf("text at depth 5 did not wrap: %q", got)
		}
		for _, l := range node {
			if w := text.Width(l); w <= 20 || w > 30 {
				t.Errorf("line %q is %d wide, want text of at most 10 columns after 20", l, w)
			}
		}
		if !res.Wide {
			t.Error("not wide")
		}
	})
	t.Run("long word", func(t *testing.T) {
		word := strings.Repeat("x", 36)
		res := Render("mermaid", "mindmap\n  R\n    "+word, 30)
		if got := lines(res); got != "R\n└── "+word+"\n" || !res.Wide {
			t.Errorf("got %q, wide %v", got, res.Wide)
		}
	})
}

func TestPlantumlType(t *testing.T) {
	tests := []struct{ src, want string }{
		{"@startuml\nAlice -> Bob: hello\n@enduml", "sequence"},
		{"Alice -> Bob: hello", "sequence"},
		{"participant A", "sequence"},
		{"@startuml\nclass User\n@enduml", "class"},
		{"interface I", "class"},
		{"enum E", "class"},
		{"abstract A", "class"},
		{"usecase U", "usecase"},
		{"state S", "state"},
		{"[*] --> A", "state"},
		{"component C", "component"},
		{"object O", "object"},
		{"node N", "deployment"},
		{"start\n:step;\nstop", "activity"},
		{":step;", "activity"},
		{"@startmindmap\n* Root\n@endmindmap", "mindmap"},
		{"@startgantt\n[Task] lasts 5 days\n@endgantt", "gantt"},
		{"@startwbs\n* Root\n@endwbs", "wbs"},
		{"skinparam monochrome true", ""},
	}
	for _, tt := range tests {
		if got := plantumlType(tt.src); got != tt.want {
			t.Errorf("plantumlType(%q) = %q, want %q", tt.src, got, tt.want)
		}
	}
}

func TestPlantumlSequence(t *testing.T) {
	tests := []struct{ name, puml, mermaid string }{
		{
			"messages",
			"Alice -> Bob : hello\nBob --> Alice : ok\nAlice ->> Bob : async\nAlice -> Alice : self",
			"Alice->>Bob: hello\nBob-->>Alice: ok\nAlice-)Bob: async\nAlice->>Alice: self",
		},
		{
			"participants",
			"actor User\ndatabase \"Main DB\" as DB\nUser -> DB : query",
			"actor User\nparticipant DB as Main DB\nUser->>DB: query",
		},
		{
			"reversed and decorated arrows",
			"participant Alice\nparticipant Bob\nBob <- Alice : hi\nAlice ->x Bob : lost\nAlice -[#red]> Bob : red\nAlice <-> Bob : both",
			"participant Alice\nparticipant Bob\nAlice->>Bob: hi\nAlice->>Bob: lost\nAlice->>Bob: red\nAlice->>Bob: both",
		},
		{
			"blocks",
			"A -> B : start\nalt valid\nA -> B : ok\nelse invalid\nA -> B : no\nend\nloop 3 times\nA -> B : ping\nend\npar first\nA -> B : x\nelse second\nB -> A : y\nend",
			"A->>B: start\nalt valid\nA->>B: ok\nelse invalid\nA->>B: no\nend\nloop 3 times\nA->>B: ping\nend\npar first\nA->>B: x\nand second\nB->>A: y\nend",
		},
		{
			"group, divider and reference",
			"participant A\nparticipant B\nparticipant C\ngroup Retry\nA -> B : try\nend\n== Phase 2 ==\nref over A, B : Login",
			"participant A\nparticipant B\nparticipant C\nnote over A,C: Retry\nA->>B: try\nnote over A,C: Phase 2\nnote over A,B: ref Login",
		},
		{
			"multi-line note",
			"A -> B : x\nnote right of B\nfirst\nsecond\nend note",
			"A->>B: x\nnote right of B: first second",
		},
		{
			"notes attached to a message",
			"participant A\nparticipant B\nparticipant C\nA -> B : call\nnote left : sent\nnote right : received\nnote across : all",
			"participant A\nparticipant B\nparticipant C\nA->>B: call\nnote left of A: sent\nnote right of B: received\nnote over A,C: all",
		},
		{"activation and return", "A -> B ++ : call\nreturn ok", "A->>+B: call\nB-->>-A: ok"},
		{
			"ignored lines",
			"skinparam monochrome true\nskinparam participant {\nBorderColor red\n}\nhide footbox\n' a comment\nheader\nPage 1\nendheader\nA -> B : x\n...\n|||\n!theme plain\nB -> A : y",
			"A->>B: x\nB->>A: y",
		},
		{"title and line breaks", "title Login\nA -> B : first\\nsecond", "---\ntitle: Login\n---\nsequenceDiagram\nA->>B: first second"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Render("plantuml", "@startuml\n"+tt.puml+"\n@enduml", 80)
			if res.Fallback {
				t.Fatalf("fell back:\n%s", lines(res))
			}
			mermaid := tt.mermaid
			if !strings.HasPrefix(mermaid, "---") {
				mermaid = "sequenceDiagram\n" + mermaid
			}
			if got, want := lines(res), lines(Render("mermaid", mermaid, 80)); got != want {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
	t.Run("unsupported line", func(t *testing.T) {
		res := Render("plantuml", "@startuml\nA -> B : call\n[-> A : in\n@enduml", 80)
		if !res.Fallback || !strings.Contains(res.Lines[0].String(), "plantuml - sequence - unsupported line 3") {
			t.Errorf("got\n%s", lines(res))
		}
	})
}

func TestPlantumlMindmap(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			"stars",
			"* Debian\n** Ubuntu\n*** Linux Mint\n*** Kubuntu\n** LMDE",
			[]string{"Debian", "├── Ubuntu", "│   ├── Linux Mint", "│   └── Kubuntu", "└── LMDE"},
		},
		{"sides", "+ OS\n++ Linux\n-- Windows", []string{"OS", "├── Linux", "└── Windows"}},
		{
			"node decorations",
			"* Root\n**[#Orange] Colored\n**_ Boxless\n**:First\nsecond;",
			[]string{"Root", "├── Colored", "├── Boxless", "└── First", "    second"},
		},
		{"title", "title Ideas\n* Root", []string{"Ideas", "", "Root"}},
		{"line break", "* Root\n** a\\nb", []string{"Root", "└── a", "    b"}},
		{"left side and comments", "' a comment\n* Root\n** A\nleft side\n** B", []string{"Root", "├── A", "└── B"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Render("plantuml", "@startmindmap\n"+tt.src+"\n@endmindmap", 80)
			if res.Fallback {
				t.Fatalf("fell back:\n%s", lines(res))
			}
			if got, want := lines(res), strings.Join(tt.want, "\n")+"\n"; got != want {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
	errs := []struct{ name, src, reason string }{
		{"more than one root", "* A\n* B", "more than one root"},
		{"unclosed node", "* Root\n**:open", "unclosed node"},
		{"unsupported line", "* Root\nplain text", "unsupported line 3"},
		{"no nodes", "", "no nodes"},
	}
	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			res := Render("plantuml", "@startmindmap\n"+tt.src+"\n@endmindmap", 80)
			if label, want := res.Lines[0].String(), "plantuml - mindmap - "+tt.reason; !res.Fallback || !strings.Contains(label, want) {
				t.Errorf("label %q lacks %q", label, want)
			}
		})
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
