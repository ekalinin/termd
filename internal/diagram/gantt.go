package diagram

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ekalinin/termd/internal/text"
)

// ganttTask is a task of a gantt chart. start and end are its scheduled
// dates; bar is the end of its bar, before the excluded days that end it.
type ganttTask struct {
	name, id string
	// line is the number of the line of the task in the block, from 1.
	line int
	// section is the index of the section of the task, -1 before the first
	// section.
	section                       int
	done, active, milestone, vert bool
	// startSpec and endSpec are the start and the end items; a task with
	// prevStart starts at the end of the task prev.
	startSpec, endSpec string
	prevStart          bool
	prev               int
	start, end, bar    time.Time
	hasStart, hasEnd   bool
}

// ganttChart is a parsed and scheduled gantt chart.
type ganttChart struct {
	title, axisFormat string
	sections          []string
	tasks             []*ganttTask

	format   dateFormat
	weekends bool
	// weekend is the first day of the weekend.
	weekend            time.Weekday
	excludedDays       map[time.Weekday]bool
	excludes, includes map[string]bool
	inclusive          bool
}

var (
	accLineRE  = regexp.MustCompile(`^acc(Title|Descr)\s*:`)
	accBlockRE = regexp.MustCompile(`^accDescr\s*\{`)
	durationRE = regexp.MustCompile(`^(\d+(?:\.\d+)?)(ms|[Mdhmswy])$`)
	tokenSepRE = regexp.MustCompile(`[\s,]+`)
)

var ganttTags = []string{"done", "active", "crit", "milestone", "vert"}

// ignoredGanttKeywords are the keywords of lines that do not change the
// text output.
var ignoredGanttKeywords = []string{"click", "todayMarker", "tickInterval", "weekday"}

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday, "wednesday": time.Wednesday,
	"thursday": time.Thursday, "friday": time.Friday, "saturday": time.Saturday,
}

// maxExcludedDays bounds the days an end can move over excluded days, as
// Mermaid does, so excluding every day does not loop forever.
const maxExcludedDays = 10000

// ganttKeyword returns the value of the line s when it is the keyword k
// followed by whitespace and a value.
func ganttKeyword(s, k string) (string, bool) {
	rest, ok := strings.CutPrefix(s, k)
	if !ok || rest == "" || (rest[0] != ' ' && rest[0] != '\t') {
		return "", false
	}
	v := strings.TrimSpace(rest)
	return v, v != ""
}

// parseGantt parses a mermaid gantt chart and schedules its tasks.
func parseGantt(src string) (*ganttChart, error) {
	lines := strings.Split(src, "\n")
	h := headerLine(lines)
	c := &ganttChart{
		axisFormat:   "%Y-%m-%d",
		weekend:      time.Saturday,
		excludedDays: map[time.Weekday]bool{},
		excludes:     map[string]bool{},
		includes:     map[string]bool{},
	}
	format := "YYYY-MM-DD"
	var excludes, includes []string
	title, hasTitle := "", false
	section := -1
	for i := h + 1; i < len(lines); i++ {
		s := strings.TrimSpace(lines[i])
		if v, ok := ganttKeyword(s, "dateFormat"); ok {
			format = v
			continue
		}
		if v, ok := ganttKeyword(s, "axisFormat"); ok {
			c.axisFormat = v
			continue
		}
		if v, ok := ganttKeyword(s, "excludes"); ok {
			excludes = append(excludes, tokenSepRE.Split(v, -1)...)
			continue
		}
		if v, ok := ganttKeyword(s, "includes"); ok {
			includes = append(includes, tokenSepRE.Split(v, -1)...)
			continue
		}
		if v, ok := ganttKeyword(s, "weekend"); ok {
			switch v {
			case "friday":
				c.weekend = time.Friday
			case "saturday":
				c.weekend = time.Saturday
			default:
				return nil, fmt.Errorf("unsupported line %d", i+1)
			}
			continue
		}
		if v, ok := ganttKeyword(s, "title"); ok {
			title, hasTitle = v, true
			continue
		}
		if v, ok := ganttKeyword(s, "section"); ok {
			c.sections = append(c.sections, v)
			section = len(c.sections) - 1
			continue
		}
		if accBlockRE.MatchString(s) {
			for j := i; ; j++ {
				if j == len(lines) {
					return nil, fmt.Errorf("unsupported line %d", i+1)
				}
				if strings.Contains(lines[j], "}") {
					i = j
					break
				}
			}
			continue
		}
		switch {
		case s == "inclusiveEndDates":
			c.inclusive = true
			continue
		case s == "" || strings.HasPrefix(s, "%%") || s == "topAxis" || accLineRE.MatchString(s):
			continue
		case slices.ContainsFunc(ignoredGanttKeywords, func(k string) bool { _, ok := ganttKeyword(s, k); return ok }):
			continue
		}
		t, err := parseGanttTask(s, i+1, section, len(c.tasks)-1)
		if err != nil {
			return nil, err
		}
		c.tasks = append(c.tasks, t)
	}
	drawn := false
	for _, t := range c.tasks {
		drawn = drawn || !t.vert
	}
	if !drawn {
		return nil, errors.New("no tasks")
	}
	f, err := parseDateFormat(format)
	if err != nil {
		return nil, err
	}
	c.format = f
	c.setDays(excludes, includes)
	if hasTitle {
		c.title = title
	} else {
		c.title = frontmatterTitle(lines[:h])
	}
	return c, c.schedule()
}

