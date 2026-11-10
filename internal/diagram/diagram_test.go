package diagram

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	for _, name := range []string{"seq-notes.mmd", "seq-blocks.mmd", "seq-activation.mmd", "flow-lr.mmd", "flow-td.mmd", "er.mmd", "title.mmd", "mindmap.mmd", "gantt.mmd", "gantt-basic.mmd", "plantuml.puml", "plantuml-seq.puml", "plantuml-mindmap.puml", "state.mmd", "state-composite.mmd", "plantuml-state.puml"} {
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
		{"mermaid", "class.mmd", []string{"mermaid", "classDiagram", "not supported"}},
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

func TestPlantumlState(t *testing.T) {
	puml := func(lines ...string) string { return "@startuml\n" + strings.Join(lines, "\n") + "\n@enduml" }
	state := func(lines ...string) string { return "stateDiagram-v2\n    " + strings.Join(lines, "\n    ") }
	tests := []struct {
		name        string
		plant, want string
	}{
		{
			"transitions and arrows",
			puml("[*] --> Idle", "Idle -> Loading : open", "Loading -down-> Ready", "Ready -[#red]-> Done", "Done <- Loading", "Ready -r-> Done", "Done --> [*]"),
			state("[*] --> Idle", "Idle --> Loading : open", "Loading --> Ready", "Ready --> Done", "Loading --> Done", "Ready --> Done", "Done --> [*]"),
		},
		{
			"declarations",
			puml(`state "Waiting" as W`, `state L as "Loading data"`, "state X #pink", "W : press a key", "W --> L", "L --> X"),
			state(`state "Waiting" as W`, `state "Loading data" as L`, "state X", "W : press a key", "W --> L", "L --> X"),
		},
		{
			"composite state with regions",
			puml("[*] --> Active", "state Active #pink {", "[*] -> A", "||", "[*] -> B", "}"),
			state("[*] --> Active", "state Active {", "[*] --> A", "--", "[*] --> B", "}"),
		},
		{
			"special states",
			puml("state s <<start>>", "state c <<choice>>", "state p <<entryPoint>>", "state e <<end>>", "s --> c", "c --> p", "p --> e"),
			state(`state "●" as s`, "state c <<choice>>", "state p", `state "◉" as e`, "s --> c", "c --> p", "p --> e"),
		},
		{
			"history",
			puml("state Active {", "[*] --> Running", "Running --> [H]", "}", "Idle --> Active[H]"),
			state("state Active {", "[*] --> Running", `state "H" as hist`, "Running --> hist", "}", "Idle --> hist"),
		},
		{
			"line breaks",
			puml(`state "Long\nname" as L`, `L --> M : a\nb`, `note right of L : x\ny`),
			state(`state "Long<br>name" as L`, "L --> M : a b", "note right of L : x y"),
		},
	}
	for _, tt := range tests {
		sameRender(t, tt.name, "plantuml", tt.plant, "mermaid", tt.want, 80)
	}

	got := drawn(t, "plantuml", puml("[*] --> A", "A --> B", "note on link : retry", "note top of A : first", "note bottom of B", "waits", "end note"), 80)
	if want := drawn(t, "mermaid", state("[*] --> A", "A --> B"), 80) + "\n\nA --> B: retry\nA: first\nB: waits"; got != want {
		t.Errorf("notes:\n%s\nwant\n%s", got, want)
	}

	got = drawn(t, "plantuml", puml("title Player", "left to right direction", "skinparam monochrome true", "skinparam state {", "BackgroundColor red", "}",
		"hide empty description", "show stereotype", "scale 2", "' a comment", "/' a block", "comment '/", "header h", "footer f", "caption c",
		"legend", "text", "endlegend", "!theme plain", "[*] --> A", "A --> B"), 80)
	if want := "Player\n\n" + drawn(t, "mermaid", state("direction LR", "[*] --> A", "A --> B"), 80); got != want {
		t.Errorf("direction, title and ignored lines:\n%s\nwant\n%s", got, want)
	}

	errors := []struct{ src, want string }{
		{puml("[*] --> A", `note "free" as N1`), "plantuml - state - unsupported line 3"},
		{puml("state Active {", "[*] --> Running"), "plantuml - state - unclosed block on line 2"},
	}
	for _, tt := range errors {
		res := Render("plantuml", tt.src, 80)
		if !res.Fallback || !strings.Contains(res.Lines[0].String(), tt.want) {
			t.Errorf("%q: got\n%s\nwant a frame labelled %q", tt.src, lines(res), tt.want)
		}
	}
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

// sameRender reports whether two diagrams are drawn as the same lines at
// width, and neither falls back to its source.
func sameRender(t *testing.T, name, gotLang, got, wantLang, want string, width int) {
	t.Helper()
	g, w := Render(gotLang, got, width), Render(wantLang, want, width)
	if g.Fallback || w.Fallback {
		t.Errorf("%s fell back to source:\n%s\nwant\n%s", name, lines(g), lines(w))
		return
	}
	if lines(g) != lines(w) {
		t.Errorf("%s:\n%s\nwant\n%s", name, lines(g), lines(w))
	}
}

func TestStateDiagram(t *testing.T) {
	tests := []struct{ name, state, flowchart string }{
		{
			"transitions",
			"stateDiagram-v2\n    [*] --> Idle\n    Idle --> Loading : open\n    Loading --> [*]",
			"flowchart TB\n    a[\"●\"]\n    b[\"Idle\"]\n    c[\"Loading\"]\n    d[\"◉\"]\n    a --> b\n    b -->|open| c\n    c --> d",
		},
		{
			"descriptions",
			"stateDiagram-v2\n    state \"Waiting\" as W\n    W : press a key\n    L : Loading data\n    W --> L",
			"flowchart TB\n    a[\"Waiting<br>press a key\"]\n    b[\"Loading data\"]\n    a --> b",
		},
		{
			"choice, fork and join",
			"stateDiagram-v2\n    state c <<choice>>\n    state f [[fork]]\n    state j <<join>>\n    [*] --> c\n    c --> f : ok\n    f --> j\n    j --> [*]",
			"flowchart TB\n    s[\"●\"]\n    c[\"◇\"]\n    f[\"━━━\"]\n    j[\"━━━\"]\n    e[\"◉\"]\n    s --> c\n    c -->|ok| f\n    f --> j\n    j --> e",
		},
	}
	for _, tt := range tests {
		sameRender(t, tt.name, "mermaid", tt.state, "mermaid", tt.flowchart, 80)
	}

	plain := "stateDiagram-v2\n    [*] --> Idle\n    Idle --> Done"
	ignored := "stateDiagram-v2\n    %% a comment\n    classDef warn fill:#f00\n    class Idle warn\n    style Idle fill:#f00\n" +
		"    accTitle: States\n    accDescr: Some states\n    accDescr {\n        text\n    }\n    [*] --> Idle:::warn\n    Idle --> Done"
	sameRender(t, "ignored lines", "mermaid", ignored, "mermaid", plain, 80)
	sameRender(t, "stateDiagram", "mermaid", strings.Replace(plain, "-v2", "", 1), "mermaid", plain, 80)

	errors := []struct{ src, want string }{
		{"stateDiagram-v2\n    [*] --> A\n    A -> B", "mermaid - stateDiagram-v2 - unsupported line 3"},
		{"stateDiagram-v2\n    A --> B\n    }", "mermaid - stateDiagram-v2 - unsupported line 3"},
		{"stateDiagram-v2\n    A --> B\n    --", "mermaid - stateDiagram-v2 - unsupported line 3"},
		{"stateDiagram-v2\n    state Active {\n    [*] --> Running", "mermaid - stateDiagram-v2 - unclosed block on line 2"},
		{"stateDiagram-v2\n    A --> B\n    note left of A\n    text", "mermaid - stateDiagram-v2 - unclosed block on line 3"},
		{"stateDiagram", "mermaid - stateDiagram - no states"},
	}
	for _, tt := range errors {
		res := Render("mermaid", tt.src, 80)
		if !res.Fallback || !strings.Contains(res.Lines[0].String(), tt.want) {
			t.Errorf("%q: got\n%s\nwant a frame labelled %q", tt.src, lines(res), tt.want)
		}
	}
}

// drawn returns the lines of a diagram rendered at width, without the
// trailing newline of lines.
func drawn(t *testing.T, lang, src string, width int) string {
	t.Helper()
	res := Render(lang, src, width)
	if res.Fallback {
		t.Fatalf("fell back to source:\n%s", lines(res))
	}
	return strings.TrimSuffix(lines(res), "\n")
}

func TestStateLayout(t *testing.T) {
	check := func(t *testing.T, name, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("%s:\n%s\nwant\n%s", name, got, want)
		}
	}
	state := func(lines ...string) string { return "stateDiagram-v2\n    " + strings.Join(lines, "\n    ") }
	flow := func(lines ...string) string { return "flowchart " + strings.Join(lines, "\n    ") }

	t.Run("composite state", func(t *testing.T) {
		got := drawn(t, "mermaid", state("[*] --> Idle", "Idle --> Active : start", "state Active {", "[*] --> Running",
			"Running --> Paused : pause", "Paused --> Running : resume", "}", "Active --> [*] : stop"), 80)
		top := drawn(t, "mermaid", flow("TB", `s["●"]`, `i["Idle"]`, `a["Active"]`, `e["◉"]`, "s --> i", "i -->|start| a", "a -->|stop| e"), 80)
		inner := drawn(t, "mermaid", flow("TB", `s["●"]`, `r["Running"]`, `p["Paused"]`, "s --> r", "r -->|pause| p", "p -->|resume| r"), 80)
		check(t, "composite state", got, top+"\n\nActive\n\n"+inner)
	})
	t.Run("transition into a composite state", func(t *testing.T) {
		got := drawn(t, "mermaid", state("[*] --> Idle", "state Active {", "[*] --> Running", "}", "Idle --> Running : resume"), 80)
		top := drawn(t, "mermaid", flow("TB", `s["●"]`, `i["Idle"]`, `a["Active"]`, "s --> i", "i -->|resume| a"), 80)
		inner := drawn(t, "mermaid", flow("TB", `s["●"]`, `r["Running"]`, "s --> r"), 80)
		check(t, "transition into a composite state", got, top+"\n\nActive\n\n"+inner)
	})
	t.Run("concurrent regions", func(t *testing.T) {
		got := drawn(t, "mermaid", state("[*] --> Active", "state Active {", "[*] --> A", "--", "[*] --> B", "}"), 80)
		top := drawn(t, "mermaid", flow("TB", `s["●"]`, `a["Active"]`, "s --> a"), 80)
		a := drawn(t, "mermaid", flow("TB", `s["●"]`, `a["A"]`, "s --> a"), 80)
		b := drawn(t, "mermaid", flow("TB", `s["●"]`, `b["B"]`, "s --> b"), 80)
		divider := strings.Repeat("┈", max(blockWidth(a), blockWidth(b)))
		check(t, "concurrent regions", got, top+"\n\nActive\n\n"+a+"\n\n"+divider+"\n\n"+b)
	})
	t.Run("nested composite states", func(t *testing.T) {
		got := drawn(t, "mermaid", state("state A {", "state B {", "x", "}", "}", "state C {", "y", "}"), 80)
		var headers []string
		for l := range strings.SplitSeq(got, "\n") {
			if l == "A" || l == "B" || l == "C" {
				headers = append(headers, l)
			}
		}
		check(t, "nested composite states", strings.Join(headers, " "), "A B C")
	})
	t.Run("direction", func(t *testing.T) {
		got := drawn(t, "mermaid", state("direction LR", "[*] --> A", "A --> B"), 80)
		check(t, "direction", got, drawn(t, "mermaid", flow("LR", `s["●"]`, `a["A"]`, `b["B"]`, "s --> a", "a --> b"), 80))
	})
	t.Run("notes", func(t *testing.T) {
		got := drawn(t, "mermaid", state("[*] --> Idle", "note right of Idle : waits for input", "note left of Idle", "first", "second", "end note"), 80)
		top := drawn(t, "mermaid", flow("TB", `s["●"]`, `i["Idle"]`, "s --> i"), 80)
		check(t, "notes", got, top+"\n\nIdle: waits for input\nIdle: first second")
	})
	t.Run("long note", func(t *testing.T) {
		res := Render("mermaid", state("[*] --> A", "note right of A : this note has many words and wraps at the width"), 30)
		var notes []string
		for _, l := range res.Lines {
			if s := l.String(); strings.HasPrefix(s, "A: ") || len(notes) > 0 {
				notes = append(notes, s)
			}
		}
		if len(notes) < 2 || res.Wide {
			t.Fatalf("note not wrapped, wide=%v:\n%s", res.Wide, lines(res))
		}
		for _, l := range notes[1:] {
			if !strings.HasPrefix(l, "  ") || strings.HasPrefix(l, "   ") || text.Width(l) > 30 {
				t.Errorf("continuation line %q", l)
			}
		}
	})
	t.Run("links in the text of a state", func(t *testing.T) {
		got := drawn(t, "mermaid", state(`state "Save & exit" as S`, `state "a --> b" as T`, "S --> T : x --> y"), 80)
		for _, w := range []string{"│ Save & exit │", "a -‑› b", "x --> y"} {
			if !strings.Contains(got, w) {
				t.Errorf("output lacks %q:\n%s", w, got)
			}
		}
	})
	t.Run("title", func(t *testing.T) {
		got := drawn(t, "mermaid", "---\ntitle: Player\n---\n"+state("[*] --> A"), 80)
		check(t, "title", got, "Player\n\n"+drawn(t, "mermaid", flow("TB", `s["●"]`, `a["A"]`, "s --> a"), 80))
	})
	t.Run("horizontal state diagram re-laid out", func(t *testing.T) {
		src := state("direction LR", "[*] --> First", "First --> Second", "Second --> Third", "Third --> [*]")
		nodes := []string{`s["●"]`, `f["First"]`, `n["Second"]`, `t["Third"]`, `e["◉"]`, "s --> f", "f --> n", "n --> t", "t --> e"}
		check(t, "re-laid out", drawn(t, "mermaid", src, 30), drawn(t, "mermaid", flow(append([]string{"TD"}, nodes...)...), 30))
		if res := Render("mermaid", src, 5); !res.Wide || res.Fallback {
			t.Errorf("wide=%v fallback=%v, want a wide diagram", res.Wide, res.Fallback)
		}
	})
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

func TestFlowchartQuotedAmpersand(t *testing.T) {
	res := Render("mermaid", "flowchart LR\n    a[\"Tom & Jerry\"] --> b", 200)
	if !strings.Contains(lines(res), "│ Tom & Jerry ├") {
		t.Errorf("quoted label split at &:\n%s", lines(res))
	}
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

// at parses a UTC time written as 2006-01-02, 2006-01-02 15:04 or
// 2006-01-02 15:04:05.
func at(t *testing.T, s string) time.Time {
	t.Helper()
	for _, layout := range []string{time.DateOnly, "2006-01-02 15:04", time.DateTime} {
		if v, err := time.Parse(layout, s); err == nil {
			return v
		}
	}
	t.Fatalf("bad time %q", s)
	return time.Time{}
}

func TestGanttDates(t *testing.T) {
	// Every task is "start, end", or "start, end, bar" when its bar ends
	// before its end.
	tests := []struct {
		name  string
		src   string
		tasks []string
	}{
		{"task without a start", "A :a, 2024-01-01, 10d\nB :10d", []string{"2024-01-01, 2024-01-11", "2024-01-11, 2024-01-21"}},
		{"after several tasks", "A :a, 2024-01-01, 5d\nB :b, 2024-01-01, 10d\nC :after a b, 10d", []string{"2024-01-01, 2024-01-06", "2024-01-01, 2024-01-11", "2024-01-11, 2024-01-21"}},
		{"until a later task", "A :a, 2024-01-01, until b\nB :b, 2024-01-11, 10d", []string{"2024-01-01, 2024-01-11", "2024-01-11, 2024-01-21"}},
		{"until the earliest task", "A :2024-01-01, until b c\nB :b, 2024-01-10, 1d\nC :c, 2024-01-05, 1d", []string{"2024-01-01, 2024-01-05", "2024-01-10, 2024-01-11", "2024-01-05, 2024-01-06"}},
		{"weeks", "A :2024-01-01, 1w", []string{"2024-01-01, 2024-01-08"}},
		{"days rounded", "A :2024-01-01, 1.5d", []string{"2024-01-01, 2024-01-03"}},
		{"month to a shorter month", "A :2024-01-31, 1M", []string{"2024-01-31, 2024-02-29"}},
		{"fraction of a month", "A :2024-01-31, 1.5M", []string{"2024-01-31, 2024-02-29"}},
		{"year from a leap day", "A :2024-02-29, 1y", []string{"2024-02-29, 2025-02-28"}},
		{"hours", "A :2024-01-01, 36h", []string{"2024-01-01, 2024-01-02 12:00"}},
		{"minutes, seconds and milliseconds", "A :a, 2024-01-01, 90m\nB :30s\nC :1500ms", []string{"2024-01-01, 2024-01-01 01:30", "2024-01-01 01:30, 2024-01-01 01:30:30", "2024-01-01 01:30:30, 2024-01-01 01:30:31.5"}},
		{"times", "dateFormat HH:mm\nA :a, 17:49, 10m\nB :5m", []string{"1970-01-01 17:49, 1970-01-01 17:59", "1970-01-01 17:59, 1970-01-01 18:04"}},
		{"unix seconds", "dateFormat X\nA :0, 71", []string{"1970-01-01, 1970-01-01 00:01:11"}},
		{"unix milliseconds", "dateFormat x\nA :1000, 3000", []string{"1970-01-01 00:00:01, 1970-01-01 00:00:03"}},
		{"dateFormat after the tasks", "A :01.02.2024, 1d\ndateFormat DD.MM.YYYY", []string{"2024-02-01, 2024-02-02"}},
		{"short tokens", "dateFormat D/M/YY\nA :a, 5/1/24, 1d\nB :b, 1/1/69, 1d\nC :12/11/68, 1d", []string{"2024-01-05, 2024-01-06", "1969-01-01, 1969-01-02", "2068-11-12, 2068-11-13"}},
		{"end date", "A :2024-01-01, 2024-01-10", []string{"2024-01-01, 2024-01-10"}},
		{"inclusive end dates", "inclusiveEndDates\nA :2024-01-01, 2024-01-10", []string{"2024-01-01, 2024-01-11"}},
		{"invalid end", "A :2024-01-01, 3dX", []string{"2024-01-01, 2024-01-01"}},
		{"excluded weekend", "excludes weekends\nA :a, 2024-01-01, 5d\nB :after a, 1d", []string{"2024-01-01, 2024-01-08, 2024-01-06", "2024-01-08, 2024-01-09"}},
		{"weekend in the middle", "excludes weekends\nA :2024-01-05, 2d", []string{"2024-01-05, 2024-01-09"}},
		{"included day", "excludes weekends\nincludes 2024-01-06\nA :2024-01-01, 6d", []string{"2024-01-01, 2024-01-08, 2024-01-07"}},
		{"weekend friday", "excludes weekends\nweekend friday\nA :2024-01-01, 5d", []string{"2024-01-01, 2024-01-08"}},
		{"weekday name", "excludes Sunday\nA :2024-01-06, 2d", []string{"2024-01-06, 2024-01-09"}},
		{"excluded dates", "dateFormat DD.MM.YYYY\nexcludes 02.01.2024, 2024-01-03\nA :01.01.2024, 2d", []string{"2024-01-01, 2024-01-05"}},
		{"end date not moved", "excludes weekends\nA :2024-01-01, 2024-01-07", []string{"2024-01-01, 2024-01-07"}},
		{"until not moved", "excludes weekends\nA :2024-01-01, until b\nB :b, 2024-01-06, 1d", []string{"2024-01-01, 2024-01-06", "2024-01-06, 2024-01-08, 2024-01-07"}},
		{"vert task before the next one", "A :a, 2024-01-01, 2d\nV :vert, 2024-01-05, 1d\nB :1d", []string{"2024-01-01, 2024-01-03", "2024-01-05, 2024-01-06", "2024-01-06, 2024-01-07"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parseGantt("gantt\n" + tt.src)
			if err != nil {
				t.Fatal(err)
			}
			if len(c.tasks) != len(tt.tasks) {
				t.Fatalf("got %d tasks, want %d", len(c.tasks), len(tt.tasks))
			}
			for i, want := range tt.tasks {
				var ts []time.Time
				for s := range strings.SplitSeq(want, ", ") {
					ts = append(ts, at(t, s))
				}
				if len(ts) == 2 {
					ts = append(ts, ts[1])
				}
				got := c.tasks[i]
				if !got.start.Equal(ts[0]) || !got.end.Equal(ts[1]) || !got.bar.Equal(ts[2]) {
					t.Errorf("task %d: got %v, %v, %v, want %v, %v, %v", i, got.start, got.end, got.bar, ts[0], ts[1], ts[2])
				}
			}
		})
	}
	errs := []struct{ name, src, reason string }{
		{"unsupported line", "gantt\ndateFormat YYYY-MM-DD\nA 2024-01-01, 3d", "unsupported line 3"},
		{"only tags", "gantt\nA :done, crit", "unsupported line 2"},
		{"no items", "gantt\nA :", "unsupported line 2"},
		{"too many items", "gantt\nA :a, 2024-01-01, 3d, x", "unsupported line 2"},
		{"unknown weekend", "gantt\nweekend monday\nA :2024-01-01, 1d", "unsupported line 2"},
		{"invalid date", "gantt\nA :2024-13-01, 3d", "invalid date on line 2"},
		{"no such day", "gantt\nA :2024-02-30, 3d", "invalid date on line 2"},
		{"line in frontmatter numbering", "---\ntitle: x\n---\ngantt\nA :1 Jan, 3d", "invalid date on line 5"},
		{"unsupported dateFormat", "gantt\ndateFormat MMM D YYYY\nA :Jan 1 2024, 3d", "unsupported dateFormat"},
		{"timestamp inside a format", "gantt\ndateFormat X.SSS\nA :1.000, 3d", "unsupported dateFormat"},
		{"no start date", "gantt\nA :3d", "no start date on line 2"},
		{"unknown task in after", "gantt\nA :after x, 3d", "unknown task x"},
		{"unknown task in until", "gantt\nA :2024-01-01, until y", "unknown task y"},
		{"circular dependency", "gantt\nA :a, after b, 1d\nB :b, after a, 1d", "circular dependency"},
		{"no tasks", "gantt\ntitle Plan", "no tasks"},
		{"only a vert task", "gantt\nV :vert, 2024-01-01, 1d", "no tasks"},
		{"every day excluded", "gantt\nexcludes weekends monday tuesday wednesday thursday friday\nA :2024-01-01, 1d", "every day is excluded"},
	}
	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseGantt(tt.src); err == nil || err.Error() != tt.reason {
				t.Errorf("got error %v, want %q", err, tt.reason)
			}
		})
	}
}

