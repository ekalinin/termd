package render

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"

	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/text"
)

func plain(width int) Options {
	return Options{Width: width}
}

func styled(width int) Options {
	return Options{
		Width: width,
		Style: style.Options{Styled: true, Hyperlinks: true, Depth: style.TrueColor},
		Theme: func() highlight.Theme { return highlight.Dark },
	}
}

var escRE = regexp.MustCompile(`\x1b\[[0-9;]*m|\x1b\]8;;[^\x1b]*\x1b\\`)

func stripEscapes(s string) string {
	return escRE.ReplaceAllString(s, "")
}

func renderLines(src string, opts Options) []string {
	return strings.Split(strings.TrimSuffix(Render([]byte(src), opts), "\n"), "\n")
}

func TestParseGFM(t *testing.T) {
	src := "| a | b |\n|---|---|\n| 1 | 2 |\n\n- [x] done\n\n~~text~~ https://example.com\n"
	found := map[string]bool{}
	_ = ast.Walk(Parse([]byte(src)), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.(type) {
			case *east.Table:
				found["table"] = true
			case *east.TaskCheckBox:
				found["task"] = true
			case *east.Strikethrough:
				found["strikethrough"] = true
			case *ast.AutoLink:
				found["autolink"] = true
			}
		}
		return ast.WalkContinue, nil
	})
	for _, k := range []string{"table", "task", "strikethrough", "autolink"} {
		if !found[k] {
			t.Errorf("no %s node in the AST", k)
		}
	}
}

func TestParagraphWrapping(t *testing.T) {
	para := strings.Repeat("word ", 50)
	for _, l := range renderLines(para, plain(40)) {
		if w := text.Width(l); w > 40 {
			t.Errorf("line %q is %d wide", l, w)
		}
	}
	long := strings.Repeat("x", 120)
	got := renderLines(long, plain(80))
	if len(got) != 2 || len(got[0]) != 80 || len(got[1]) != 40 {
		t.Errorf("120-character word at width 80 rendered as %q", got)
	}
}

func TestHeadings(t *testing.T) {
	got := renderLines("# Title\n\n### Section\n", plain(80))
	want := []string{"# Title", "", "### Section"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got %q, want %q", got, want)
	}
	out := Render([]byte("# Title\n"), styled(80))
	if !strings.Contains(out, "\x1b[1m# Title") {
		t.Errorf("heading is not bold: %q", out)
	}
}

func TestInlineFormatting(t *testing.T) {
	src := "**bold** *italic* ~~gone~~ `code`\n"
	if got := renderLines(src, plain(80))[0]; got != "bold italic gone code" {
		t.Errorf("plain output = %q", got)
	}
	out := Render([]byte(src), styled(80))
	for _, seq := range []string{"\x1b[1mbold", "\x1b[3mitalic", "\x1b[9mgone", "\x1b[36mcode"} {
		if !strings.Contains(out, seq) {
			t.Errorf("styled output %q lacks %q", out, seq)
		}
	}
	if strings.ContainsAny(stripEscapes(out), "*~`") {
		t.Errorf("styled output keeps markdown markers: %q", stripEscapes(out))
	}
}

func TestPlainOutputHasNoEscapes(t *testing.T) {
	src := "# H\n\n**b** [l](https://x.y) `c`\n\n```go\nfunc main() {}\n```\n"
	if out := Render([]byte(src), plain(80)); strings.ContainsRune(out, 0x1b) {
		t.Errorf("plain output contains ESC: %q", out)
	}
}

func TestListContinuationAlignment(t *testing.T) {
	src := "- a list item that is long enough to wrap onto more lines\n"
	got := renderLines(src, plain(20))
	if !strings.HasPrefix(got[0], "• ") {
		t.Fatalf("first line %q has no bullet", got[0])
	}
	for _, l := range got[1:] {
		if !strings.HasPrefix(l, "  ") || strings.HasPrefix(l, "   ") {
			t.Errorf("continuation line %q is not aligned to the item text", l)
		}
	}
}

