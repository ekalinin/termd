package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/termd/internal/golden"
)

func TestHTMLStyles(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"bold", "\x1b[1mx\x1b[0m", `<span class="b">x</span>`},
		{"faint", "\x1b[2mx\x1b[0m", `<span class="f">x</span>`},
		{"italic", "\x1b[3mx\x1b[0m", `<span class="i">x</span>`},
		{"underline", "\x1b[4mx\x1b[0m", `<span class="u">x</span>`},
		{"strike", "\x1b[9mx\x1b[0m", `<span class="s">x</span>`},
		{"basic color", "\x1b[36mx\x1b[0m", `<span style="color:var(--ansi-36)">x</span>`},
		{"bright color", "\x1b[94mx\x1b[0m", `<span style="color:var(--ansi-94)">x</span>`},
		{"24-bit color", "\x1b[38;2;255;0;16mx\x1b[0m", `<span style="color:#ff0010">x</span>`},
		{"combined", "\x1b[1;3;38;2;1;2;3mx\x1b[0m", `<span class="b i" style="color:#010203">x</span>`},
		{"link style", "\x1b[4;34mx\x1b[0m", `<span class="u" style="color:var(--ansi-34)">x</span>`},
		{"reset", "\x1b[1ma\x1b[0mb", `<span class="b">a</span>b`},
		{"empty reset", "\x1b[1ma\x1b[mb", `<span class="b">a</span>b`},
		{"plain", "a\nb", "a\nb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTML(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("HTML(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestHTMLLinks(t *testing.T) {
	in := "see \x1b]8;;https://e.com/?a=1&b=\"x\"<y>\x1b\\\x1b[4;34mdocs\x1b[0m\x1b]8;;\x1b\\ now"
	want := `see <a href="https://e.com/?a=1&amp;b=&#34;x&#34;&lt;y&gt;"><span class="u" style="color:var(--ansi-34)">docs</span></a> now`
	got, err := HTML(in)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHTMLEscapesText(t *testing.T) {
	got, err := HTML(`<a & 'b' "c">`)
	if err != nil {
		t.Fatal(err)
	}
	if want := `&lt;a &amp; &#39;b&#39; &#34;c&#34;&gt;`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHTMLUnsupported(t *testing.T) {
	for name, in := range map[string]string{
		"256 colors":           "\x1b[38;5;99mx\x1b[0m",
		"cursor movement":      "a\x1b[2Ab",
		"unterminated OSC 8":   "\x1b]8;;https://e.com docs",
		"OSC 8 before a CSI":   "\x1b]8;;https://e.com\x1b[1mdocs\x1b\\",
		"OSC 8 with params":    "\x1b]8;id=1;https://e.com\x1b\\docs",
		"window title":         "\x1b]0;title\x07",
		"unknown SGR":          "\x1b[22mx",
		"unterminated CSI":     "x\x1b[1",
		"bare escape":          "x\x1bcy",
		"short 24-bit color":   "\x1b[38;2;1;2mx",
		"24-bit out of range":  "\x1b[38;2;1;2;256mx",
		"non-numeric SGR code": "\x1b[1;;mx",
	} {
		t.Run(name, func(t *testing.T) {
			if got, err := HTML(in); err == nil {
				t.Errorf("HTML(%q) = %q, want an error", in, got)
			}
		})
	}
}

func TestHTMLWideClusters(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"CJK", "a中b", `a<span class="w">中</span>b`},
		{"ZWJ emoji", "👩‍💻!", `<span class="w">👩‍💻</span>!`},
		{"variation selector", "⚠️ x", `<span class="w">⚠️</span> x`},
		{"ASCII", "plain text", "plain text"},
		{"styled", "\x1b[1m日本\x1b[0m", `<span class="b"><span class="w">日</span><span class="w">本</span></span>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTML(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("HTML(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestHTMLGolden converts every testdata/*.ansi file (styled termd output)
// and compares the HTML with testdata/<name>.html.golden.
func TestHTMLGolden(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.ansi"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no inputs found: %v", err)
	}
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".ansi")
		t.Run(name, func(t *testing.T) {
			in, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			got, err := HTML(string(in))
			if err != nil {
				t.Fatal(err)
			}
			golden.Assert(t, filepath.Join("testdata", fmt.Sprintf("%s.html.golden", name)), got)
		})
	}
}