func TestGantt(t *testing.T) {
	sp, full := func(n int) string { return strings.Repeat(" ", n) }, func(n int) string { return strings.Repeat("█", n) }
	tests := []struct {
		name  string
		src   string
		width int
		want  []string
	}{
		{
			"bars",
			"gantt\n  dateFormat YYYY-MM-DD\n  section Build\n    Design :d, 2024-01-01, 10d\n    Code :after d, 10d",
			40,
			[]string{sp(10) + "2024-01-01" + sp(10) + "2024-01-21", "Build", "  Design  " + full(15), "  Code" + sp(19) + full(15)},
		},
		{
			"title line over frontmatter",
			"---\ntitle: Old\n---\ngantt\n  title Plan\n  A :2024-01-01, 20d",
			40,
			[]string{"Plan", "", sp(3) + "2024-01-01" + sp(17) + "2024-01-21", "A  " + full(37)},
		},
		{
			"frontmatter title",
			"---\ntitle: Ideas\n---\ngantt\n  A :2024-01-01, 20d",
			40,
			[]string{"Ideas", "", sp(3) + "2024-01-01" + sp(17) + "2024-01-21", "A  " + full(37)},
		},
		{
			"task states",
			"gantt\n  Design :done, d, 2024-01-01, 10d\n  Build :active, b, after d, 10d\n  Test :t, after b, 10d\n  Fix :crit, f, after t, 10d\n  Release :milestone, after f, 0d",
			40,
			[]string{
				sp(9) + "2024-01-01" + sp(11) + "2024-02-10",
				"Design   " + strings.Repeat("░", 8),
				"Build    " + sp(8) + strings.Repeat("▓", 8),
				"Test     " + sp(16) + full(7),
				"Fix      " + sp(23) + full(8),
				"Release  " + sp(30) + "◆",
			},
		},
		{
			"active wins over done",
			"gantt\n  A :done, active, 2024-01-01, 20d",
			40,
			[]string{sp(3) + "2024-01-01" + sp(17) + "2024-01-21", "A  " + strings.Repeat("▓", 37)},
		},
		{
			"tasks without a section",
			"gantt\n  Kickoff :k, 2024-01-01, 1d\n  section Work\n  Build :after k, 5d",
			40,
			[]string{sp(9) + "2024-01-01" + sp(11) + "2024-01-07", "Kickoff  " + full(5), "Work", "  Build  " + sp(5) + full(26)},
		},
		{
			"vertical marker",
			"gantt\n  A :a, 2024-01-01, 10d\n  Deadline :vert, 2024-01-05, 0d",
			40,
			[]string{sp(3) + "2024-01-01" + sp(17) + "2024-01-11", "A  " + full(37)},
		},
		{
			"section of vertical markers",
			"gantt\n  A :a, 2024-01-01, 10d\n  section Marks\n  Deadline :vert, 2024-01-05, 0d",
			40,
			[]string{sp(3) + "2024-01-01" + sp(17) + "2024-01-11", "A  " + full(37)},
		},
		{
			"milestone in the middle of its duration",
			"gantt\n  A :a, 2024-01-01, 10d\n  M :milestone, 2024-01-05, 2d",
			40,
			[]string{sp(3) + "2024-01-01" + sp(17) + "2024-01-11", "A  " + full(37), "M  " + sp(19) + "◆"},
		},
		{
			"chart of one instant",
			"gantt\n  M :milestone, 2024-01-05, 0d",
			40,
			[]string{sp(3) + "2024-01-05" + sp(17) + "2024-01-05", "M  ◆"},
		},
		{
			"bar at the end of the chart",
			"gantt\n  A :a, 2024-01-01, 10d\n  B :0d",
			40,
			[]string{sp(3) + "2024-01-01" + sp(17) + "2024-01-11", "A  " + full(37), "B  " + sp(36) + full(1)},
		},
		{
			"excluded weekend between bars",
			"gantt\n  excludes weekends\n  A :a, 2024-01-01, 5d\n  B :after a, 1d",
			40,
			// 37 columns for 8 days: Saturday 2024-01-06 is column 23,
			// Monday 2024-01-08 column 32.
			[]string{sp(3) + "2024-01-01" + sp(17) + "2024-01-09", "A  " + full(23), "B  " + sp(32) + full(5)},
		},
		{
			"axis format",
			"gantt\n  axisFormat %-d %b\n  A :2024-01-01, 20d",
			40,
			[]string{sp(3) + "1 Jan" + sp(26) + "21 Jan", "A  " + full(37)},
		},
		{
			"axis with only the start",
			"gantt\n  axisFormat %Y-%m-%dT%H:%M\n  A :2024-01-01, 5d",
			30,
			[]string{sp(3) + "2024-01-01T00:00", "A  " + full(27)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Render("mermaid", tt.src, tt.width)
			if res.Fallback {
				t.Fatalf("fell back:\n%s", lines(res))
			}
			if got, want := lines(res), strings.Join(tt.want, "\n")+"\n"; got != want {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
	t.Run("ignored lines", func(t *testing.T) {
		tasks := "  A :a, 2024-01-01, 10d\n  B :5d"
		with := "gantt\n  todayMarker off\n  tickInterval 1week\n  weekday monday\n  topAxis\n  accTitle: Plan\n  accDescr: About the plan\n" +
			"  accDescr {\n    A longer\n    description\n  }\n  click a href \"https://example.com\"\n  %% a comment\n" + tasks
		got, want := Render("mermaid", with, 40), Render("mermaid", "gantt\n"+tasks, 40)
		if got.Fallback || lines(got) != lines(want) {
			t.Errorf("got\n%s\nwant\n%s", lines(got), lines(want))
		}
	})
	t.Run("axis codes", func(t *testing.T) {
		at := time.Date(2024, time.March, 5, 14, 7, 9, 45e6, time.UTC)
		f := "%Y %y %m %d %e %H %I %M %S %L %p %a %A %b %B %j %s %% %-m %-H %Q %"
		want := "2024 24 03 05  5 14 02 07 09 045 PM Tue Tuesday Mar March 065 1709647629 % 3 14 %Q %"
		if got := formatAxis(at, f); got != want {
			t.Errorf("got  %q\nwant %q", got, want)
		}
	})
	errs := []struct{ name, src, reason string }{
		{"unsupported line", "gantt\ndateFormat YYYY-MM-DD\nA 2024-01-01, 3d", "unsupported line 3"},
		{"invalid date", "gantt\nA :2024-13-01, 3d", "invalid date on line 2"},
		{"unsupported dateFormat", "gantt\ndateFormat MMM D YYYY\nA :Jan 1 2024, 3d", "unsupported dateFormat"},
		{"no start date", "gantt\nA :3d", "no start date on line 2"},
		{"unknown task", "gantt\nA :after x, 3d", "unknown task x"},
		{"circular dependency", "gantt\nA :a, after b, 1d\nB :b, after a, 1d", "circular dependency"},
		{"no tasks", "gantt\ntitle Plan", "no tasks"},
	}
	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			res := Render("mermaid", tt.src, 80)
			if !res.Fallback {
				t.Fatalf("did not fall back:\n%s", lines(res))
			}
			if label, want := res.Lines[0].String(), "mermaid - gantt - "+tt.reason; !strings.Contains(label, want) {
				t.Errorf("label %q lacks %q", label, want)
			}
		})
	}
}

func TestGanttWidth(t *testing.T) {
	full := func(n int) string { return strings.Repeat("█", n) }
	split := func(res Result) []string { return strings.Split(strings.TrimSuffix(lines(res), "\n"), "\n") }
	t.Run("long names", func(t *testing.T) {
		res := Render("mermaid", "gantt\n  section Work\n  Implement the parser and the tests :2024-01-01, 5d", 40)
		// Two dates do not fit into 20 columns, so the axis has only the
		// start.
		want := []string{strings.Repeat(" ", 20) + "2024-01-01", "Work", "  Implement the     " + full(20), "  parser and the", "  tests"}
		if got := split(res); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
		if res.Wide || width(res) > 40 {
			t.Errorf("wider than 40: wide %v, width %d", res.Wide, width(res))
		}
	})
	t.Run("narrow output", func(t *testing.T) {
		res := Render("mermaid", "gantt\n  section Work\n  Implement the parser :2024-01-01, 5d", 25)
		got := split(res)
		if len(got) < 5 || got[2] != "  Implemen  "+full(20) || got[3] != "  t the" || got[4] != "  parser" {
			t.Errorf("got %q", got)
		}
		if !res.Wide {
			t.Error("not wide")
		}
	})
	t.Run("long word", func(t *testing.T) {
		res := Render("mermaid", "gantt\n  Supercalifragilisticexpialidocious :a, 2024-01-01, 5d\n  B :after a, 5d", 40)
		got := split(res)
		if len(got) != 4 || got[1] != "Supercalifragilist  "+full(10) || got[2] != "icexpialidocious" || got[3] != "B"+strings.Repeat(" ", 29)+full(10) {
			t.Errorf("got %q", got)
		}
	})
	t.Run("long section name", func(t *testing.T) {
		res := Render("mermaid", "gantt\n  section A very long section name that does not fit\n  A :2024-01-01, 5d", 30)
		got := split(res)
		if len(got) != 4 || got[1] != "A very long section name that" || got[2] != "does not fit" {
			t.Errorf("got %q", got)
		}
		if res.Wide {
			t.Error("wide")
		}
	})
}

// foreign returns the non-ASCII characters of out that src does not contain,
// each once.
func foreign(out, src string) string {
	var b strings.Builder
	for _, r := range out {
		if r > 127 && !strings.ContainsRune(src, r) && !strings.ContainsRune(b.String(), r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// drawnASCII is drawn in ASCII mode.
func drawnASCII(t *testing.T, lang, src string, width int) string {
	t.Helper()
	res := RenderASCII(lang, src, width)
	if res.Fallback {
		t.Fatalf("fell back to source:\n%s", lines(res))
	}
	return strings.TrimSuffix(lines(res), "\n")
}

func TestASCIIMode(t *testing.T) {
	state := func(lines ...string) string { return "stateDiagram-v2\n    " + strings.Join(lines, "\n    ") }
	flow := func(lines ...string) string { return "flowchart " + strings.Join(lines, "\n    ") }

	for _, src := range []string{"flowchart TD\n    A --> B", "sequenceDiagram\n    Alice->>Bob: Hi"} {
		got := strings.Split(drawnASCII(t, "mermaid", src, 80), "\n")
		if f := foreign(strings.Join(got, "\n"), ""); f != "" {
			t.Errorf("%q: non-ASCII characters %q", src, f)
		}
		want := strings.Split(drawn(t, "mermaid", src, 80), "\n")
		if len(got) != len(want) {
			t.Errorf("%q: %d lines, want %d", src, len(got), len(want))
			continue
		}
		for i := range want {
			if text.Width(got[i]) != text.Width(want[i]) {
				t.Errorf("%q: line %d is %q, want the width of %q", src, i, got[i], want[i])
			}
		}
	}
	if got := drawnASCII(t, "mermaid", "flowchart TD\n    A --> B", 80); !strings.Contains(got, "+---+") || !strings.Contains(got, "| A |") {
		t.Errorf("boxes are not drawn with +, - and |:\n%s", got)
	}

	t.Run("state symbols", func(t *testing.T) {
		got := drawnASCII(t, "mermaid", state("state c <<choice>>", "state f [[fork]]", "state j <<join>>", "[*] --> c", "c --> f : ok", "f --> j", "j --> [*]"), 80)
		want := drawnASCII(t, "mermaid", flow("TB", `s["*"]`, `c["<>"]`, `f["==="]`, `j["==="]`, `e["(*)"]`, "s --> c", "c -->|ok| f", "f --> j", "j --> e"), 80)
		if got != want {
			t.Errorf("got\n%s\nwant\n%s", got, want)
		}
	})
	t.Run("concurrent regions", func(t *testing.T) {
		got := drawnASCII(t, "mermaid", state("[*] --> Active", "state Active {", "[*] --> A", "--", "[*] --> B", "}"), 80)
		top := drawnASCII(t, "mermaid", flow("TB", `s["*"]`, `a["Active"]`, "s --> a"), 80)
		a := drawnASCII(t, "mermaid", flow("TB", `s["*"]`, `a["A"]`, "s --> a"), 80)
		b := drawnASCII(t, "mermaid", flow("TB", `s["*"]`, `b["B"]`, "s --> b"), 80)
		divider := strings.Repeat(".", max(blockWidth(a), blockWidth(b)))
		if want := top + "\n\nActive\n\n" + a + "\n\n" + divider + "\n\n" + b; got != want {
			t.Errorf("got\n%s\nwant\n%s", got, want)
		}
	})
	t.Run("links in the text of a state", func(t *testing.T) {
		if got := drawnASCII(t, "mermaid", state(`state "a --> b" as T`, "[*] --> T"), 80); !strings.Contains(got, "| a - - > b |") {
			t.Errorf("output lacks a - - > b:\n%s", got)
		}
	})
	t.Run("mindmap", func(t *testing.T) {
		got := drawnASCII(t, "mermaid", "mindmap\n  root((Plan))\n    Goals\n      Faster\n      Simpler\n    Risks", 80)
		if want := "Plan\n|-- Goals\n|   |-- Faster\n|   `-- Simpler\n`-- Risks"; got != want {
			t.Errorf("got\n%s\nwant\n%s", got, want)
		}
	})
	t.Run("gantt", func(t *testing.T) {
		src := "gantt\n  Design :done, d, 2024-01-01, 10d\n  Build :active, b, after d, 10d\n  Test :t, after b, 10d\n  Release :milestone, after t, 0d"
		got := drawnASCII(t, "mermaid", src, 40)
		if want := strings.NewReplacer("░", "-", "▓", "=", "█", "#", "◆", "*").Replace(drawn(t, "mermaid", src, 40)); got != want {
			t.Errorf("got\n%s\nwant\n%s", got, want)
		}
		l := strings.Split(got, "\n")
		if len(l) != 5 || !strings.HasSuffix(l[1], "--------") || !strings.HasSuffix(l[2], "=======") ||
			!strings.HasSuffix(l[3], "#######") || strings.Count(l[4], "*") != 1 {
			t.Errorf("bars:\n%s", got)
		}
	})
	t.Run("framed source", func(t *testing.T) {
		src := "classDiagram\n    class User"
		want := []string{
			"+- mermaid - classDiagram - not supported +",
			"| classDiagram                            |",
			"|     class User                          |",
			"+" + strings.Repeat("-", 41) + "+",
		}
		if got := FrameASCII("mermaid - classDiagram - not supported", src); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("FrameASCII:\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
		if res := RenderASCII("mermaid", src, 80); !res.Fallback || lines(res) != strings.Join(want, "\n")+"\n" {
			t.Errorf("RenderASCII:\n%s", lines(res))
		}
	})
	t.Run("text of the diagram", func(t *testing.T) {
		got := drawnASCII(t, "mermaid", "flowchart TD\n    A[Café] --> B", 80)
		if !strings.Contains(got, "Café") || foreign(got, "é") != "" {
			t.Errorf("got\n%s", got)
		}
	})
	for _, name := range []string{"plantuml-seq.puml", "plantuml-mindmap.puml", "plantuml-state.puml"} {
		src := readFixture(t, name)
		if f := foreign(drawnASCII(t, "plantuml", src, 80), src); f != "" {
			t.Errorf("%s: non-ASCII characters %q", name, f)
		}
	}
}
