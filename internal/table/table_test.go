package table

import (
	"strings"
	"testing"

	"github.com/rivo/uniseg"

	"github.com/ekalinin/termd/internal/text"
)

func cells(ss ...string) []Cell {
	out := make([]Cell, len(ss))
	for i, s := range ss {
		if s != "" {
			out[i] = Cell{{Text: s}}
		}
	}
	return out
}

func lineStrings(lines []text.Line) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.String()
	}
	return out
}

// separatorColumns returns the display columns of the column separators
// (│ and ┼) in a line.
func separatorColumns(line string) []int {
	var cols []int
	col := 0
	g := uniseg.NewGraphemes(line)
	for g.Next() {
		if s := g.Str(); s == "│" || s == "┼" {
			cols = append(cols, col)
		}
		col += g.Width()
	}
	return cols
}

func assertAligned(t *testing.T, lines []string) {
	t.Helper()
	want := separatorColumns(lines[0])
	for _, l := range lines[1:] {
		got := separatorColumns(l)
		if len(got) != len(want) {
			t.Errorf("line %q has separators at %v, want %v", l, got, want)
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("line %q has separators at %v, want %v", l, got, want)
				break
			}
		}
	}
}

const longDescription = "Максимальная ширина вывода в колонках терминала, после которой текст переносится"

func paramsTable() Table {
	return Table{
		Header: cells("Параметр", "Тип", "По умолчанию", "Описание"),
		Rows: [][]Cell{
			cells("--width", "int", "80", longDescription),
			cells("--theme", "string", "auto", "Цветовая тема: dark, light, auto, или путь к JSON-файлу с темой"),
		},
	}
}

func TestMeasure(t *testing.T) {
	tbl := Table{Header: cells("По умолчанию"), Rows: [][]Cell{cells("80"), cells("auto")}}
	mins, naturals := tbl.Measure()
	if mins[0] != 9 || naturals[0] != 12 {
		t.Errorf("min %d natural %d, want 9 and 12", mins[0], naturals[0])
	}
}

func TestWidthsFitNaturally(t *testing.T) {
	tbl := Table{Header: cells("a", "bb"), Rows: [][]Cell{cells("ccc", "d e")}}
	widths, overflow := tbl.Widths(80)
	if overflow || widths[0] != 3 || widths[1] != 3 {
		t.Errorf("widths %v overflow %v, want [3 3] false", widths, overflow)
	}
	lines, wide := tbl.Render(80)
	if wide || len(lines) != 3 {
		t.Errorf("table that fits naturally wrapped: %q", lineStrings(lines))
	}
}

func TestParamsTableAtWidth60(t *testing.T) {
	tbl := paramsTable()
	lines, wide := tbl.Render(60)
	got := lineStrings(lines)
	if wide {
		t.Fatal("table marked wide")
	}
	for _, l := range got {
		if w := text.Width(l); w > 60 {
			t.Errorf("line %q is %d wide", l, w)
		}
	}
	joined := strings.Join(got, "\n")
	for _, s := range []string{"--width", "--theme", "string", "auto", "По умолчанию"} {
		if !strings.Contains(joined, s) {
			t.Errorf("%q is not on one line:\n%s", s, joined)
		}
	}
	widths, _ := tbl.Widths(60)
	mins, naturals := tbl.Measure()
	for i := range 3 {
		if widths[i] != naturals[i] {
			t.Errorf("column %d got %d, want natural %d (min %d)", i, widths[i], naturals[i], mins[i])
		}
	}
	if widths[3] >= naturals[3] {
		t.Errorf("description column does not wrap: %d", widths[3])
	}
	assertAligned(t, got)
}

func TestNarrowTableKeepsWordsAndHeaders(t *testing.T) {
	tbl := paramsTable()
	lines, _ := tbl.Render(40)
	joined := strings.Join(lineStrings(lines), "\n")
	if strings.Contains(joined, "…") || strings.Contains(joined, "strin\n") {
		t.Errorf("table truncated or split a word:\n%s", joined)
	}
	for w := range strings.FieldsSeq(longDescription) {
		if !strings.Contains(joined, w) {
			t.Errorf("word %q was split:\n%s", w, joined)
		}
	}
	if !strings.Contains(joined, "По") || !strings.Contains(joined, "умолчанию") {
		t.Errorf("header was shortened:\n%s", joined)
	}
}

