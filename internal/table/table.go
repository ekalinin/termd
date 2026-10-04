package table

import (
	"slices"
	"strings"

	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/text"
)

// Align is the alignment of a column.
type Align int

const (
	// AlignNone is the default and renders like AlignLeft.
	AlignNone Align = iota
	AlignLeft
	AlignCenter
	AlignRight
)

// Cell is the rendered inline content of a table cell.
type Cell []text.Span

// Table holds the cells of a GFM table. The header defines the number of
// columns. A table without a header takes the number of columns from its
// widest row and is rendered without a header line and rule.
type Table struct {
	Header []Cell
	Rows   [][]Cell
	Align  []Align
	// HeaderStyle is applied under the spans of the header cells: its
	// attributes are added, and its color is used for spans without one.
	HeaderStyle style.Style
	// BorderStyle styles the column separators and the rule under the header.
	BorderStyle style.Style
}

const (
	sep      = " │ "
	sepWidth = 3
	ruleSep  = "─┼─"
)

// cell returns cell i of row, or an empty cell for missing ones.
func cell(row []Cell, i int) Cell {
	if i < len(row) {
		return row[i]
	}
	return nil
}

// columns returns the number of columns and the rows to measure: the header
// and the body, or only the body when there is no header.
func (t Table) columns() (int, [][]Cell) {
	if len(t.Header) > 0 {
		return len(t.Header), append([][]Cell{t.Header}, t.Rows...)
	}
	n := 0
	for _, row := range t.Rows {
		n = max(n, len(row))
	}
	return n, t.Rows
}

// Measure returns the minimum (widest word) and natural (one line) width of
// every column.
func (t Table) Measure() (mins, naturals []int) {
	n, rows := t.columns()
	mins, naturals = make([]int, n), make([]int, n)
	for _, row := range rows {
		for i := range n {
			c := cell(row, i)
			mins[i] = max(mins[i], text.MaxWordWidth(c))
			naturals[i] = max(naturals[i], text.NaturalWidth(c))
		}
	}
	return mins, naturals
}

// Widths allocates column widths for an output width. overflow is true when
// even the minimum widths do not fit and the table is wider than width.
func (t Table) Widths(width int) (widths []int, overflow bool) {
	mins, naturals := t.Measure()
	n := len(mins)
	if n == 0 {
		return nil, false
	}
	avail := width - sepWidth*(n-1)

	// Step 1: everything fits on one line.
	if sum(naturals) <= avail {
		return naturals, false
	}
	// Step 3: even the minimum widths do not fit.
	if sum(mins) > avail {
		return mins, true
	}

	// Step 2: start at the minimum, give columns with the smallest shortfall
	// their natural width first, then share the rest proportionally.
	cur := slices.Clone(mins)
	budget := avail - sum(mins)
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		return (naturals[a] - mins[a]) - (naturals[b] - mins[b])
	})
	for _, i := range order {
		need := naturals[i] - cur[i]
		if need > budget {
			break
		}
		cur[i] = naturals[i]
		budget -= need
	}
	distribute(cur, naturals, budget)
	return cur, false
}

// distribute shares budget among columns that are narrower than their
// natural width, proportionally to their shortfall, using the largest
// remainder method so the whole budget is used.
func distribute(cur, naturals []int, budget int) {
	total := 0
	for i := range cur {
		total += naturals[i] - cur[i]
	}
	if budget <= 0 || total == 0 {
		return
	}
	type rem struct{ i, r int }
	var rems []rem
	given := 0
	for i := range cur {
		need := naturals[i] - cur[i]
		if need == 0 {
			continue
		}
		share := budget * need / total
		cur[i] += share
		given += share
		rems = append(rems, rem{i, budget * need % total})
	}
	slices.SortStableFunc(rems, func(a, b rem) int { return b.r - a.r })
	for k := 0; given < budget && k < len(rems); k++ {
		if cur[rems[k].i] < naturals[rems[k].i] {
			cur[rems[k].i]++
			given++
		}
	}
}

func sum(xs []int) int {
	s := 0
	for _, x := range xs {
		s += x
	}
	return s
}

// Render lays out the table for width. wide is true when the table does not
// fit and is emitted at its minimum widths.
func (t Table) Render(width int) (lines []text.Line, wide bool) {
	widths, overflow := t.Widths(width)
	if len(widths) == 0 {
		return nil, false
	}
	if len(t.Header) > 0 {
		header := make([]Cell, len(t.Header))
		for i, c := range t.Header {
			header[i] = withStyle(c, t.HeaderStyle)
		}
		lines = append(lines, t.row(header, widths)...)

		parts := make([]string, len(widths))
		for i, w := range widths {
			parts[i] = strings.Repeat("─", w)
		}
		rule := text.Plain(strings.Join(parts, ruleSep))
		for i := range rule {
			rule[i].Style = t.BorderStyle
		}
		lines = append(lines, rule)
	}

	for _, row := range t.Rows {
		lines = append(lines, t.row(row, widths)...)
	}
	return lines, overflow
}

// row wraps every cell of a row to its column width and joins the cells
// line by line.
func (t Table) row(row []Cell, widths []int) []text.Line {
	wrapped := make([][]text.Line, len(widths))
	height := 1
	for i, w := range widths {
		wrapped[i] = text.Wrap(cell(row, i), w, false)
		height = max(height, len(wrapped[i]))
	}
	lines := make([]text.Line, height)
	for j := range height {
		var l text.Line
		for i, w := range widths {
			if i > 0 {
				l = append(l, text.Span{Text: sep, Style: t.BorderStyle})
			}
			var part text.Line
			if j < len(wrapped[i]) {
				part = wrapped[i][j]
			}
			l = append(l, text.Pad(part, w, t.alignByte(i))...)
		}
		lines[j] = trimRight(l)
	}
	return lines
}

func (t Table) alignByte(i int) byte {
	if i >= len(t.Align) {
		return 'l'
	}
	switch t.Align[i] {
	case AlignCenter:
		return 'c'
	case AlignRight:
		return 'r'
	}
	return 'l'
}

// withStyle applies each span over s: the attributes of s are added, and
// the color of s is used for spans without a color of their own.
func withStyle(c Cell, s style.Style) Cell {
	out := make(Cell, len(c))
	for i, sp := range c {
		sp.Style = style.Layer(s, sp.Style)
		out[i] = sp
	}
	return out
}

// trimRight drops trailing padding spaces from the end of a line.
func trimRight(l text.Line) text.Line {
	for len(l) > 0 {
		last := l[len(l)-1]
		trimmed := strings.TrimRight(last.Text, " ")
		if trimmed != "" {
			last.Text = trimmed
			l[len(l)-1] = last
			return l
		}
		l = l[:len(l)-1]
	}
	return l
}