// parseGanttTask parses the task line s, the line-th line of the block.
// prev is the index of the task before it, -1 for none.
func parseGanttTask(s string, line, section, prev int) (*ganttTask, error) {
	name, data, ok := strings.Cut(s, ":")
	if !ok {
		return nil, fmt.Errorf("unsupported line %d", line)
	}
	t := &ganttTask{name: strings.Join(strings.Fields(name), " "), line: line, section: section, prev: prev}
	items := strings.Split(data, ",")
	for i := range items {
		items[i] = strings.TrimSpace(items[i])
	}
	for len(items) > 0 && slices.Contains(ganttTags, items[0]) {
		switch items[0] {
		case "done":
			t.done = true
		case "active":
			t.active = true
		case "milestone":
			t.milestone = true
		case "vert":
			t.vert = true
		}
		items = items[1:]
	}
	switch {
	case len(items) == 1 && items[0] != "":
		t.endSpec, t.prevStart = items[0], true
	case len(items) == 2:
		t.startSpec, t.endSpec = items[0], items[1]
	case len(items) == 3:
		t.id, t.startSpec, t.endSpec = items[0], items[1], items[2]
	default:
		return nil, fmt.Errorf("unsupported line %d", line)
	}
	if t.prevStart && prev < 0 {
		return nil, fmt.Errorf("no start date on line %d", line)
	}
	return t, nil
}

// setDays reads the tokens of the excludes and includes lines: weekends,
// weekday names and dates.
func (c *ganttChart) setDays(excludes, includes []string) {
	iso, _ := parseDateFormat("YYYY-MM-DD")
	day := func(tok string) (string, bool) {
		for _, f := range []dateFormat{c.format, iso} {
			if d, ok := f.parse(tok); ok {
				return d.Format(time.DateOnly), true
			}
		}
		return "", false
	}
	for _, tok := range excludes {
		low := strings.ToLower(tok)
		if wd, ok := weekdays[low]; ok {
			c.excludedDays[wd] = true
		} else if low == "weekends" {
			c.weekends = true
		} else if d, ok := day(tok); ok {
			c.excludes[d] = true
		}
	}
	for _, tok := range includes {
		if d, ok := day(tok); ok {
			c.includes[d] = true
		}
	}
}

// excluded reports whether the day of d is excluded.
func (c *ganttChart) excluded(d time.Time) bool {
	key := d.Format(time.DateOnly)
	if c.includes[key] {
		return false
	}
	if c.weekends && (d.Weekday() == c.weekend || d.Weekday() == (c.weekend+1)%7) {
		return true
	}
	return c.excludedDays[d.Weekday()] || c.excludes[key]
}

