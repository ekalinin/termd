package text

import (
	"strings"
	"testing"

	"github.com/ekalinin/termd/internal/style"
)

func TestWidth(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"ZWJ family emoji", "👨‍👩‍👧", 2},
		{"emoji with variation selector", "⚠️", 2},
		{"CJK", "日本語", 6},
		{"Cyrillic", "кириллица", 9},
		{"combining mark", "é", 1},
		{"ASCII", "string", 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Width(tt.in); got != tt.want {
				t.Errorf("Width(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func lineStrings(lines []Line) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.String()
	}
	return out
}

func TestWrapNoLineExceedsWidth(t *testing.T) {
	para := strings.Repeat("Максимальная ширина вывода в колонках терминала 日本語 ", 6)
	for _, width := range []int{10, 20, 40, 80} {
		for _, l := range Wrap([]Span{{Text: para}}, width, true) {
			if l.Width() > width {
				t.Errorf("width %d: line %q is %d wide", width, l.String(), l.Width())
			}
		}
	}
}

func TestWrapBreaksAtSpaces(t *testing.T) {
	got := lineStrings(Wrap([]Span{{Text: "one two  three\tfour"}}, 9, true))
	want := []string{"one two", "three", "four"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapLongWord(t *testing.T) {
	long := strings.Repeat("x", 120)
	got := Wrap([]Span{{Text: long}}, 80, true)
	if len(got) != 2 || got[0].Width() != 80 || got[1].Width() != 40 {
		t.Fatalf("got widths %v, want [80 40]", widths(got))
	}
}

func TestWrapLongWordKeptWhole(t *testing.T) {
	long := strings.Repeat("x", 120)
	got := Wrap([]Span{{Text: "a " + long + " b"}}, 80, false)
	if len(got) != 3 || got[1].Width() != 120 {
		t.Fatalf("got widths %v, want [1 120 1]", widths(got))
	}
}

func TestWrapForcedBreak(t *testing.T) {
	got := lineStrings(Wrap([]Span{{Text: "a"}, Break, {Text: "b"}}, 80, true))
	if strings.Join(got, "|") != "a|b" {
		t.Errorf("got %q", got)
	}
}

func TestWrapKeepsStyles(t *testing.T) {
	bold := style.Style{Bold: true}
	got := Wrap([]Span{{Text: "plain "}, {Text: "bold", Style: bold}, {Text: ", tail"}}, 80, true)
	if len(got) != 1 {
		t.Fatalf("got %d lines", len(got))
	}
	var styled string
	for _, sp := range got[0] {
		if sp.Style == bold {
			styled += sp.Text
		}
	}
	if styled != "bold" {
		t.Errorf("bold text = %q, want %q", styled, "bold")
	}
}

func widths(lines []Line) []int {
	out := make([]int, len(lines))
	for i, l := range lines {
		out[i] = l.Width()
	}
	return out
}

func TestStyledLineWidth(t *testing.T) {
	l := Line{{Text: "bold", Style: style.Style{Bold: true}}}
	if got := l.Width(); got != 4 {
		t.Errorf("Width() = %d, want 4", got)
	}
	rendered := l.Render(style.Options{Styled: true})
	if !strings.Contains(rendered, "\x1b[1m") {
		t.Errorf("rendered line %q has no bold sequence", rendered)
	}
}

func TestMaxWordAndNaturalWidth(t *testing.T) {
	spans := []Span{{Text: "По умолчанию"}}
	if got := MaxWordWidth(spans); got != 9 {
		t.Errorf("MaxWordWidth = %d, want 9", got)
	}
	if got := NaturalWidth(spans); got != 12 {
		t.Errorf("NaturalWidth = %d, want 12", got)
	}
}

func TestPad(t *testing.T) {
	l := Plain("42")
	for _, tt := range []struct {
		align byte
		want  string
	}{{'l', "42   "}, {'r', "   42"}, {'c', " 42  "}} {
		if got := Pad(l, 5, tt.align).String(); got != tt.want {
			t.Errorf("align %c: got %q, want %q", tt.align, got, tt.want)
		}
	}
}

func TestLinkSpans(t *testing.T) {
	url := "https://example.com/very/long/path"
	label := []Span{{Text: "docs"}}

	t.Run("hyperlinks enabled", func(t *testing.T) {
		l := Line(LinkSpans(label, url, true))
		if l.String() != "docs" || l.Width() != 4 {
			t.Errorf("visible %q (%d wide), want docs (4 wide)", l.String(), l.Width())
		}
		out := l.Render(style.Options{Styled: true, Hyperlinks: true})
		if !strings.Contains(out, "\x1b]8;;"+url+"\x1b\\") || !strings.Contains(out, "\x1b]8;;\x1b\\") {
			t.Errorf("rendered %q has no OSC 8 link", out)
		}
	})
	t.Run("hyperlinks disabled", func(t *testing.T) {
		l := Line(LinkSpans(label, url, false))
		if got := l.String(); got != "docs ("+url+")" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("autolink", func(t *testing.T) {
		l := Line(LinkSpans([]Span{{Text: "https://example.com"}}, "https://example.com", false))
		if got := l.String(); got != "https://example.com" {
			t.Errorf("got %q", got)
		}
	})
}
