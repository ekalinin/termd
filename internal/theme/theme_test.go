package theme

import (
	"slices"
	"testing"

	"github.com/alecthomas/chroma/v2/styles"

	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/style"
)

// specNames are the themes of the color-themes spec, in its order.
var specNames = []string{
	"dark", "light", "dracula", "nord", "onedark", "monokai",
	"solarized-dark", "solarized-light", "gruvbox", "gruvbox-light",
	"catppuccin-mocha", "catppuccin-latte",
}

func TestNames(t *testing.T) {
	if got := Names(); !slices.Equal(got, specNames) {
		t.Errorf("Names() = %q, want %q", got, specNames)
	}
}

func TestGet(t *testing.T) {
	for _, name := range specNames {
		th, ok := Get(name)
		if !ok || th.Name != name {
			t.Errorf("Get(%q) = %q, %v", name, th.Name, ok)
		}
	}
	for _, name := range []string{"auto", "blue", "", "Dracula"} {
		if _, ok := Get(name); ok {
			t.Errorf("Get(%q) found a theme", name)
		}
	}
}

func TestCodeIsChromaStyle(t *testing.T) {
	chroma := styles.Names()
	for _, name := range specNames {
		th, _ := Get(name)
		if !slices.Contains(chroma, string(th.Code)) {
			t.Errorf("%s: highlighting style %q is not a chroma style", name, th.Code)
		}
	}
}

func TestDefaultThemes(t *testing.T) {
	dark, _ := Get("dark")
	light, _ := Get("light")
	if dark.Palette != Default || light.Palette != Default {
		t.Error("dark and light do not use the default palette")
	}
	if dark.Code != highlight.Dark || light.Code != highlight.Light {
		t.Errorf("codes are %q and %q, want %q and %q", dark.Code, light.Code, highlight.Dark, highlight.Light)
	}
}

func TestDefaultPalette(t *testing.T) {
	want := Palette{
		Heading:        style.Style{Bold: true},
		InlineCode:     style.Style{ANSI: 36},
		Link:           style.Style{Underline: true, ANSI: 34},
		Marker:         style.Style{Faint: true},
		Note:           style.Style{ANSI: 34},
		Tip:            style.Style{ANSI: 32},
		Important:      style.Style{ANSI: 35},
		Warning:        style.Style{ANSI: 33},
		Caution:        style.Style{ANSI: 31},
		TableHeader:    style.Style{Bold: true},
		FrontmatterKey: style.Style{Bold: true},
	}
	if Default != want {
		t.Errorf("Default = %+v, want %+v", Default, want)
	}
}

func TestNamedPalettes(t *testing.T) {
	for _, name := range specNames[2:] {
		th, _ := Get(name)
		p := th.Palette
		colored := map[string]style.Style{
			"heading": p.Heading, "inline code": p.InlineCode, "link": p.Link, "marker": p.Marker,
			"note": p.Note, "tip": p.Tip, "important": p.Important, "warning": p.Warning, "caution": p.Caution,
			"table header": p.TableHeader, "table border": p.TableBorder, "frontmatter key": p.FrontmatterKey,
		}
		for field, s := range colored {
			if !s.FG.Set || s.ANSI != 0 {
				t.Errorf("%s: %s has no 24-bit color: %+v", name, field, s)
			}
			if s.Faint {
				t.Errorf("%s: %s is faint", name, field)
			}
		}
		if !p.Heading.Bold || p.TableHeader != p.Heading || p.FrontmatterKey != p.Heading {
			t.Errorf("%s: heading, table header and frontmatter key differ or are not bold", name)
		}
		if !p.Link.Underline {
			t.Errorf("%s: link is not underlined", name)
		}
		if p.TableBorder != p.Marker {
			t.Errorf("%s: table border %+v differs from marker %+v", name, p.TableBorder, p.Marker)
		}
		for field, s := range map[string]style.Style{
			"inline code": p.InlineCode, "marker": p.Marker, "note": p.Note, "tip": p.Tip,
			"important": p.Important, "warning": p.Warning, "caution": p.Caution,
		} {
			if s != (style.Style{FG: s.FG}) {
				t.Errorf("%s: %s has attributes besides its color: %+v", name, field, s)
			}
		}
	}
}

func TestSpotColors(t *testing.T) {
	tests := []struct {
		theme string
		field func(Palette) style.Style
		want  style.Color
	}{
		{"dracula", func(p Palette) style.Style { return p.Heading }, style.RGB(0xbd, 0x93, 0xf9)},
		{"nord", func(p Palette) style.Style { return p.Warning }, style.RGB(0xeb, 0xcb, 0x8b)},
		{"catppuccin-mocha", func(p Palette) style.Style { return p.Marker }, style.RGB(0x6c, 0x70, 0x86)},
	}
	for _, tt := range tests {
		th, _ := Get(tt.theme)
		if got := tt.field(th.Palette).FG; got != tt.want {
			t.Errorf("%s: color %+v, want %+v", tt.theme, got, tt.want)
		}
	}
}

func TestWindowColors(t *testing.T) {
	for _, name := range specNames {
		th, _ := Get(name)
		// github, the style of light, defines no text color: a light
		// terminal keeps its own.
		if fg, bg := th.Code.Colors(); !bg.Set || (!fg.Set && name != "light") {
			t.Errorf("%s: text color %+v, background %+v", name, fg, bg)
		}
	}
}