// hasExcludes reports whether any day can be excluded.
func (c *ganttChart) hasExcludes() bool {
	return c.weekends || len(c.excludedDays) > 0 || len(c.excludes) > 0
}

// refs returns the ids after the keyword k of the item s, as in "after a b".
func refs(s, k string) ([]string, bool) {
	rest, ok := strings.CutPrefix(s, k)
	if !ok || rest == "" || (rest[0] != ' ' && rest[0] != '\t') {
		return nil, false
	}
	ids := strings.Fields(rest)
	return ids, len(ids) > 0
}

// schedule computes the dates of the tasks in passes, until every start and
// end is known. An id can name a task before or after the task that uses
// it.
func (c *ganttChart) schedule() error {
	ids := map[string]*ganttTask{}
	for _, t := range c.tasks {
		if t.id != "" {
			ids[t.id] = t
		}
	}
	for _, t := range c.tasks {
		if r, ok := refs(t.startSpec, "after"); ok {
			if err := known(r, ids); err != nil {
				return err
			}
		} else if !t.prevStart {
			if _, ok := c.format.parse(t.startSpec); !ok {
				return fmt.Errorf("invalid date on line %d", t.line)
			}
		}
		if r, ok := refs(t.endSpec, "until"); ok {
			if err := known(r, ids); err != nil {
				return err
			}
		}
	}
	for {
		progress, done := false, true
		for _, t := range c.tasks {
			if !t.hasStart {
				t.start, t.hasStart = c.startOf(t, ids)
				progress = progress || t.hasStart
			}
			if t.hasStart && !t.hasEnd {
				if err := c.resolveEnd(t, ids); err != nil {
					return err
				}
				progress = progress || t.hasEnd
			}
			done = done && t.hasEnd
		}
		if done {
			return nil
		}
		if !progress {
			return errors.New("circular dependency")
		}
	}
}

// known returns an error for the first id that no task has.
func known(r []string, ids map[string]*ganttTask) error {
	for _, id := range r {
		if ids[id] == nil {
			return fmt.Errorf("unknown task %s", id)
		}
	}
	return nil
}

// startOf returns the start of t when the tasks it depends on are
// scheduled: the latest end of the tasks of its after, the end of the task
// before it, or its start date.
func (c *ganttChart) startOf(t *ganttTask, ids map[string]*ganttTask) (time.Time, bool) {
	if t.prevStart {
		p := c.tasks[t.prev]
		return p.end, p.hasEnd
	}
	if r, ok := refs(t.startSpec, "after"); ok {
		var latest time.Time
		for i, id := range r {
			o := ids[id]
			if !o.hasEnd {
				return time.Time{}, false
			}
			if i == 0 || o.end.After(latest) {
				latest = o.end
			}
		}
		return latest, true
	}
	d, _ := c.format.parse(t.startSpec)
	return d, true
}

// resolveEnd sets the end of t, whose start is known, when the tasks it
// depends on are scheduled: the earliest start of the tasks of its until,
// its end date, or its start plus its duration, moved over excluded days.
// Any other end is the start.
func (c *ganttChart) resolveEnd(t *ganttTask, ids map[string]*ganttTask) error {
	if r, ok := refs(t.endSpec, "until"); ok {
		var earliest time.Time
		for i, id := range r {
			o := ids[id]
			if !o.hasStart {
				return nil
			}
			if i == 0 || o.start.Before(earliest) {
				earliest = o.start
			}
		}
		t.end = earliest
	} else if d, ok := c.format.parse(t.endSpec); ok {
		if c.inclusive {
			d = d.AddDate(0, 0, 1)
		}
		t.end = d
	} else if d, ok := addDuration(t.start, t.endSpec); ok {
		end, bar, err := c.skipExcluded(t.start, d)
		if err != nil {
			return err
		}
		t.end, t.bar, t.hasEnd = end, bar, true
		return nil
	} else {
		t.end = t.start
	}
	t.bar, t.hasEnd = t.end, true
	return nil
}