func TestOverflowUsesMinimumWidths(t *testing.T) {
	header := make([]string, 12)
	row := make([]string, 12)
	for i := range header {
		header[i] = "column"
		row[i] = "longword" + strings.Repeat("x", 3) + " a"
	}
	tbl := Table{Header: cells(header...), Rows: [][]Cell{cells(row...)}}
	mins, _ := tbl.Measure()
	widths, overflow := tbl.Widths(80)
	if !overflow {
		t.Fatal("overflow not reported")
	}
	for i := range widths {
		if widths[i] != mins[i] {
			t.Fatalf("widths %v, want minimums %v", widths, mins)
		}
	}
	lines, wide := tbl.Render(80)
	if !wide {
		t.Error("overflowing table not marked wide")
	}
	total := sum(mins) + sepWidth*11
	for _, l := range lineStrings(lines) {
		if strings.Contains(l, "longwordxxx") && text.Width(l) > total {
			t.Errorf("line %q wider than %d", l, total)
		}
	}
	if w := text.Width(lineStrings(lines)[1]); w != total {
		t.Errorf("rule is %d wide, want %d", w, total)
	}
}

func TestAlignment(t *testing.T) {
	tbl := Table{
		Header: cells("n", "centered"),
		Align:  []Align{AlignRight, AlignCenter},
		Rows:   [][]Cell{cells("1", "a"), cells("42", "bbb"), cells("1000", "cc")},
	}
	lines, _ := tbl.Render(80)
	got := lineStrings(lines)
	want := []string{
		"   n │ centered",
		"─────┼─────────",
		"   1 │    a",
		"  42 │   bbb",
		"1000 │    cc",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestSeparatorsWithEmojiAndCJK(t *testing.T) {
	tbl := Table{
		Header: cells("Статус", "Имя", "Число"),
		Align:  []Align{AlignCenter, AlignNone, AlignRight},
		Rows: [][]Cell{
			cells("✅", "жирный", "1"),
			cells("⚠️", "日本語テキスト", "1000"),
			cells("👨‍👩‍👧", "обычный", "42"),
		},
	}
	lines, _ := tbl.Render(80)
	assertAligned(t, lineStrings(lines))
}

func TestHeaderless(t *testing.T) {
	description := strings.TrimSpace(strings.Repeat("render tables and diagrams ", 6) + "in every modern terminal with care")
	if n := len(strings.Fields(description)); n != 30 {
		t.Fatalf("description has %d words, want 30", n)
	}
	tbl := Table{Rows: [][]Cell{cells("title", "Doc"), cells("description", description)}}
	lines, wide := tbl.Render(40)
	got := lineStrings(lines)
	if wide {
		t.Error("table marked wide")
	}
	if len(got) == 0 {
		t.Fatal("headerless table rendered no lines")
	}
	if !strings.HasPrefix(got[0], "title") {
		t.Errorf("first line %q is not the title row", got[0])
	}
	for _, l := range got {
		if strings.Contains(l, "─") {
			t.Errorf("headerless table has a rule: %q", l)
		}
		if w := text.Width(l); w > 40 {
			t.Errorf("line %q is %d wide", l, w)
		}
	}
	assertAligned(t, got)
}

func TestRaggedRows(t *testing.T) {
	tbl := Table{
		Header: cells("a", "b", "c"),
		Rows:   [][]Cell{cells("1", "2"), cells("1", "2", "3", "extra")},
	}
	lines, _ := tbl.Render(80)
	got := lineStrings(lines)
	if strings.Contains(strings.Join(got, "\n"), "extra") {
		t.Error("extra cell rendered")
	}
	// The short row keeps both separators.
	if n := strings.Count(got[2], "│"); n != 2 {
		t.Errorf("short row %q has %d separators, want 2", got[2], n)
	}
	assertAligned(t, got[:2])
}
