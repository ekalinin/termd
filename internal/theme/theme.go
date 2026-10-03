package theme

import (
	"slices"
	"strconv"

	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/style"
)

// Palette is the style of every document element a theme styles outside
// code blocks.
type Palette struct {
	// Heading styles the # markers and the text of every heading level.
	Heading    style.Style
	InlineCode style.Style
	// Link styles the text of links and autolinks and the labels of images.
	Link style.Style
	// Marker styles list markers, horizontal lines and the quote marker of a
	// regular block quote.
	Marker style.Style
	// Note, Tip, Important, Warning and Caution style the title and the quote
	// marker of a GitHub alert; the title is also bold.
	Note, Tip, Important, Warning, Caution style.Style
	TableHeader                            style.Style
	// TableBorder styles the column separators and the line under the
	// header of a table.
	TableBorder    style.Style
	FrontmatterKey style.Style
}

// Theme is a built-in theme: its name, the palette of the document and the
// highlighting style of code blocks.
type Theme struct {
	Name    string
	Palette Palette
	Code    highlight.Theme
}

// Default is the palette of the dark and light themes: the attributes and
// the basic palette colors of the terminal.
var Default = Palette{
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

// themes are the built-in themes in the order they are listed to the user.
var themes = []Theme{
	{Name: "dark", Palette: Default, Code: highlight.Dark},
	{Name: "light", Palette: Default, Code: highlight.Light},
	// Heading, inline code, link, marker, note, tip, important, warning, caution.
	named("dracula", "#bd93f9", "#50fa7b", "#8be9fd", "#6272a4", "#bd93f9", "#50fa7b", "#ff79c6", "#f1fa8c", "#ff5555"),
	named("nord", "#88c0d0", "#8fbcbb", "#81a1c1", "#616e87", "#81a1c1", "#a3be8c", "#b48ead", "#ebcb8b", "#bf616a"),
	named("onedark", "#e06c75", "#98c379", "#61afef", "#7f848e", "#61afef", "#98c379", "#c678dd", "#e5c07b", "#e06c75"),
	named("monokai", "#a6e22e", "#fd971f", "#66d9ef", "#75715e", "#66d9ef", "#a6e22e", "#ae81ff", "#e6db74", "#f92672"),
	named("solarized-dark", "#cb4b16", "#2aa198", "#268bd2", "#586e75", "#268bd2", "#859900", "#d33682", "#b58900", "#dc322f"),
	named("solarized-light", "#cb4b16", "#2aa198", "#268bd2", "#93a1a1", "#268bd2", "#859900", "#d33682", "#b58900", "#dc322f"),
	named("gruvbox", "#b8bb26", "#8ec07c", "#83a598", "#928374", "#83a598", "#b8bb26", "#d3869b", "#fabd2f", "#fb4934"),
	named("gruvbox-light", "#79740e", "#427b58", "#076678", "#928374", "#076678", "#79740e", "#8f3f71", "#b57614", "#9d0006"),
	named("catppuccin-mocha", "#fab387", "#a6e3a1", "#89b4fa", "#6c7086", "#89b4fa", "#a6e3a1", "#cba6f7", "#f9e2af", "#f38ba8"),
	named("catppuccin-latte", "#fe640b", "#40a02b", "#1e66f5", "#9ca0b0", "#1e66f5", "#40a02b", "#8839ef", "#df8e1d", "#d20f39"),
}

// named returns a theme with 24-bit colors given in the order heading,
// inline code, link, marker, note, tip, important, warning and caution.
// Headings, table headers and frontmatter keys are bold in the heading
// color, links are underlined, and table borders take the marker color. The
// highlighting style is the chroma style of the same name.
func named(name string, colors ...string) Theme {
	if len(colors) != 9 {
		panic("theme: " + name + " needs 9 colors, got " + strconv.Itoa(len(colors)))
	}
	fg := func(i int) style.Style { return style.Style{FG: rgb(colors[i])} }
	heading := style.Style{Bold: true, FG: rgb(colors[0])}
	marker := fg(3)
	return Theme{
		Name: name,
		Code: highlight.Theme(name),
		Palette: Palette{
			Heading:        heading,
			InlineCode:     fg(1),
			Link:           style.Style{Underline: true, FG: rgb(colors[2])},
			Marker:         marker,
			Note:           fg(4),
			Tip:            fg(5),
			Important:      fg(6),
			Warning:        fg(7),
			Caution:        fg(8),
			TableHeader:    heading,
			TableBorder:    marker,
			FrontmatterKey: heading,
		},
	}
}

// rgb parses a #rrggbb color.
func rgb(hex string) style.Color {
	v, err := strconv.ParseUint(hex[min(1, len(hex)):], 16, 32)
	if len(hex) != 7 || hex[0] != '#' || err != nil {
		panic("theme: invalid color " + strconv.Quote(hex))
	}
	return style.RGB(uint8(v>>16), uint8(v>>8), uint8(v))
}

// Get returns the built-in theme with the given name.
func Get(name string) (Theme, bool) {
	i := slices.IndexFunc(themes, func(t Theme) bool { return t.Name == name })
	if i < 0 {
		return Theme{}, false
	}
	return themes[i], true
}

// Names returns the names of the built-in themes in the order they are
// listed to the user.
func Names() []string {
	names := make([]string, len(themes))
	for i, t := range themes {
		names[i] = t.Name
	}
	return names
}
