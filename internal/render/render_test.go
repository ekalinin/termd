package render

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	emast "github.com/yuin/goldmark-emoji/ast"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"golang.org/x/net/html"

	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/text"
	"github.com/ekalinin/termd/internal/theme"
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
	src := "| a | b |\n|---|---|\n| 1 | 2 |\n\n- [x] done\n\n~~text~~ https://example.com\n\nText[^1].\n\n[^1]: Note.\n\nTerm\n: Definition.\n\n:rocket:\n"
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
			case *east.FootnoteLink:
				found["footnote reference"] = true
			case *east.FootnoteList:
				found["footnote list"] = true
			case *east.DefinitionList:
				found["definition list"] = true
			case *east.DefinitionTerm:
				found["definition term"] = true
			case *east.DefinitionDescription:
				found["definition description"] = true
			case *emast.Emoji:
				found["emoji"] = true
			}
		}
		return ast.WalkContinue, nil
	})
	for _, k := range []string{"table", "task", "strikethrough", "autolink", "footnote reference", "footnote list", "definition list", "definition term", "definition description", "emoji"} {
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
	if !strings.HasSuffix(got, "\n▾ More\n  text\n") {
		t.Errorf("raw HTML not interpreted: %q", got)
	}
	opts := plain(80)
	opts.HTMLSource = true
	if got := Render([]byte("<details><summary>More</summary>text</details>\n"), opts); got != "<details><summary>More</summary>text</details>\n" {
		t.Errorf("raw HTML not kept as source: %q", got)
	}
}

