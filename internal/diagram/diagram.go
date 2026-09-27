package diagram

import (
	"fmt"
	"regexp"
	"strings"

	mdiagram "github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	mrender "github.com/AlexanderGrooff/mermaid-ascii/pkg/render"

	"github.com/ekalinin/termd/internal/text"
)

// Language returns the diagram language named by a code block info string
// and whether the block is a diagram at all.
func Language(info string) (string, bool) {
	fields := strings.Fields(info)
	if len(fields) == 0 {
		return "", false
	}
	switch lang := strings.ToLower(fields[0]); lang {
	case "mermaid", "plantuml", "puml":
		return lang, true
	}
	return "", false
}

// headerLine returns the index of the line that declares the mermaid diagram
// type: the first line that is not blank, not a %% comment and not part of a
// leading --- frontmatter block. It returns -1 when there is none.
func headerLine(lines []string) int {
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
		for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "---"; i++ {
		}
		i++
	}
	for ; i < len(lines); i++ {
		s := strings.TrimSpace(lines[i])
		if s == "" || strings.HasPrefix(s, "%%") {
			continue
		}
		return i
	}
	return -1
}

// MermaidType returns the diagram type keyword of a mermaid source, for
// example "sequenceDiagram" or "flowchart", or "" when there is none.
func MermaidType(src string) string {
	lines := strings.Split(src, "\n")
	i := headerLine(lines)
	if i < 0 {
		return ""
	}
	return strings.Fields(lines[i])[0]
}

// Result is a rendered diagram block.
type Result struct {
	Lines []text.Line
	// Wide is true when the block is wider than the output width.
	Wide bool
	// Fallback is true when the source is shown in a frame.
	Fallback bool
}

// renderDiagram is the library entry point; tests replace it to inject
// failures.
var renderDiagram = mrender.RenderDiagramWithStatus

// safeRender calls the library and turns a panic into an error, so one
// broken diagram never aborts the document.
func safeRender(src string, cfg *mdiagram.Config) (out string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("renderer failed: %v", r)
		}
	}()
	out, _, err = renderDiagram(src, cfg)
	return clean(out), err
}

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// clean strips escape sequences, trailing spaces and trailing empty lines
// from library output.
func clean(out string) string {
	lines := strings.Split(ansiRE.ReplaceAllString(out, ""), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// blockWidth returns the width of the widest line of s.
func blockWidth(s string) int {
	w := 0
	for l := range strings.SplitSeq(s, "\n") {
		w = max(w, text.Width(l))
	}
	return w
}

// Render draws a diagram block for an output width. Supported mermaid
// diagrams are drawn as text; everything else is shown as framed source.
func Render(lang, src string, width int) Result {
	src = strings.TrimRight(src, "\n")
	if lang != "mermaid" {
		return fallback(src, width, lang, "not supported")
	}
	typ := MermaidType(src)
	var out string
	var err error
	switch typ {
	case "sequenceDiagram", "erDiagram":
		out, err = safeRender(src, mdiagram.DefaultConfig())
	case "flowchart", "graph":
		out, err = renderFlowchart(src, width)
	case "":
		return fallback(src, width, lang, "no diagram type")
	default:
		return fallback(src, width, lang, typ, "not supported")
	}
	if err != nil {
		return fallback(src, width, lang, typ, reason(err))
	}
	lines := strings.Split(out, "\n")
	res := Result{Lines: make([]text.Line, len(lines))}
	for i, l := range lines {
		res.Lines[i] = text.Plain(l)
	}
	res.Wide = blockWidth(out) > width
	return res
}

// renderFlowchart fits a flowchart into width: first with compacted spacing,
// then, for a horizontal flowchart, laid out top-to-bottom. The narrower
// result wins when neither fits.
func renderFlowchart(src string, width int) (string, error) {
	cfg := mdiagram.DefaultConfig()
	cfg.MaxWidth = width
	out, err := safeRender(src, cfg)
	if err != nil || blockWidth(out) <= width {
		return out, err
	}
	vertical, ok := toTopDown(src)
	if !ok {
		return out, nil
	}
	if out2, err2 := safeRender(vertical, cfg); err2 == nil && blockWidth(out2) < blockWidth(out) {
		return out2, nil
	}
	return out, nil
}

// toTopDown rewrites an LR or RL flowchart header to TD.
func toTopDown(src string) (string, bool) {
	lines := strings.Split(src, "\n")
	i := headerLine(lines)
	if i < 0 {
		return src, false
	}
	fields := strings.Fields(lines[i])
	if len(fields) < 2 {
		return src, false
	}
	switch strings.ToUpper(fields[1]) {
	case "LR", "RL":
	default:
		return src, false
	}
	indent := lines[i][:len(lines[i])-len(strings.TrimLeft(lines[i], " \t"))]
	fields[1] = "TD"
	lines[i] = indent + strings.Join(fields, " ")
	return strings.Join(lines, "\n"), true
}

// libraryPrefixRE matches the "failed to parse sequence diagram: " style
// prefix of library errors; the label already names the diagram type.
var libraryPrefixRE = regexp.MustCompile(`^failed to (parse|render|detect) [^:]*: `)

// reason returns the first line of err for a frame label.
func reason(err error) string {
	line, _, _ := strings.Cut(err.Error(), "\n")
	return libraryPrefixRE.ReplaceAllString(line, "$1 error: ")
}

// fallback shows src in a frame labelled with the language, the detected
// type and the reason.
func fallback(src string, width int, labelParts ...string) Result {
	lines := Frame(strings.Join(labelParts, " - "), src)
	res := Result{Fallback: true, Lines: make([]text.Line, len(lines))}
	for i, l := range lines {
		res.Lines[i] = text.Plain(l)
		res.Wide = res.Wide || text.Width(l) > width
	}
	return res
}

// Frame draws src inside a box whose top border carries label.
func Frame(label, src string) []string {
	body := strings.Split(strings.ReplaceAll(src, "\t", "    "), "\n")
	inner := text.Width(label) + 1
	for _, l := range body {
		inner = max(inner, text.Width(l))
	}
	out := make([]string, 0, len(body)+2)
	out = append(out, "┌─ "+label+" "+strings.Repeat("─", inner-text.Width(label)-1)+"┐")
	for _, l := range body {
		out = append(out, "│ "+l+strings.Repeat(" ", inner-text.Width(l))+" │")
	}
	out = append(out, "└"+strings.Repeat("─", inner+2)+"┘")
	return out
}