// skipExcluded checks the days after start one by one up to end, as
// fixTaskDates of Mermaid does: every excluded day moves the end one day
// later. It returns the moved end and the end of the bar, before the
// excluded days right before the moved end.
func (c *ganttChart) skipExcluded(start, end time.Time) (time.Time, time.Time, error) {
	if !c.hasExcludes() {
		return end, end, nil
	}
	bar, excluded, moved := end, false, 0
	for d := start.AddDate(0, 0, 1); !d.After(end); d = d.AddDate(0, 0, 1) {
		if !excluded {
			bar = end
		}
		if excluded = c.excluded(d); excluded {
			if moved++; moved > maxExcludedDays {
				return end, bar, errors.New("every day is excluded")
			}
			end = end.AddDate(0, 0, 1)
		}
	}
	return end, bar, nil
}

// addDuration adds a duration such as 3d or 1.5h to t: days and weeks as
// whole days, rounded, months and years as calendar months without the
// fractional part, the other units exactly.
func addDuration(t time.Time, s string) (time.Time, bool) {
	m := durationRE.FindStringSubmatch(s)
	if m == nil {
		return t, false
	}
	n, _ := strconv.ParseFloat(m[1], 64)
	switch m[2] {
	case "d":
		return t.AddDate(0, 0, int(math.Round(n))), true
	case "w":
		return t.AddDate(0, 0, int(math.Round(7*n))), true
	case "M":
		return addMonths(t, int(n)), true
	case "y":
		return addMonths(t, 12*int(n)), true
	}
	unit := map[string]time.Duration{"ms": time.Millisecond, "s": time.Second, "m": time.Minute, "h": time.Hour}[m[2]]
	return t.Add(time.Duration(n * float64(unit))), true
}

// addMonths adds n calendar months to t and keeps the day within the target
// month, as dayjs does: January 31 plus one month is the last day of
// February, where AddDate would give March 2.
func addMonths(t time.Time, n int) time.Time {
	y, m, d := t.Date()
	first := time.Date(y, m+time.Month(n), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(d, last)-1)
}

// dateToken is a token of a dateFormat: a field with its number of digits,
// 0 for one or two, or a literal.
type dateToken struct {
	field  byte
	digits int
	lit    string
}

// dateFormat is a parsed Mermaid dateFormat. unix is "X" or "x" for a
// timestamp format.
type dateFormat struct {
	unix   string
	tokens []dateToken
}

// dateTokens are the tokens of dayjs, longest first. A token without a field
// is one termd does not read.
var dateTokens = []struct {
	tok    string
	field  byte
	digits int
}{
	{"YYYY", 'Y', 4}, {"YY", 'y', 2},
	{"MMMM", 0, 0}, {"MMM", 0, 0}, {"MM", 'M', 2}, {"M", 'M', 0},
	{"Do", 0, 0}, {"DDDD", 0, 0}, {"DDD", 0, 0}, {"DD", 'D', 2}, {"D", 'D', 0},
	{"HH", 'H', 2}, {"H", 'H', 0}, {"hh", 0, 0}, {"h", 0, 0},
	{"mm", 'm', 2}, {"m", 'm', 0}, {"ss", 's', 2}, {"s", 's', 0},
	{"SSS", 0, 0}, {"SS", 0, 0}, {"S", 0, 0}, {"ZZ", 0, 0}, {"Z", 0, 0},
	{"Q", 0, 0}, {"A", 0, 0}, {"a", 0, 0}, {"X", 0, 0}, {"x", 0, 0},
}

var unixRE = regexp.MustCompile(`^[+-]?\d+(\.\d+)?$`)

// parseDateFormat splits a dateFormat into tokens and literals.
func parseDateFormat(f string) (dateFormat, error) {
	if f == "X" || f == "x" {
		return dateFormat{unix: f}, nil
	}
	var df dateFormat
	for f != "" {
		matched := false
		for _, t := range dateTokens {
			if rest, ok := strings.CutPrefix(f, t.tok); ok {
				if t.field == 0 {
					return df, errors.New("unsupported dateFormat")
				}
				df.tokens = append(df.tokens, dateToken{field: t.field, digits: t.digits})
				f, matched = rest, true
				break
			}
		}
		if !matched {
			_, n := utf8.DecodeRuneInString(f)
			df.tokens = append(df.tokens, dateToken{lit: f[:n]})
			f = f[n:]
		}
	}
	return df, nil
}

