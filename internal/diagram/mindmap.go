package diagram

import (
	"errors"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/ekalinin/termd/internal/text"
)

// mindNode is a node of a mindmap: its text, split at its line breaks, and
// its children.
type mindNode struct {
	text     []string
	children []*mindNode
}

// mindShapes are the openers of the node shapes with their closers. The
// two-character openers come first, so (( is not read as (.
var mindShapes = []struct{ open, close string }{
	{"((", "))"}, {"))", "(("}, {"{{", "}}"}, {"(", ")"}, {")", "("}, {"[", "]"},
}

var (
	brRE       = regexp.MustCompile(`(?i)<br\s*/?>`)
	strongRE   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	emphasisRE = regexp.MustCompile(`\*(.+?)\*`)
)

// renderMindmap draws a mermaid mindmap as an indented tree, with the title
// of its frontmatter above it.
func renderMindmap(src string, width int) (string, error) {
	title, root, err := parseMindmap(src)
	if err != nil {
		return "", err
	}
	var out []string
	if title != "" {
		out = append(out, title, "")
	}
	out = root.layout(out, "", "", 0, width)
	return strings.Join(out, "\n"), nil
}

// layout appends the lines of n and its children to out. prefix starts the
// first line of n; guides starts its other lines and the lines of its
// children. The text of a node at depth d wraps at the width minus the 4
// columns of every level, but never at less than 10 columns, and words are
// never split.
func (n *mindNode) layout(out []string, prefix, guides string, depth, width int) []string {
	first := true
	for _, part := range n.text {
		for _, l := range text.Wrap([]text.Span{{Text: part}}, max(width-4*depth, 10), false) {
			p := guides
			if first {
				p, first = prefix, false
			}
			out = append(out, strings.TrimRight(p+l.String(), " "))
		}
	}
	for i, c := range n.children {
		conn, g := "├── ", "│   "
		if i == len(n.children)-1 {
			conn, g = "└── ", "    "
		}
		out = c.layout(out, guides+conn, guides+g, depth+1, width)
	}
	return out
}

// parseMindmap returns the frontmatter title and the root of a mindmap. A
// node is a child of the nearest node before it with less indentation.
func parseMindmap(src string) (string, *mindNode, error) {
	lines := strings.Split(strings.ReplaceAll(src, "\t", "    "), "\n")
	h := headerLine(lines)
	type open struct {
		node   *mindNode
		indent int
	}
	var root *mindNode
	var stack []open
	for i := h + 1; i < len(lines); i++ {
		s := strings.TrimSpace(lines[i])
		if s == "" || strings.HasPrefix(s, "%%") || strings.HasPrefix(s, "::icon(") || strings.HasPrefix(s, ":::") {
			continue
		}
		indent := len(lines[i]) - len(strings.TrimLeft(lines[i], " "))
		txt, last, err := nodeText(lines, i)
		if err != nil {
			return "", nil, err
		}
		i = last
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		n := &mindNode{text: txt}
		switch {
		case root == nil:
			root = n
		case len(stack) == 0:
			return "", nil, errors.New("more than one root")
		default:
			parent := stack[len(stack)-1].node
			parent.children = append(parent.children, n)
		}
		stack = append(stack, open{n, indent})
	}
	if root == nil {
		return "", nil, errors.New("no nodes")
	}
	return frontmatterTitle(lines[:h]), root, nil
}

// nodeText returns the text of the node on line i, split at its line breaks,
// and the index of the last line of the node: a markdown string can continue
// on the next lines.
func nodeText(lines []string, i int) ([]string, int, error) {
	s := strings.TrimSpace(lines[i])
	at, shape := -1, -1
	for k, sh := range mindShapes {
		if j := strings.Index(s, sh.open); j >= 0 && (at < 0 || j < at) {
			at, shape = j, k
		}
	}
	if at < 0 {
		return splitText(s), i, nil
	}
	rest := s[at+len(mindShapes[shape].open):]
	closer := mindShapes[shape].close
	if body, ok := strings.CutPrefix(rest, "\"`"); ok {
		for !strings.Contains(body, "`\"") {
			if i++; i >= len(lines) {
				return nil, i, errors.New("unclosed markdown string")
			}
			body += "\n" + strings.TrimSpace(lines[i])
		}
		md, after, _ := strings.Cut(body, "`\"")
		if strings.TrimSpace(after) != closer {
			return nil, i, errors.New("unclosed shape")
		}
		md = emphasisRE.ReplaceAllString(strongRE.ReplaceAllString(md, "$1"), "$1")
		return splitText(md), i, nil
	}
	body, ok := strings.CutSuffix(rest, closer)
	if !ok {
		return nil, i, errors.New("unclosed shape")
	}
	return splitText(body), i, nil
}

// splitText splits the text of a node at its line breaks: <br> tags and
// the line breaks of a markdown string.
func splitText(s string) []string {
	parts := strings.Split(brRE.ReplaceAllString(s, "\n"), "\n")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

// frontmatterTitle returns the title of a leading --- block of lines, or ""
// when there is none or the block is not valid YAML.
func frontmatterTitle(lines []string) string {
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i == len(lines) || strings.TrimSpace(lines[i]) != "---" {
		return ""
	}
	var fm struct {
		Title string `yaml:"title"`
	}
	end := i + 1
	for end < len(lines) && strings.TrimSpace(lines[end]) != "---" {
		end++
	}
	if yaml.Unmarshal([]byte(strings.Join(lines[i+1:end], "\n")), &fm) != nil {
		return ""
	}
	return fm.Title
}