func TestNestedAndTaskLists(t *testing.T) {
	got := renderLines("- parent\n  - child\n- [ ] todo\n- [x] done\n", plain(80))
	want := []string{"• parent", "  • child", "• [ ] todo", "• [x] done"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBlockQuoteMarker(t *testing.T) {
	src := "> " + strings.Repeat("quoted words ", 12) + "\n"
	got := renderLines(src, plain(30))
	if len(got) < 3 {
		t.Fatalf("quote did not wrap: %q", got)
	}
	for _, l := range got {
		if !strings.HasPrefix(l, "│ ") || text.Width(l) > 30 {
			t.Errorf("quote line %q", l)
		}
	}
	alert := renderLines("> [!NOTE]\n"+src, plain(30))
	if len(alert) < 4 || alert[0] != "│ Note" {
		t.Fatalf("alert did not wrap: %q", alert)
	}
	for _, l := range alert {
		if !strings.HasPrefix(l, "│ ") || text.Width(l) > 30 {
			t.Errorf("alert line %q", l)
		}
	}
}

func TestAlerts(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"note", "> [!NOTE]\n> Useful info.\n", []string{"│ Note", "│ Useful info."}},
		{"lower case", "> [!warning]\n> Careful.\n", []string{"│ Warning", "│ Careful."}},
		{"hard line break", "> [!TIP]  \n> Hint.\n", []string{"│ Tip", "│ Hint."}},
		{"marker only", "> [!TIP]\n", []string{"│ Tip"}},
		{"several blocks", "> [!IMPORTANT]\n> First.\n>\n> Second.\n", []string{"│ Important", "│ First.", "│", "│ Second."}},
		{"nested quote", "> [!NOTE]\n> Text.\n>\n> > Quoted.\n", []string{"│ Note", "│ Text.", "│", "│ │ Quoted."}},
		{"list item", "- item\n  > [!TIP]\n  > Hint.\n", []string{"• item", "  │ Tip", "  │ Hint."}},
		{"text after the marker", "> [!NOTE] Useful info.\n", []string{"│ [!NOTE] Useful info."}},
		{"unknown marker", "> [!FOO]\n> Text.\n", []string{"│ [!FOO] Text."}},
		{"escaped marker", "> \\[!NOTE]\n> Text.\n", []string{"│ [!NOTE] Text."}},
		{"dotless i", "> [!tıp]\n> Text.\n", []string{"│ [!tıp] Text."}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderLines(tt.src, plain(80))
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAlertStyles(t *testing.T) {
	tests := []struct {
		marker, title, color string
	}{
		{"[!NOTE]", "Note", "34"},
		{"[!TIP]", "Tip", "32"},
		{"[!IMPORTANT]", "Important", "35"},
		{"[!WARNING]", "Warning", "33"},
		{"[!CAUTION]", "Caution", "31"},
	}
	for _, tt := range tests {
		src := "> " + tt.marker + "\n> Text.\n>\n> More text.\n"
		got := renderLines(src, styled(80))
		marker := "\x1b[" + tt.color + "m│"
		if want := marker + " \x1b[0m\x1b[1;" + tt.color + "m" + tt.title + "\x1b[0m"; got[0] != want {
			t.Errorf("%s title line = %q, want %q", tt.marker, got[0], want)
		}
		for _, l := range got {
			if !strings.HasPrefix(l, marker) {
				t.Errorf("%s line %q does not start with %q", tt.marker, l, marker)
			}
		}
		if out := Render([]byte(src), plain(80)); strings.ContainsRune(out, 0x1b) {
			t.Errorf("plain %s contains ESC: %q", tt.marker, out)
		}
	}
	got := renderLines("> [!NOTE]\n> Text.\n>\n> > Quoted.\n", styled(80))
	if want := "\x1b[34m│ \x1b[0m\x1b[2m│ \x1b[0mQuoted."; got[3] != want {
		t.Errorf("nested quote = %q, want %q", got[3], want)
	}
}

func TestCodeBlockWhitespace(t *testing.T) {
	code := "  two  spaces\n\tone tab\n    four"
	got := renderLines("```\n"+code+"\n```\n", plain(80))
	if strings.Join(got, "\n") != code {
		t.Errorf("got %q, want %q", strings.Join(got, "\n"), code)
	}
}

func TestThematicBreak(t *testing.T) {
	got := renderLines("a\n\n---\n\nb\n", plain(33))
	if got[2] != strings.Repeat("─", 33) {
		t.Errorf("break = %q", got[2])
	}
}

func TestImagesAndHTML(t *testing.T) {
	got := Render([]byte("![architecture](docs/arch.png)\n\n<details><summary>More</summary>text</details>\n"), plain(80))
	if !strings.Contains(got, "[image: architecture] (docs/arch.png)") {
		t.Errorf("image not rendered: %q", got)
	}
	if !strings.Contains(got, "<details><summary>More</summary>text</details>") {
		t.Errorf("raw HTML not kept: %q", got)
	}
}

func TestWideBlockIsNotTruncated(t *testing.T) {
	wide := strings.Repeat("0123456789", 12)
	got := renderLines("```\n"+wide+"\n```\n", plain(80))
	if len(got) != 1 || got[0] != wide {
		t.Errorf("wide code line changed: %q", got)
	}
	r := &renderer{src: []byte("```\n" + wide + "\n```\n"), opts: plain(80)}
	blocks := r.blocks(Parse(r.src), 80)
	if len(blocks) != 1 || !blocks[0].Wide {
		t.Errorf("code block wider than the output is not marked wide")
	}
}

func TestCodeWithoutKnownLanguageIsUncolored(t *testing.T) {
	for _, src := range []string{"```\nx := 1\n```\n", "```foo\nx := 1\n```\n", "    x := 1\n"} {
		out := Render([]byte(src), styled(80))
		if strings.ContainsRune(out, 0x1b) {
			t.Errorf("%q rendered with escapes: %q", src, out)
		}
	}
	if out := Render([]byte("```go\nfunc main() {}\n```\n"), styled(80)); !strings.Contains(out, "\x1b[") {
		t.Errorf("go block is not highlighted: %q", out)
	}
}

func TestThemeIsResolvedOnlyWhenNeeded(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		opts  func(Options) Options
		calls int
	}{
		{"no code", "# Title\n\ntext\n", nil, 0},
		{"unknown language", "```foo\nx\n```\n", nil, 0},
		{"diagram only", "```mermaid\ngraph TD\nA-->B\n```\n", nil, 0},
		{"two go blocks", "```go\na\n```\n\n```go\nb\n```\n", nil, 1},
		{"plain mode", "```go\na\n```\n", func(o Options) Options { o.Style.Styled = false; return o }, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			opts := styled(80)
			opts.Theme = func() highlight.Theme { calls++; return highlight.Dark }
			if tt.opts != nil {
				opts = tt.opts(opts)
			}
			Render([]byte(tt.src), opts)
			if calls != tt.calls {
				t.Errorf("theme resolved %d times, want %d", calls, tt.calls)
			}
		})
	}
}

func TestCode(t *testing.T) {
	wide := strings.Repeat("0123456789", 12)
	fence := "package main\n\n/*\n```\n# Title\n```\n*/\nfunc main() {}\n"
	tests := []struct {
		name, file, src, want string
	}{
		{"final line break", "x.go", "a\nb\n", "a\nb\n"},
		{"no final line break", "x.go", "a\nb", "a\nb\n"},
		{"trailing empty line", "x.go", "a\n\n", "a\n\n"},
		{"empty file", "x.go", "", ""},
		{"comments and dashes", "config.yaml", "---\n# Server settings\nport: 8080\n", "---\n# Server settings\nport: 8080\n"},
		{"fence line", "main.go", fence, fence},
		{"wide line", "x.go", wide + "\n", wide + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Code([]byte(tt.src), tt.file, plain(80)); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	src := "// main does nothing.\nfunc main() {}\n"
	out := Code([]byte(src), "main.go", styled(80))
	if !strings.Contains(out, "\x1b[") || stripEscapes(out) != src {
		t.Errorf("styled main.go = %q", out)
	}
}

func TestCodeThemeIsResolvedOnlyWhenNeeded(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		styled bool
		calls  int
	}{
		{"styled", "func main() {}\n", true, 1},
		{"plain mode", "func main() {}\n", false, 0},
		{"empty file", "", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			opts := styled(80)
			opts.Style.Styled = tt.styled
			opts.Theme = func() highlight.Theme { calls++; return highlight.Dark }
			Code([]byte(tt.src), "main.go", opts)
			if calls != tt.calls {
				t.Errorf("theme resolved %d times, want %d", calls, tt.calls)
			}
		})
	}
}

func TestHyperlinks(t *testing.T) {
	url := "https://example.com/very/long/path"
	src := "[docs](" + url + ")\n"
	out := Render([]byte(src), styled(80))
	if stripEscapes(strings.TrimSpace(out)) != "docs" || !strings.Contains(out, "\x1b]8;;"+url+"\x1b\\") {
		t.Errorf("styled link = %q", out)
	}
	if got := renderLines(src, plain(80))[0]; got != "docs ("+url+")" {
		t.Errorf("plain link = %q", got)
	}
	if got := renderLines("https://example.com\n", plain(80))[0]; got != "https://example.com" {
		t.Errorf("autolink = %q", got)
	}
}

func TestResolveLink(t *testing.T) {
	const dir, host = "/home/u/proj", "box"
	tests := []struct {
		dest, want string
	}{
		{"docs/guide.md", "file://box/home/u/proj/docs/guide.md"},
		{"./a.md", "file://box/home/u/proj/a.md"},
		{"../other/b.md", "file://box/home/u/other/b.md"},
		{"my notes.md", "file://box/home/u/proj/my%20notes.md"},
		{"my%20notes.md", "file://box/home/u/proj/my%20notes.md"},
		{"файл.md", "file://box/home/u/proj/%D1%84%D0%B0%D0%B9%D0%BB.md"},
		{`my\_notes.md`, "file://box/home/u/proj/my_notes.md"},
		{"guide.md#setup", "file://box/home/u/proj/guide.md#setup"},
		{"logo.png?raw=true", "file://box/home/u/proj/logo.png"},
		{"https://example.com", "https://example.com"},
		{"mailto:me@example.com", "mailto:me@example.com"},
		{"C:/x", "C:/x"},
		{"#usage", "#usage"},
		{"?tab=1", "?tab=1"},
		{"/docs/x.md", "/docs/x.md"},
		{"//example.com/x", "//example.com/x"},
		{"", ""},
		{"%zz", "%zz"},
	}
	for _, tt := range tests {
		if got := resolveLink(tt.dest, dir, host); got != tt.want {
			t.Errorf("resolveLink(%q) = %q, want %q", tt.dest, got, tt.want)
		}
	}
	if got := resolveLink("docs/guide.md", "", host); got != "docs/guide.md" {
		t.Errorf("without a directory: %q", got)
	}
	if got := resolveLink("docs/guide.md", dir, ""); got != "file:///home/u/proj/docs/guide.md" {
		t.Errorf("without a host: %q", got)
	}
}

func TestRelativeLinks(t *testing.T) {
	host, _ := os.Hostname()
	src := "[guide](docs/guide.md) ![arch](img/arch.png) [empty]()\n"
	opts := Options{Width: 80, Style: style.Options{Hyperlinks: true}, Dir: "/home/u/proj"}
	out := Render([]byte(src), opts)
	for _, want := range []string{
		"\x1b]8;;file://" + host + "/home/u/proj/docs/guide.md\x1b\\guide\x1b]8;;\x1b\\",
		"\x1b]8;;file://" + host + "/home/u/proj/img/arch.png\x1b\\[image: arch]\x1b]8;;\x1b\\",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q lacks %q", out, want)
		}
	}
	if strings.Count(out, "\x1b]8;;\x1b\\") != 2 || !strings.HasSuffix(out, " empty\n") {
		t.Errorf("empty destination got a hyperlink: %q", out)
	}

	opts.Style.Hyperlinks = false
	if got := renderLines(src, opts)[0]; got != "guide (docs/guide.md) [image: arch] (img/arch.png) empty" {
		t.Errorf("plain output = %q", got)
	}

	opts.Style.Hyperlinks, opts.Dir = true, ""
	if out := Render([]byte(src), opts); !strings.Contains(out, "\x1b]8;;docs/guide.md\x1b\\guide") {
		t.Errorf("without a directory: %q", out)
	}
}