// parse reads a date in the format, in UTC. The parts that the format does
// not have are those of 1970-01-01 00:00:00. A date whose fields
// time.Date would normalize, such as 2024-02-30, is invalid.
func (f dateFormat) parse(s string) (time.Time, bool) {
	if f.unix != "" {
		if !unixRE.MatchString(s) {
			return time.Time{}, false
		}
		if f.unix == "x" {
			ms, err := strconv.ParseInt(s, 10, 64)
			return time.UnixMilli(ms).UTC(), err == nil
		}
		sec, err := strconv.ParseFloat(s, 64)
		return time.UnixMilli(int64(math.Round(sec * 1000))).UTC(), err == nil
	}
	v := map[byte]int{'Y': 1970, 'M': 1, 'D': 1}
	for _, t := range f.tokens {
		if t.field == 0 {
			rest, ok := strings.CutPrefix(s, t.lit)
			if !ok {
				return time.Time{}, false
			}
			s = rest
			continue
		}
		limit := t.digits
		if limit == 0 {
			limit = 2
		}
		n := 0
		for n < len(s) && n < limit && '0' <= s[n] && s[n] <= '9' {
			n++
		}
		if n == 0 || t.digits > 0 && n < t.digits {
			return time.Time{}, false
		}
		num, _ := strconv.Atoi(s[:n])
		s = s[n:]
		switch t.field {
		case 'y':
			if num > 68 {
				num += 1900
			} else {
				num += 2000
			}
			v['Y'] = num
		default:
			v[t.field] = num
		}
	}
	if s != "" {
		return time.Time{}, false
	}
	d := time.Date(v['Y'], time.Month(v['M']), v['D'], v['H'], v['m'], v['s'], 0, time.UTC)
	ok := d.Year() == v['Y'] && int(d.Month()) == v['M'] && d.Day() == v['D'] &&
		d.Hour() == v['H'] && d.Minute() == v['m'] && d.Second() == v['s']
	return d, ok
}

// renderGantt draws a mermaid gantt chart as bars on a time axis.
func renderGantt(src string, width int) (string, error) {
	c, err := parseGantt(src)
	if err != nil {
		return "", err
	}
	return c.layout(width), nil
}

// layout draws the chart for an output width: the title, the axis, and the
// tasks under their sections with their bars. The bars get at least 20
// columns; names wrap, splitting long words, so that every bar starts in
// the same column.
func (c *ganttChart) layout(width int) string {
	var drawn []*ganttTask
	var start, end time.Time
	names := 0
	for _, t := range c.tasks {
		if t.vert {
			continue
		}
		e := t.end
		if e.Before(t.start) {
			e = t.start
		}
		if len(drawn) == 0 || t.start.Before(start) {
			start = t.start
		}
		if len(drawn) == 0 || e.After(end) {
			end = e
		}
		drawn = append(drawn, t)
		names = max(names, indent(t)+text.Width(t.name))
	}
	bars := width - names - 2
	if bars < 20 {
		// The names give up columns to the bars, down to 10.
		names = min(names, max(width-22, 10))
		bars = max(width-names-2, 20)
	}
	span := end.Sub(start)
	col := func(at time.Time) int {
		if span <= 0 {
			return 0
		}
		return int(math.Round(float64(bars) * float64(at.Sub(start)) / float64(span)))
	}

	var out []string
	if c.title != "" {
		out = append(out, c.title, "")
	}
	from, to := formatAxis(start, c.axisFormat), formatAxis(end, c.axisFormat)
	axis := strings.Repeat(" ", names+2) + from
	if gap := bars - text.Width(from) - text.Width(to); gap >= 1 {
		axis += strings.Repeat(" ", gap) + to
	}
	out = append(out, strings.TrimRight(axis, " "))
	section := -1
	for _, t := range drawn {
		if t.section != section {
			section = t.section
			for _, l := range text.Wrap([]text.Span{{Text: c.sections[section]}}, width, false) {
				out = append(out, l.String())
			}
		}
		pad := strings.Repeat(" ", indent(t))
		for i, l := range text.Wrap([]text.Span{{Text: t.name}}, names-indent(t), true) {
			line := pad + l.String()
			if i == 0 {
				line += strings.Repeat(" ", max(names-text.Width(line), 0)+2) + c.bar(t, col, bars)
			}
			out = append(out, strings.TrimRight(line, " "))
		}
	}
	return strings.Join(out, "\n")
}