func TestImageWithoutAlt(t *testing.T) {
	got := renderLines("![](logo.png) ![ ](icon.png) ![architecture](docs/arch.png)\n", plain(80))
	want := "[image] (logo.png) [image] (icon.png) [image: architecture] (docs/arch.png)"
	if len(got) != 1 || got[0] != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHTMLTokens(t *testing.T) {
	// tok writes a token as type:data, with the attributes of a tag.
	tok := func(t html.Token) string {
		s := t.Type.String() + ":" + t.Data
		for _, a := range t.Attr {
			s += " " + a.Key + "=" + a.Val
		}
		return s
	}
	tests := []struct {
		src  string
		want []string
	}{
		{`<a href="x&amp;y">A &lt;b&gt; &#27;</a>`, []string{"StartTag:a href=x&y", "Text:A <b> ␛", "EndTag:a"}},
		{`<img alt="&#7;" src=y>`, []string{"StartTag:img alt=␇ src=y"}},
		{"ok <b>x</b> <broken <p", []string{"Text:ok ", "StartTag:b", "Text:x", "EndTag:b", "Text: ", "Text:<broken <p"}},
		{`<p align="center`, []string{`Text:<p align="center`}},
		{"a < b", []string{"Text:a < b"}},
	}
	for _, tt := range tests {
		var got []string
		for _, t := range htmlTokens(tt.src) {
			got = append(got, tok(t))
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("htmlTokens(%q) = %q, want %q", tt.src, got, tt.want)
		}
	}
}

func TestInlineHTML(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"line break", "one<br>two\n", []string{"one", "two"}},
		{"link", "See <a href=\"https://example.com\">the *guide*</a>.\n", []string{"See the guide (https://example.com)."}},
		{"link without href", "An <a name=\"x\">anchor</a>.\n", []string{"An anchor."}},
		{"formatting tags", "Press <kbd>Ctrl</kbd>+<kbd>C</kbd>, <b>bold</b> and <em>italic</em>\n", []string{"Press Ctrl+C, bold and italic"}},
		{"removed tags", "x<sup>2</sup> <span class=\"note\">text</span>\n", []string{"x2 text"}},
		{"unknown tag in text", "Connect with ssh user@<host>.\n", []string{"Connect with ssh user@."}},
		{"unclosed and stray tags", "<b>bold\n\nnext</b> text\n", []string{"bold", "", "next text"}},
		{"block tags break the line", "one <div>two</div> three\n", []string{"one", "two", "three"}},
		{"comment", "a <!-- note --> b\n", []string{"a b"}},
		{"image", "<img src=\"logo.png\" alt=\"logo\"> and <img src=\"x.png\">\n", []string{"[image: logo] (logo.png) and [image] (x.png)"}},
		{"table cell", "| a | b |\n|---|---|\n| one<br>two | x |\n", []string{"a   │ b", "────┼──", "one │ x", "two │"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderLines(tt.src, plain(80))
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
	t.Run("source", func(t *testing.T) {
		opts := plain(80)
		opts.HTMLSource = true
		if got := renderLines("Press <kbd>Ctrl</kbd>.\n", opts); len(got) != 1 || got[0] != "Press <kbd>Ctrl</kbd>." {
			t.Errorf("got %q", got)
		}
	})
}

func TestInlineHTMLStyles(t *testing.T) {
	same := []struct{ html, md string }{
		{"<kbd>Ctrl</kbd>\n", "`Ctrl`\n"},
		{"<code>x</code>\n", "`x`\n"},
		{"<b>bold</b> <strong>strong</strong>\n", "**bold** **strong**\n"},
		{"<i>italic</i> <em>em</em>\n", "*italic* *em*\n"},
		{"<a href=\"https://example.com\">docs</a>\n", "[docs](https://example.com)\n"},
		{"<a href=\"https://example.com\"><img src=\"x.png\" alt=\"x\"></a>\n", "[![x](x.png)](https://example.com)\n"},
	}
	for _, tt := range same {
		if got, want := Render([]byte(tt.html), styled(80)), Render([]byte(tt.md), styled(80)); got != want {
			t.Errorf("%q = %q, want %q as for %q", tt.html, got, want, tt.md)
		}
	}
	t.Run("unclosed tag", func(t *testing.T) {
		got := renderLines("<b>bold\n\nnext</b> text\n", styled(80))
		want := []string{sgr("1") + "bold" + style.Reset, "", "next text"}
		if !slices.Equal(got, want) {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("control characters in a destination", func(t *testing.T) {
		opts := Options{Width: 80, Style: style.Options{Hyperlinks: true}}
		got := Render([]byte("<a href=\"https://example.com/&#27;]0;pwned&#7;\">docs</a>\n"), opts)
		if want := "\x1b]8;;https://example.com/␛]0;pwned␇\x1b\\docs\x1b]8;;\x1b\\\n"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestHTMLBlocks(t *testing.T) {
	badges := "<p align=\"center\">\n" +
		"<a href=\"https://ci.example.com\"><img src=\"https://ci.example.com/badge.svg\" alt=\"CI\"></a>\n" +
		"<a href=\"https://pkg.go.dev/x\"><img src=\"https://pkg.go.dev/badge/x.svg\"></a>\n" +
		"</p>\n"
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"centered logo", "<p align=\"center\">\n  <img src=\"logo.png\" alt=\"termd\" width=\"200\">\n</p>\n", []string{"[image: termd] (logo.png)"}},
		{"image without alt", "<img src=\"logo.png\">\n", []string{"[image] (logo.png)"}},
		{"linked images in a pipe", badges, []string{
			"[image: CI] (https://ci.example.com/badge.svg) (https://ci.example.com) [image]",
			"(https://pkg.go.dev/badge/x.svg) (https://pkg.go.dev/x)",
		}},
		{
			"header of a README",
			"<p align=\"center\"><img src=\"logo.png\" alt=\"logo\"></p>\n<h1 align=\"center\">termd</h1>\n<p align=\"center\">Markdown in the terminal</p>\n",
			[]string{"[image: logo] (logo.png)", "", "# termd", "", "Markdown in the terminal"},
		},
		{"heading level", "<h3>Usage</h3>\n", []string{"### Usage"}},
		{"paragraphs", "<div>one</div><div>two</div>\n", []string{"one", "", "two"}},
		{"text after a tag line", "<img src=\"a.png\">\ntext after\n", []string{"[image] (a.png) text after"}},
		{
			"script, style and comments",
			"A\n\n<!-- note -->\n\n<script>alert(1)</script>\n\n<style>p {}</style>\n\nB\n",
			[]string{"A", "", "B"},
		},
		{"character references", "<p>A &amp; B &lt;C&gt;</p>\n", []string{"A & B <C>"}},
		{"markdown in an HTML block", "<div>**bold**</div>\n", []string{"**bold**"}},
		{"broken HTML", "<p align=\"center\n", []string{"<p align=\"center"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderLines(tt.src, plain(80))
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
	t.Run("row of badges", func(t *testing.T) {
		out := Render([]byte(badges), Options{Width: 80, Style: style.Options{Hyperlinks: true}})
		if got := stripEscapes(out); got != "[image: CI] [image]\n" {
			t.Errorf("visible output %q", got)
		}
		for _, want := range []string{
			"\x1b]8;;https://ci.example.com\x1b\\[image: CI]\x1b]8;;\x1b\\",
			"\x1b]8;;https://pkg.go.dev/x\x1b\\[image]\x1b]8;;\x1b\\",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("output %q lacks %q", out, want)
			}
		}
	})
	t.Run("relative image in a file", func(t *testing.T) {
		host, _ := os.Hostname()
		opts := Options{Width: 80, Style: style.Options{Hyperlinks: true}, Dir: "/home/u/proj"}
		out := Render([]byte("<img src=\"img/arch.png\" alt=\"arch\">\n"), opts)
		if want := "\x1b]8;;file://" + host + "/home/u/proj/img/arch.png\x1b\\[image: arch]"; !strings.Contains(out, want) {
			t.Errorf("output %q lacks %q", out, want)
		}
	})
	t.Run("source", func(t *testing.T) {
		opts := plain(80)
		opts.HTMLSource = true
		src := "<p align=\"center\">\n  <img src=\"logo.png\" alt=\"termd\" width=\"200\">\n</p>\n"
		if got := Render([]byte(src), opts); got != src {
			t.Errorf("got %q, want %q", got, src)
		}
	})
}

func TestDetails(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			"around markdown blocks",
			"<details>\n<summary>Install</summary>\n\nRun **this**:\n\n```sh\ngo install x\n```\n\n</details>\n\nAfter.\n",
			[]string{"▾ Install", "  Run this:", "", "  go install x", "", "After."},
		},
		{"in one block", "<details><summary>More</summary>text</details>\n", []string{"▾ More", "  text"}},
		{"without summary", "<details>text</details>\n", []string{"▾ Details", "  text"}},
		{"without content", "<details><summary>More</summary></details>\n", []string{"▾ More"}},
		{
			"nested",
			"<details><summary>A</summary>\n\n<details><summary>B</summary>\n\ntext\n\n</details>\n\n</details>\n",
			[]string{"▾ A", "  ▾ B", "    text"},
		},
		{
			"not closed in a block quote",
			"> <details><summary>S</summary>\n>\n> text\n\nAfter.\n",
			[]string{"│ ▾ S", "│   text", "", "After."},
		},
		{
			"in a list item",
			"- item\n\n  <details><summary>S</summary>\n\n  text\n\n  </details>\n- next\n",
			[]string{"• item", "", "  ▾ S", "    text", "", "• next"},
		},
		{"stray end tag", "</details>\n\ntext\n", []string{"text"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderLines(tt.src, plain(40))
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
	t.Run("long content", func(t *testing.T) {
		got := renderLines("<details><summary>"+strings.Repeat("title ", 10)+"</summary>\n\n"+strings.Repeat("word ", 20)+"\n\n</details>\n", plain(40))
		if len(got) < 4 || !strings.HasPrefix(got[0], "▾ title") || !strings.HasPrefix(got[1], "  title") {
			t.Fatalf("summary did not wrap: %q", got)
		}
		for _, l := range got[1:] {
			if !strings.HasPrefix(l, "  ") || text.Width(l) > 40 {
				t.Errorf("line %q", l)
			}
		}
	})
	t.Run("source", func(t *testing.T) {
		opts := plain(40)
		opts.HTMLSource = true
		src := "<details>\n<summary>Install</summary>\n\ntext\n\n</details>\n"
		if got := Render([]byte(src), opts); got != src {
			t.Errorf("got %q, want %q", got, src)
		}
	})
	t.Run("marker style", func(t *testing.T) {
		opts := themed(t, "dracula")
		got := renderLines("<details><summary>More</summary>text</details>\n", opts)
		want := text.Line{{Text: "▾ ", Style: opts.Palette.Marker}, {Text: "More"}}.Render(opts.Style)
		if len(got) != 2 || got[0] != want {
			t.Errorf("got %q, want %q first", got, want)
		}
	})
}

func TestEmojiShortcodes(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"paragraph", "Fast :rocket: and tested :white_check_mark:\n", []string{"Fast 🚀 and tested ✅"}},
		{"heading", "# Features :sparkles:\n", []string{"# Features ✨"}},
		{"alert", "> [!NOTE]\n> :warning: Read this.\n", []string{"│ Note", "│ ⚠️ Read this."}},
		{"link text and destination", "[docs :book:](https://example.com/:book:)\n", []string{"docs 📖 (https://example.com/:book:)"}},
		{"between inline HTML tags", "Press <kbd>:arrow_up:</kbd> to scroll.\n", []string{"Press ⬆️ to scroll."}},
		{"after a letter or a digit", "a:b:c, foo:rocket:bar, 1:zap: and x :b: y\n", []string{"a:b:c, foo:rocket:bar, 1:zap: and x 🅱️ y"}},
		{"adjacent", ":rocket::zap:\n", []string{"🚀⚡"}},
		{"names that are not emoji", ":octocat: and :not_an_emoji:\n", []string{":octocat: and :not_an_emoji:"}},
		{
			"code and bare URLs",
			"`:zap:` and https://example.com/:smile:/x\n\n```\necho :zap:\n```\n",
			[]string{":zap: and https://example.com/:smile:/x", "", "echo :zap:"},
		},
		{"image alt text", "![logo :star:](logo.png)\n", []string{"[image: logo :star:] (logo.png)"}},
		{"HTML block", "<p align=\"center\">:rocket: Fast</p>\n", []string{":rocket: Fast"}},
		{"frontmatter", "---\ntitle: \":rocket: Launch\"\n---\n", []string{"title │ :rocket: Launch"}},
		{"definition list", "Term :one:\n: Def :two:\n", []string{"Term 1️⃣", "    Def 2️⃣"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderLines(tt.src, plain(80))
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestEmojiStyles(t *testing.T) {
	for _, tt := range []struct{ src, same string }{
		{"**:tada:**\n", "**🎉**\n"},
		{"# Features :sparkles:\n", "# Features ✨\n"},
	} {
		if got, want := Render([]byte(tt.src), styled(80)), Render([]byte(tt.same), styled(80)); got != want {
			t.Errorf("%q = %q, want %q as for %q", tt.src, got, want, tt.same)
		}
	}
	got := renderLines("| Feature | Status |\n|---|---|\n| Tables | :white_check_mark: |\n| Diagrams | :warning: |\n", plain(80))
	if len(got) != 4 || !strings.HasSuffix(got[2], "│ ✅") || !strings.HasSuffix(got[3], "│ ⚠️") {
		t.Fatalf("got %q", got)
	}
	if column(got[0], "│") != column(got[1], "┼") || column(got[0], "│") != column(got[2], "│") || column(got[0], "│") != column(got[3], "│") {
		t.Errorf("separators are not aligned: %q", got)
	}
	if text.Width("✅") != 2 || text.Width("⚠️") != 2 {
		t.Errorf("emoji widths %d and %d, want 2", text.Width("✅"), text.Width("⚠️"))
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
		{"control characters", "x.go", "a \x1b]0;x\x07 b\n", "a ␛]0;x␇ b\n"},
		{"crlf and lone cr", "x.go", "a\r\nb\rc\n", "a\nb\nc\n"},
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
	out = Code([]byte("// \x1b]0;x\x07\nfunc main() {}\n"), "main.go", styled(80))
	if want := "// ␛]0;x␇\nfunc main() {}\n"; stripEscapes(out) != want {
		t.Errorf("styled main.go with control characters = %q, want %q without styles", out, want)
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

func TestGitHubLinks(t *testing.T) {
	links := Options{Width: 80, Style: style.Options{Hyperlinks: true}}
	tests := []struct {
		name, src, url, text string
	}{
		{"pull request", "https://github.com/ekalinin/termd/pull/19", "https://github.com/ekalinin/termd/pull/19", "ekalinin/termd#19"},
		{"issue in angle brackets", "<https://github.com/o/r/issues/3>", "https://github.com/o/r/issues/3", "o/r#3"},
		{"discussion", "https://github.com/o/r/discussions/7", "https://github.com/o/r/discussions/7", "o/r#7"},
		{"issue comment", "https://github.com/o/r/issues/12#issuecomment-123", "https://github.com/o/r/issues/12#issuecomment-123", "o/r#12 (comment)"},
		{"review comment", "https://github.com/o/r/pull/12#discussion_r456", "https://github.com/o/r/pull/12#discussion_r456", "o/r#12 (comment)"},
		{"discussion comment", "https://github.com/o/r/discussions/12#discussioncomment-789", "https://github.com/o/r/discussions/12#discussioncomment-789", "o/r#12 (comment)"},
		{"review", "https://github.com/o/r/pull/12#pullrequestreview-5", "https://github.com/o/r/pull/12#pullrequestreview-5", "o/r#12 (review)"},
		{"commit", "https://github.com/o/r/commit/0123456789abcdef0123456789abcdef01234567", "https://github.com/o/r/commit/0123456789abcdef0123456789abcdef01234567", "o/r@0123456"},
		{"pull request commit", "https://github.com/o/r/pull/12/commits/abcdef1234567", "https://github.com/o/r/pull/12/commits/abcdef1234567", "o/r@abcdef1"},
		{"repository with a dot", "https://github.com/vercel/next.js/pull/12", "https://github.com/vercel/next.js/pull/12", "vercel/next.js#12"},
		{"scheme and host case", "http://GitHub.com/o/r/issues/2", "http://GitHub.com/o/r/issues/2", "o/r#2"},
		{"own text", "[PR](https://github.com/o/r/pull/1)", "https://github.com/o/r/pull/1", "PR"},
		{"own text equal to the url", "[https://github.com/o/r/pull/1](https://github.com/o/r/pull/1)", "https://github.com/o/r/pull/1", "https://github.com/o/r/pull/1"},
	}
	for _, url := range []string{
		"https://github.com/o/r",
		"https://github.com/o/r/blob/main/README.md",
		"https://github.com/o/r/pull/12/files",
		"https://github.com/o/r/pull/12/",
		"https://github.com/o/r/issues/12?x=1",
		"https://github.com/o/r/issues/12#top",
		"https://github.com/o/r/commit/abc12",
		"https://www.github.com/o/r/issues/1",
		"https://gitlab.com/o/r/-/issues/1",
	} {
		tests = append(tests, struct{ name, src, url, text string }{"not shortened " + url, url, url, url})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := "\x1b]8;;" + tt.url + "\x1b\\" + tt.text + "\x1b]8;;\x1b\\\n"
			if got := Render([]byte(tt.src+"\n"), links); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
	t.Run("width", func(t *testing.T) {
		out := Render([]byte("https://github.com/ekalinin/termd/pull/19\n"), links)
		if w := text.Width(stripEscapes(strings.TrimSpace(out))); w != 17 {
			t.Errorf("width = %d, want 17", w)
		}
	})
	t.Run("plain", func(t *testing.T) {
		const url = "https://github.com/ekalinin/termd/pull/19"
		if got := Render([]byte(url+"\n"), plain(80)); got != url+"\n" {
			t.Errorf("got %q, want %q", got, url+"\n")
		}
	})
	t.Run("table", func(t *testing.T) {
		const url = "https://github.com/o/r/pull/12"
		table := func(cell string) string { return "| PR | Note |\n|---|---|\n| " + cell + " | x |\n" }
		got := stripEscapes(Render([]byte(table(url)), links))
		if want := Render([]byte(table("o/r#12")), plain(80)); got != want {
			t.Errorf("with hyperlinks got\n%s\nwant\n%s", got, want)
		}
		if got := Render([]byte(table(url)), plain(80)); !strings.Contains(got, url) {
			t.Errorf("plain table has no %s:\n%s", url, got)
		}
	})
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
	if b, _ := frontmatterBlock(fm, 20, theme.Default); !b.Wide {
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
	source := Options{Width: 80, HTMLSource: true}
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
		{"inline html", "<span>\x1b[31m</span>\n", plain(80), "␛[31m\n"},
		{"inline html as source", "<span>\x1b[31m</span>\n", source, "<span>␛[31m</span>\n"},
		{"html block", "<div>\n\x1b[31m\n</div>\n", plain(80), "␛[31m\n"},
		{"html block as source", "<div>\n\x1b[31m\n</div>\n", source, "<div>\n␛[31m\n</div>\n"},
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

// footnoteBug was rendered as "Text^1 (Note.)." when [^1]: Note. was parsed
// as a link reference definition.
const footnoteBug = "Text[^1].\n\n[^1]: Note.\n"

func TestFootnoteReferences(t *testing.T) {
	tests := []struct {
		name, src, first string
	}{
		{"bug case", footnoteBug, "Text[1]."},
		{"numbered by first reference", "A[^b] B[^a] C[^b]\n\n[^a]: Alpha.\n[^b]: Beta.\n", "A[1] B[2] C[1]"},
		{"undefined reference", "Text[^x].\n", "Text[^x]."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderLines(tt.src, plain(40))[0]; got != tt.first {
				t.Errorf("first line = %q, want %q", got, tt.first)
			}
		})
	}
	if out := Render([]byte(footnoteBug), plain(40)); strings.Contains(out, "^1") || strings.Contains(out, "(Note.)") {
		t.Errorf("definition rendered as a link: %q", out)
	}
	out := Render([]byte(footnoteBug), styled(40))
	if !strings.Contains(out, "[1]") || strings.Contains(out, "\x1b]8;") {
		t.Errorf("styled reference = %q", out)
	}
}

func TestFootnotes(t *testing.T) {
	line := strings.Repeat("─", 40)
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"bug case", footnoteBug, []string{"Text[1].", "", line, "", "1. Note."}},
		{
			"numbered by first reference",
			"A[^b] B[^a] C[^b]\n\n[^a]: Alpha.\n[^b]: Beta.\n",
			[]string{"A[1] B[2] C[1]", "", line, "", "1. Beta.", "2. Alpha."},
		},
		{
			"definition in the middle",
			"First[^1].\n\n[^1]: Note.\n\nLast.\n",
			[]string{"First[1].", "", "Last.", "", line, "", "1. Note."},
		},
		{
			"unreferenced definition",
			"Text[^1].\n\n[^1]: Used.\n[^2]: Unused.\n",
			[]string{"Text[1].", "", line, "", "1. Used."},
		},
		{"no referenced definition", "Text.\n\n[^1]: Unused.\n", []string{"Text."}},
		{
			"two paragraphs",
			"Text[^1].\n\n[^1]: First paragraph.\n\n    Second paragraph.\n",
			[]string{"Text[1].", "", line, "", "1. First paragraph.", "", "   Second paragraph."},
		},
		{
			"long definition",
			"Text[^1].\n\n[^1]: one two three four five six seven eight nine ten eleven twelve.\n",
			[]string{"Text[1].", "", line, "", "1. one two three four five six seven", "   eight nine ten eleven twelve."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderLines(tt.src, plain(40))
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestFootnotesWithFrontmatter(t *testing.T) {
	got := renderLines("---\ntitle: Doc\n---\n"+footnoteBug, plain(40))
	if got[0] != "title │ Doc" || !slices.Contains(got, "Text[1].") || got[len(got)-1] != "1. Note." {
		t.Errorf("got %q", got)
	}
}

func TestDefinitionLists(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"term and definition", "Term\n: Definition of the term\n", []string{"Term", "    Definition of the term"}},
		{
			"several terms and definitions",
			"Apple\nPomme\n: A fruit\n: Red or green\n",
			[]string{"Apple", "Pomme", "    A fruit", "    Red or green"},
		},
		{
			"several entries",
			"Term A\n: Def A\n\nTerm B\n: Def B\n",
			[]string{"Term A", "    Def A", "Term B", "    Def B"},
		},
		{
			"loose list",
			"Term A\n\n: Def A\n\nTerm B\n: Def B\n",
			[]string{"Term A", "    Def A", "", "Term B", "    Def B"},
		},
		{
			"several blocks",
			"Term\n: First paragraph.\n\n    Second paragraph.\n",
			[]string{"Term", "    First paragraph.", "", "    Second paragraph."},
		},
		{
			"inline content",
			"`--width`\n: See [docs](https://example.com).\n",
			[]string{"--width", "    See docs (https://example.com)."},
		},
		{
			"paragraph lines become terms",
			"First line\nsecond line\n: Definition\n",
			[]string{"First line", "second line", "    Definition"},
		},
		{"colon without a space", "Term\n:not a definition\n", []string{"Term :not a definition"}},
		// goldmark makes the list tight: the definition list replaces the
		// paragraph that had the blank line before it.
		{"in a list item", "- item\n\n  Term\n  : Def\n", []string{"• item", "  Term", "      Def"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderLines(tt.src, plain(40))
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
	t.Run("long definition", func(t *testing.T) {
		got := renderLines("Term\n: "+strings.Repeat("word ", 20)+"\n", plain(40))
		if len(got) < 3 || got[0] != "Term" {
			t.Fatalf("definition did not wrap: %q", got)
		}
		for _, l := range got[1:] {
			if !strings.HasPrefix(l, "    word") || text.Width(l) > 40 {
				t.Errorf("definition line %q", l)
			}
		}
	})
}

func TestDefinitionListStyles(t *testing.T) {
	src := "Term\n: Definition of the term\n"
	t.Run("bold term", func(t *testing.T) {
		lines := renderLines(src, styled(80))
		if got, want := lines[0], sgr("1")+"Term"+style.Reset; got != want {
			t.Errorf("term = %q, want %q", got, want)
		}
		if got, want := lines[1], "    Definition of the term"; got != want {
			t.Errorf("definition = %q, want %q", got, want)
		}
	})
	t.Run("named theme", func(t *testing.T) {
		if got, want := renderLines(src, themed(t, "dracula"))[0], sgr("1")+"Term"+style.Reset; got != want {
			t.Errorf("term = %q, want %q", got, want)
		}
	})
	t.Run("inline code in a term", func(t *testing.T) {
		out := Render([]byte("`code` term\n: Def\n"), themed(t, "dracula"))
		if !strings.Contains(out, sgr("1;38;2;80;250;123")+"code") {
			t.Errorf("inline code in a term is not bold in the inline code color: %q", out)
		}
	})
}

// themed returns styled options without hyperlinks, with the palette and the
// highlighting theme of a built-in theme.
func themed(t *testing.T, name string) Options {
	t.Helper()
	th, ok := theme.Get(name)
	if !ok {
		t.Fatalf("no theme %q", name)
	}
	o := styled(80)
	o.Style.Hyperlinks = false
	o.Palette = th.Palette
	o.Theme = func() highlight.Theme { return th.Code }
	return o
}

func sgr(codes string) string {
	return "\x1b[" + codes + "m"
}

// findLine returns the first line whose text without escapes is want.
func findLine(t *testing.T, lines []string, want string) string {
	t.Helper()
	for _, l := range lines {
		if stripEscapes(l) == want {
			return l
		}
	}
	t.Fatalf("no line %q in %q", want, lines)
	return ""
}

func TestPalette(t *testing.T) {
	t.Run("zero palette is the default", func(t *testing.T) {
		files, err := filepath.Glob(filepath.Join(fixtures, "*.md"))
		if err != nil || len(files) == 0 {
			t.Fatalf("no fixtures found: %v", err)
		}
		def := styled(80)
		def.Palette = theme.Default
		for _, file := range files {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if Render(src, styled(80)) != Render(src, def) {
				t.Errorf("%s: zero palette and default palette differ", file)
			}
		}
	})
	t.Run("link inside a heading", func(t *testing.T) {
		out := Render([]byte("# See [docs](https://example.com)\n"), themed(t, "dracula"))
		if !strings.Contains(out, sgr("1;38;2;189;147;249")+"# See ") {
			t.Errorf("heading text is not bold #bd93f9: %q", out)
		}
		if !strings.Contains(out, sgr("1;4;38;2;139;233;253")+"docs") {
			t.Errorf("link is not bold, underlined and #8be9fd: %q", out)
		}
	})
	t.Run("inline code inside a link", func(t *testing.T) {
		out := Render([]byte("[`run`](https://example.com)\n"), themed(t, "dark"))
		if !strings.Contains(out, sgr("4;36")+"run") {
			t.Errorf("inline code in a link is not underlined cyan: %q", out)
		}
	})
	t.Run("heading in truecolor and 256 colors", func(t *testing.T) {
		o := themed(t, "dracula")
		if got, want := renderLines("# Title\n", o)[0], sgr("1;38;2;189;147;249")+"# Title"+style.Reset; got != want {
			t.Errorf("truecolor heading = %q, want %q", got, want)
		}
		o.Style.Depth = style.Color256
		if got, want := renderLines("# Title\n", o)[0], sgr("1;38;5;141")+"# Title"+style.Reset; got != want {
			t.Errorf("256-color heading = %q, want %q", got, want)
		}
	})
	t.Run("same style for every heading level", func(t *testing.T) {
		lines := renderLines("# Title\n\n### Section\n", themed(t, "nord"))
		if got, want := lines[2], strings.Replace(lines[0], "# Title", "### Section", 1); got != want {
			t.Errorf("### Section = %q, want the style of # Title: %q", got, want)
		}
	})
	t.Run("marker is not faint", func(t *testing.T) {
		if got, want := renderLines("- item\n", themed(t, "catppuccin-mocha"))[0], sgr("38;2;108;112;134")+"• "+style.Reset+"item"; got != want {
			t.Errorf("list item = %q, want %q", got, want)
		}
	})
	t.Run("alert in a named theme", func(t *testing.T) {
		lines := renderLines("> [!WARNING]\n> Careful.\n", themed(t, "nord"))
		marker := sgr("38;2;235;203;139") + "│ " + style.Reset
		want := []string{marker + sgr("1;38;2;235;203;139") + "Warning" + style.Reset, marker + "Careful."}
		if !slices.Equal(lines, want) {
			t.Errorf("alert = %q, want %q", lines, want)
		}
	})
	t.Run("table borders", func(t *testing.T) {
		lines := renderLines("| A | `b` |\n|---|---|\n| c | d |\n", themed(t, "nord"))
		sep := sgr("38;2;97;110;135") + " │ " + style.Reset
		want := []string{
			sgr("1;38;2;136;192;208") + "A" + style.Reset + sep + sgr("1;38;2;143;188;187") + "b" + style.Reset,
			sgr("38;2;97;110;135") + "──┼──" + style.Reset,
			"c" + sep + "d",
		}
		if !slices.Equal(lines, want) {
			t.Errorf("table = %q, want %q", lines, want)
		}
	})
	t.Run("unthemed elements", func(t *testing.T) {
		lines := renderLines("Text *italic* ref[^1].\n\n- [ ] todo\n\n[^1]: Note.\n", themed(t, "gruvbox"))
		if got, want := findLine(t, lines, "Text italic ref[1]."), "Text "+sgr("3")+"italic"+style.Reset+" ref[1]."; got != want {
			t.Errorf("paragraph = %q, want %q", got, want)
		}
		if got, want := findLine(t, lines, "• [ ] todo"), sgr("38;2;146;131;116")+"• "+style.Reset+"[ ] todo"; got != want {
			t.Errorf("task item = %q, want %q", got, want)
		}
	})
	t.Run("plain mode ignores the palette", func(t *testing.T) {
		src, err := os.ReadFile(filepath.Join(fixtures, "regression.md"))
		if err != nil {
			t.Fatal(err)
		}
		o := plain(80)
		o.Palette = themed(t, "dracula").Palette
		if Render(src, o) != Render(src, plain(80)) {
			t.Error("plain output depends on the palette")
		}
	})
}