func TestFileURL(t *testing.T) {
	if got := fileURL("", "C:/proj/docs/guide.md", ""); got != "file:///C:/proj/docs/guide.md" {
		t.Errorf("Windows path: %q", got)
	}
}

func TestSplitFrontmatter(t *testing.T) {
	tests := []struct {
		name string
		src  string
		kind fmKind
		// fmSrc is the block between the delimiters, body what is left to
		// render as markdown.
		fmSrc string
		body  string
	}{
		{"mapping", "---\ntitle: Doc\n---\n# Hello\n", fmTable, "title: Doc\n", "# Hello\n"},
		{"dots closing line", "---\ntitle: Doc\n...\n# Hello\n", fmTable, "title: Doc\n", "# Hello\n"},
		{"trailing spaces", "--- \ntitle: Doc\n---  \n# Hello\n", fmTable, "title: Doc\n", "# Hello\n"},
		{"no body", "---\ntitle: Doc\n---", fmTable, "title: Doc\n", ""},
		{"empty block", "---\n---\n# Hello\n", fmEmpty, "", "# Hello\n"},
		{"comments only", "---\n# just a comment\n---\n# Hello\n", fmEmpty, "# just a comment\n", "# Hello\n"},
		{"scalar block", "---\nSome text\n---\n", fmNone, "", "---\nSome text\n---\n"},
		{"no closing line", "---\ntitle: Doc\n\n# Hello\n", fmNone, "", "---\ntitle: Doc\n\n# Hello\n"},
		{"not the first line", "Intro\n\n---\ntitle: Doc\n---\n", fmNone, "", "Intro\n\n---\ntitle: Doc\n---\n"},
		{"longer opening line", "----\ntitle: Doc\n---\n", fmNone, "", "----\ntitle: Doc\n---\n"},
		{"bom and crlf", "\ufeff---\r\ntitle: Doc\r\n---\r\n# Hello\r\n", fmTable, "title: Doc\n", "# Hello\r\n"},
		{"invalid yaml", "---\ntitle: [unclosed\n---\n# Hello\n", fmInvalid, "title: [unclosed\n", "# Hello\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body := splitFrontmatter([]byte(tt.src))
			if fm.kind != tt.kind {
				t.Errorf("kind = %v, want %v", fm.kind, tt.kind)
			}
			if fm.src != tt.fmSrc {
				t.Errorf("frontmatter source = %q, want %q", fm.src, tt.fmSrc)
			}
			if string(body) != tt.body {
				t.Errorf("body = %q, want %q", body, tt.body)
			}
			if (fm.root != nil) != (tt.kind == fmTable) {
				t.Errorf("root = %v for kind %v", fm.root, fm.kind)
			}
		})
	}
}