// indent returns the indentation of the name of t: 2 spaces in a section.
func indent(t *ganttTask) int {
	if t.section < 0 {
		return 0
	}
	return 2
}

// bar returns the cells of the bar of t, with col mapping an instant to a
// column of the bars bars columns wide.
func (c *ganttChart) bar(t *ganttTask, col func(time.Time) int, bars int) string {
	if t.milestone {
		end := t.end
		if end.Before(t.start) {
			end = t.start
		}
		return strings.Repeat(" ", min(col(t.start.Add(end.Sub(t.start)/2)), bars-1)) + "◆"
	}
	fill := "█"
	switch {
	case t.active:
		fill = "▓"
	case t.done:
		fill = "░"
	}
	from := min(col(t.start), bars-1)
	to := from
	if t.bar.After(t.start) {
		to = col(t.bar)
	}
	return strings.Repeat(" ", from) + strings.Repeat(fill, max(to, from+1)-from)
}

// axisNames are the d3 codes of axisFormat that print a name.
var axisNames = map[byte]func(time.Time) string{
	'a': func(t time.Time) string { return t.Weekday().String()[:3] },
	'A': func(t time.Time) string { return t.Weekday().String() },
	'b': func(t time.Time) string { return t.Month().String()[:3] },
	'B': func(t time.Time) string { return t.Month().String() },
	'p': func(t time.Time) string {
		if t.Hour() < 12 {
			return "AM"
		}
		return "PM"
	},
	's': func(t time.Time) string { return strconv.FormatInt(t.Unix(), 10) },
	'%': func(time.Time) string { return "%" },
}

// axisNumbers are the d3 codes of axisFormat that print a padded number,
// with their width and padding character.
var axisNumbers = map[byte]struct {
	value func(time.Time) int
	width int
	pad   string
}{
	'Y': {func(t time.Time) int { return t.Year() }, 4, "0"},
	'y': {func(t time.Time) int { return t.Year() % 100 }, 2, "0"},
	'm': {func(t time.Time) int { return int(t.Month()) }, 2, "0"},
	'd': {func(t time.Time) int { return t.Day() }, 2, "0"},
	'e': {func(t time.Time) int { return t.Day() }, 2, " "},
	'H': {func(t time.Time) int { return t.Hour() }, 2, "0"},
	'I': {func(t time.Time) int { return (t.Hour()+11)%12 + 1 }, 2, "0"},
	'M': {func(t time.Time) int { return t.Minute() }, 2, "0"},
	'S': {func(t time.Time) int { return t.Second() }, 2, "0"},
	'L': {func(t time.Time) int { return t.Nanosecond() / 1e6 }, 3, "0"},
	'j': {func(t time.Time) int { return t.YearDay() }, 3, "0"},
}

// formatAxis formats t with the d3 codes of an axisFormat, such as %Y-%m-%d.
// A - after the % drops the padding of a number; any other code is copied as
// written.
func formatAxis(t time.Time, f string) string {
	var b strings.Builder
	for i := 0; i < len(f); i++ {
		j := i + 1
		if f[i] != '%' || j == len(f) {
			b.WriteByte(f[i])
			continue
		}
		pad := true
		if f[j] == '-' && j+1 < len(f) {
			pad, j = false, j+1
		}
		if n, ok := axisNumbers[f[j]]; ok {
			s := strconv.Itoa(n.value(t))
			if pad {
				s = strings.Repeat(n.pad, max(n.width-len(s), 0)) + s
			}
			b.WriteString(s)
		} else if name, ok := axisNames[f[j]]; ok {
			b.WriteString(name(t))
		} else {
			b.WriteString(f[i : j+1])
		}
		i = j
	}
	return b.String()
}
