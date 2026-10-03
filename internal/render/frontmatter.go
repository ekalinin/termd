package render

import (
	"bytes"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/ekalinin/termd/internal/diagram"
	"github.com/ekalinin/termd/internal/table"
	"github.com/ekalinin/termd/internal/text"
	"github.com/ekalinin/termd/internal/theme"
)

// fmKind tells how the leading frontmatter block of a document is shown.
type fmKind int

const (
	// fmNone means the document does not start with frontmatter and is
	// rendered as markdown.
	fmNone fmKind = iota
	// fmEmpty is an empty or comment-only block, omitted from the output.
	fmEmpty
	// fmTable is a YAML mapping, shown as a key-value table.
	fmTable
	// fmInvalid is a block that is not valid YAML, shown as framed source.
	fmInvalid
)

// frontmatter is the YAML block at the start of a document.
type frontmatter struct {
	kind fmKind
	// src is the block between the delimiters, with LF line endings.
	src string
	// root is the mapping when kind is fmTable.
	root *yaml.Node
}

var bom = []byte("\ufeff")

// splitFrontmatter cuts a leading YAML frontmatter block off src and returns
// it with the rest of the document. The block starts with a --- line and
// ends with the next --- or ... line. When src does not start with
// frontmatter, the kind is fmNone and the body is src unchanged.
func splitFrontmatter(src []byte) (frontmatter, []byte) {
	line, rest := nextLine(bytes.TrimPrefix(src, bom))
	if !isDelimiter(line, false) {
		return frontmatter{}, src
	}
	var block strings.Builder
	for len(rest) > 0 {
		line, rest = nextLine(rest)
		if isDelimiter(line, true) {
			fm := parseFrontmatter(block.String())
			if fm.kind == fmNone {
				return fm, src
			}
			return fm, rest
		}
		block.Write(bytes.TrimSuffix(line, []byte("\r")))
		block.WriteByte('\n')
	}
	return frontmatter{}, src
}

// nextLine splits b into its first line, without the \n, and the rest.
func nextLine(b []byte) (line, rest []byte) {
	line, rest, _ = bytes.Cut(b, []byte("\n"))
	return line, rest
}

// isDelimiter reports whether line is a frontmatter delimiter: --- or, for
// the closing line, also ...; trailing whitespace is ignored.
func isDelimiter(line []byte, closing bool) bool {
	s := strings.TrimRight(string(line), " \t\r")
	return s == "---" || closing && s == "..."
}

// parseFrontmatter classifies the block between the delimiters. A panic in
// the YAML parser is treated as invalid YAML.
func parseFrontmatter(src string) (fm frontmatter) {
	defer func() {
		if recover() != nil {
			fm = frontmatter{kind: fmInvalid, src: src}
		}
	}()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
		return frontmatter{kind: fmInvalid, src: src}
	}
	if len(doc.Content) == 0 {
		return frontmatter{kind: fmEmpty, src: src}
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return frontmatter{}
	}
	if len(root.Content) == 0 {
		return frontmatter{kind: fmEmpty, src: src}
	}
	return frontmatter{kind: fmTable, src: src, root: root}
}

// frontmatterBlock renders the frontmatter as the first block of the
// document: a key-value table for a mapping, the source in a frame for
// invalid YAML. The palette styles the keys and the column separator. ok
// is false when there is nothing to show.
func frontmatterBlock(fm frontmatter, width int, pal theme.Palette) (b Block, ok bool) {
	switch fm.kind {
	case fmTable:
	case fmInvalid:
		for _, l := range diagram.Frame("frontmatter - invalid YAML", strings.TrimSuffix(fm.src, "\n")) {
			b.Lines = append(b.Lines, text.Plain(l))
			b.Wide = b.Wide || text.Width(l) > width
		}
		return b, true
	default:
		return Block{}, false
	}
	cleanValues(fm.root)
	t := table.Table{BorderStyle: pal.TableBorder}
	pairs := fm.root.Content
	for i := 0; i+1 < len(pairs); i += 2 {
		key := table.Cell{{Text: scalarOrFlow(pairs[i]), Style: pal.FrontmatterKey}}
		t.Rows = append(t.Rows, []table.Cell{key, valueSpans(pairs[i+1])})
	}
	lines, wide := t.Render(width)
	return Block{Lines: lines, Wide: wide}, true
}

// cleanValues replaces the control characters that YAML escapes such as
// "\e" put into the values of n and its descendants.
func cleanValues(n *yaml.Node) {
	n.Value = strings.Map(controlPicture, n.Value)
	for _, c := range n.Content {
		cleanValues(c)
	}
}

// valueSpans formats a frontmatter value: a scalar as its text with its line
// breaks, a list of scalars as its items joined with ", ", a mapping of
// scalars as one "key: value" line per entry, and anything else as YAML in
// flow style.
func valueSpans(n *yaml.Node) []text.Span {
	switch {
	case n.Kind == yaml.ScalarNode:
		return lineSpans(strings.Split(strings.TrimRight(n.Value, "\n"), "\n"))
	case n.Kind == yaml.SequenceNode && allScalars(n.Content):
		items := make([]string, len(n.Content))
		for i, c := range n.Content {
			items[i] = c.Value
		}
		return []text.Span{{Text: strings.Join(items, ", ")}}
	case n.Kind == yaml.MappingNode && allScalars(n.Content):
		var lines []string
		for i := 0; i+1 < len(n.Content); i += 2 {
			lines = append(lines, n.Content[i].Value+": "+n.Content[i+1].Value)
		}
		return lineSpans(lines)
	}
	return []text.Span{{Text: flow(n)}}
}

// lineSpans joins lines with forced line breaks.
func lineSpans(lines []string) []text.Span {
	var out []text.Span
	for i, l := range lines {
		if i > 0 {
			out = append(out, text.Break)
		}
		out = append(out, text.Span{Text: l})
	}
	return out
}

func allScalars(nodes []*yaml.Node) bool {
	for _, n := range nodes {
		if n.Kind != yaml.ScalarNode {
			return false
		}
	}
	return true
}

// scalarOrFlow returns the text of a scalar, or other nodes as flow YAML.
func scalarOrFlow(n *yaml.Node) string {
	if n.Kind == yaml.ScalarNode {
		return n.Value
	}
	return flow(n)
}

// flow returns n as YAML in flow style, for example [{name: A}, {name: B}].
func flow(n *yaml.Node) string {
	c := *n
	c.Style = yaml.FlowStyle
	out, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