const frontmatterValues = "---\n" +
	"title: \"Doc: x\"\n" +
	"desc: |\n  line one\n  line two\n" +
	"tags: [a, b]\n" +
	"author:\n  name: Eugene\n  url: https://example.com\n" +
	"authors:\n  - name: A\n  - name: B\n" +
	"empty:\n" +
	"---\n# Hello\n"

func TestFrontmatterValues(t *testing.T) {
	got := renderLines(frontmatterValues, plain(80))
	want := []string{
		"title   │ Doc: x",
		"desc    │ line one",
		"        │ line two",
		"tags    │ a, b",
		"author  │ name: Eugene",
		"        │ url: https://example.com",
		"authors │ [{name: A}, {name: B}]",
		"empty   │",
		"",
		"# Hello",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestFrontmatterStyles(t *testing.T) {
	if out := Render([]byte(frontmatterValues), styled(80)); !strings.Contains(out, "\x1b[1mtitle") {
		t.Errorf("key is not bold: %q", out)
	}
	if out := Render([]byte(frontmatterValues), plain(80)); strings.ContainsRune(out, 0x1b) {
		t.Errorf("plain frontmatter contains ESC: %q", out)
	}
}

func TestInvalidFrontmatter(t *testing.T) {
	got := renderLines("---\ntitle: [unclosed\n---\n# Hello\n", plain(80))
	want := []string{
		"┌─ frontmatter - invalid YAML ┐",
		"│ title: [unclosed            │",
		"└─────────────────────────────┘",
		"",
		"# Hello",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	fm, _ := splitFrontmatter([]byte("---\ntitle: [unclosed\n---\n"))
	if b, _ := frontmatterBlock(fm, 20); !b.Wide {
		t.Error("frame wider than the output is not marked wide")
	}
}

func TestTableCells(t *testing.T) {
	url := "https://example.com/very/long/path/to/documentation/page"
	src := "| Ссылка | Код |\n|---|---|\n| [docs](" + url + ") | `code \\| pipe` |\n"
	out := Render([]byte(src), styled(80))
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines: %q", len(lines), lines)
	}
	row := stripEscapes(lines[2])
	if row != "docs   │ code | pipe" {
		t.Errorf("row = %q", row)
	}
	if !strings.Contains(lines[2], "\x1b]8;;"+url+"\x1b\\") {
		t.Errorf("row has no hyperlink: %q", lines[2])
	}
	if strings.Count(stripEscapes(lines[0]), "│") != 1 {
		t.Errorf("escaped pipe changed the number of columns: %q", lines[0])
	}
}

// sameWidth reports whether all lines are equally wide.
func sameWidth(lines []string) bool {
	for _, l := range lines {
		if text.Width(l) != text.Width(lines[0]) {
			return false
		}
	}
	return true
}

// column returns the display column at which sep starts in line, or -1.
func column(line, sep string) int {
	i := strings.Index(line, sep)
	if i < 0 {
		return -1
	}
	return text.Width(line[:i])
}

func TestControlCharacters(t *testing.T) {
	linksOnly := Options{Width: 80, Style: style.Options{Hyperlinks: true}}
	tests := []struct {
		name string
		src  string
		opts Options
		want string
	}{
		{"paragraph", "hello \x1b[31mRED\x1b[0m and \x1b]0;pwned\x07 title\n", plain(80), "hello ␛[31mRED␛[0m and ␛]0;pwned␇ title\n"},
		{"styled text", "**bold \x1b[31m**\n", styled(80), "\x1b[1mbold ␛[31m\x1b[0m\n"},
		{"nul, del, c1 and invalid utf-8", "a\x00b\x7fc\u009bd\xffe\n", plain(80), "a␀b␡c�d�e\n"},
		{"heading", "# Title \x1b[5m\n", plain(80), "# Title ␛[5m\n"},
		{"plain code block", "```\necho \x07\n```\n", styled(80), "echo ␇\n"},
		{"link with hyperlinks", "[docs](https://example.com/\x1b]0;pwned\x07)\n", linksOnly, "\x1b]8;;https://example.com/␛]0;pwned␇\x1b\\docs\x1b]8;;\x1b\\\n"},
		{"link in a pipe", "[docs](https://example.com/\x1b]0;pwned\x07)\n", plain(80), "docs (https://example.com/␛]0;pwned␇)\n"},
		{"image", "![alt\x1b\x07](x.png)\n", plain(80), "[image: alt␛␇] (x.png)\n"},
		{"inline html", "<span>\x1b[31m</span>\n", plain(80), "<span>␛[31m</span>\n"},
		{"html block", "<div>\n\x1b[31m\n</div>\n", plain(80), "<div>\n␛[31m\n</div>\n"},
		{"frontmatter value", "---\nraw: a\x1bb\n---\n", plain(80), "raw │ a␛b\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Render([]byte(tt.src), tt.opts); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("highlighted code block", func(t *testing.T) {
		out := Render([]byte("```go\nfmt.Println(\"\x1b[31m\")\n```\n"), styled(80))
		if got := stripEscapes(out); got != "fmt.Println(\"␛[31m\")\n" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("table", func(t *testing.T) {
		got := renderLines("| Key | Value |\n|---|---|\n| a\x1bb | x |\n", plain(80))
		if len(got) != 3 || !strings.HasPrefix(got[2], "a␛b │ x") {
			t.Fatalf("got %q", got)
		}
		if column(got[0], "│") != 4 || column(got[1], "┼") != 4 || column(got[2], "│") != 4 {
			t.Errorf("separators are not aligned: %q", got)
		}
	})
	t.Run("invalid frontmatter", func(t *testing.T) {
		got := renderLines("---\ntitle: [\x1b[31m\n---\n", plain(80))
		if len(got) != 3 || !strings.HasPrefix(got[1], "│ title: [␛[31m ") || !sameWidth(got) {
			t.Errorf("got %q", got)
		}
	})
	t.Run("flowchart", func(t *testing.T) {
		got := renderLines("```mermaid\ngraph LR\n  A[\x1b[31mStart] --> B[End]\n```\n", plain(80))
		if len(got) < 3 || !strings.Contains(got[2], "│ ␛[31mStart ├") {
			t.Fatalf("got %q", got)
		}
		if column(got[0], "┐") != column(got[2], "├") {
			t.Errorf("box borders are not aligned: %q", got)
		}
	})
	t.Run("unsupported diagram", func(t *testing.T) {
		got := renderLines("```mermaid\nstateDiagram-v2\n  [*] --> \x1b[31mIdle\n```\n", plain(80))
		if len(got) != 4 || !strings.Contains(got[2], "[*] --> ␛[31mIdle") || !sameWidth(got) {
			t.Errorf("got %q", got)
		}
	})
}

func TestLineEndings(t *testing.T) {
	lf := "first line\nsecond line\n\n```\ncode a\ncode b\n```\n\n```go\nx := 1\ny := 2\n```\n"
	crlf := strings.ReplaceAll(lf, "\n", "\r\n")
	for _, opts := range []Options{plain(80), styled(80)} {
		got, want := Render([]byte(crlf), opts), Render([]byte(lf), opts)
		if got != want || strings.Contains(got, "\r") {
			t.Errorf("CRLF document rendered as %q, want %q", got, want)
		}
	}
	if got := Render([]byte(lf), plain(80)); !strings.Contains(got, "first line second line\n\ncode a\ncode b\n") {
		t.Errorf("LF document rendered as %q", got)
	}
	if got := Render([]byte("one\rtwo\n"), plain(80)); got != "one two\n" {
		t.Errorf("lone CR in a paragraph rendered as %q", got)
	}
	if got := Render([]byte("```\nthree\rfour\n```\n"), plain(80)); got != "three\nfour\n" {
		t.Errorf("lone CR in a code block rendered as %q", got)
	}
	got := renderLines("\ufeff---\r\ntitle: Doc\r\n---\r\n# Hello\r\n", plain(80))
	want := []string{"title │ Doc", "", "# Hello"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("BOM and CRLF frontmatter rendered as %q, want %q", got, want)
	}
}

func TestDecodedControlCharacters(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"character references", "&#27;[31m red &#7; &#13; &#x9b;\n", "␛[31m red ␇ ␍ �\n"},
		{"image alt", "![&#27;](x.png)\n", "[image: ␛] (x.png)\n"},
		{"table cell", "| Key | Value |\n|---|---|\n| &#27; | x |\n", "Key │ Value\n────┼──────\n␛   │ x\n"},
		{"yaml escapes", "---\n" +
			"title: \"\\e]0;pwned\\a\"\n" +
			"tags: [\"\\e\", b]\n" +
			"\"\\a\": x\n" +
			"nested: [{k: \"\\e\"}]\n" +
			"---\n",
			"title  │ ␛]0;pwned␇\n" +
				"tags   │ ␛, b\n" +
				"␇      │ x\n" +
				"nested │ [{k: \"␛\"}]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Render([]byte(tt.src), plain(80)); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
