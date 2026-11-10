package diagram

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/ekalinin/termd/internal/text"
)

// stateKind tells plain states from the states drawn as a symbol.
type stateKind int

const (
	statePlain stateKind = iota
	stateStart
	stateEnd
	stateChoice
	stateBar // a fork or a join
)

// state is a state of a state diagram. A composite state has regions; the
// root state of a diagram has the top region.
type state struct {
	id      string
	desc    []string
	kind    stateKind
	parent  *region
	regions []*region
	// dir is the direction of the diagrams of the regions, "" to inherit it.
	dir string
}

// region is a part of a diagram drawn as one flowchart: the top of the
// diagram or a concurrent region of a composite state.
type region struct {
	owner  *state
	states []*state
	// start and end are the [*] of the region, nil when it has none.
	start, end *state
}

// transition is an arrow between two states.
type transition struct {
	from, to *state
	label    string
}

// stateNote is a note of a state, or of a transition when of is nil.
type stateNote struct {
	of   *state
	link *transition
	text string
}

// stateDiagram is a parsed mermaid or PlantUML state diagram.
type stateDiagram struct {
	title       string
	root        *state
	transitions []*transition
	notes       []stateNote
	ids         map[string]*state
	// stack holds the open regions, the current one last, and opened the
	// numbers of the lines that opened the composite states on it.
	stack  []*region
	opened []int
}

func newStateDiagram() *stateDiagram {
	root := &state{}
	top := &region{owner: root}
	root.regions = []*region{top}
	return &stateDiagram{root: root, ids: map[string]*state{}, stack: []*region{top}}
}

// errNoStates is the reason of a state diagram without states.
var errNoStates = errors.New("no states")

func unsupportedLine(n int) error { return fmt.Errorf("unsupported line %d", n) }

func unclosedBlock(n int) error { return fmt.Errorf("unclosed block on line %d", n) }

// current returns the region of the line being read.
func (d *stateDiagram) current() *region { return d.stack[len(d.stack)-1] }

// state returns the state id, adding it to the current region when it is
// new.
func (d *stateDiagram) state(id string) *state {
	if s, ok := d.ids[id]; ok {
		return s
	}
	r := d.current()
	s := &state{id: id, parent: r}
	r.states = append(r.states, s)
	d.ids[id] = s
	return s
}

// end returns the state named by an end of a transition: the start or the
// end of the current region for [*], or a state.
func (d *stateDiagram) end(id string, source bool) *state {
	if id != "[*]" {
		return d.state(id)
	}
	r := d.current()
	if source {
		if r.start == nil {
			r.start = &state{kind: stateStart, parent: r}
		}
		return r.start
	}
	if r.end == nil {
		r.end = &state{kind: stateEnd, parent: r}
	}
	return r.end
}

// open starts the block of the composite state s on line n.
func (d *stateDiagram) open(s *state, n int) {
	if len(s.regions) == 0 {
		s.regions = []*region{{owner: s}}
	}
	d.stack = append(d.stack, s.regions[len(s.regions)-1])
	d.opened = append(d.opened, n)
}

// split starts a new concurrent region of the current composite state.
func (d *stateDiagram) split() bool {
	if len(d.stack) == 1 {
		return false
	}
	owner := d.current().owner
	r := &region{owner: owner}
	owner.regions = append(owner.regions, r)
	d.stack[len(d.stack)-1] = r
	return true
}

// close ends the block of the current composite state.
func (d *stateDiagram) close() bool {
	if len(d.stack) == 1 {
		return false
	}
	d.stack = d.stack[:len(d.stack)-1]
	d.opened = d.opened[:len(d.opened)-1]
	return true
}

// finish checks that every block is closed and that there are states.
func (d *stateDiagram) finish() error {
	if len(d.opened) > 0 {
		return unclosedBlock(d.opened[len(d.opened)-1])
	}
	if len(d.ids) == 0 && len(d.transitions) == 0 {
		return errNoStates
	}
	return nil
}

var (
	stateIDRE        = regexp.MustCompile(`^[^\s:]+$`)
	stateClassRE     = regexp.MustCompile(`(\S):::[\w-]+`)
	stateDescRE      = regexp.MustCompile(`^state\s+"([^"]*)"\s+as\s+([^\s:{]+)\s*(\{)?$`)
	stateSpecialRE   = regexp.MustCompile(`^state\s+([^\s:{]+)\s*(?:<<(choice|fork|join)>>|\[\[(choice|fork|join)\]\])$`)
	stateDeclareRE   = regexp.MustCompile(`^state\s+([^\s:{]+)\s*(\{)?$`)
	stateDirectionRE = regexp.MustCompile(`^direction\s+(TB|TD|BT|LR|RL)$`)
	stateNoteRE      = regexp.MustCompile(`^note\s+(?:left|right)\s+of\s+([^\s:]+)\s*(?::(.*))?$`)
)

// specialKinds maps the stereotypes of special states to their kinds.
var specialKinds = map[string]stateKind{"choice": stateChoice, "fork": stateBar, "join": stateBar}

// parseMermaidState reads a mermaid state diagram.
func parseMermaidState(src string) (*stateDiagram, error) {
	lines := strings.Split(src, "\n")
	h := headerLine(lines)
	d := newStateDiagram()
	d.title = frontmatterTitle(lines[:h])
	for i := h + 1; i < len(lines); i++ {
		n := i + 1
		s := strings.TrimSpace(stateClassRE.ReplaceAllString(lines[i], "$1"))
		switch {
		case s == "" || strings.HasPrefix(s, "%%") || hasKeyword(s, "classDef", "class", "style") ||
			strings.HasPrefix(s, "accTitle:") || strings.HasPrefix(s, "accDescr:"):
		case s == "accDescr {" || s == "accDescr{":
			for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "}"; i++ {
			}
			if i == len(lines) {
				return nil, unclosedBlock(n)
			}
		case s == "}":
			if !d.close() {
				return nil, unsupportedLine(n)
			}
		case s == "--":
			if !d.split() {
				return nil, unsupportedLine(n)
			}
		case stateDirectionRE.MatchString(s):
			d.current().owner.dir = stateDirectionRE.FindStringSubmatch(s)[1]
		case strings.HasPrefix(s, "note "):
			m := stateNoteRE.FindStringSubmatch(s)
			if m == nil {
				return nil, unsupportedLine(n)
			}
			text, next, ok := noteText(m[2], strings.Contains(s, ":"), lines, i)
			if !ok {
				return nil, unclosedBlock(n)
			}
			d.notes = append(d.notes, stateNote{of: d.state(m[1]), text: text})
			i = next
		case strings.HasPrefix(s, "state "):
			if !d.declare(s, n) {
				return nil, unsupportedLine(n)
			}
		default:
			if !d.transitionOrDescription(s) {
				return nil, unsupportedLine(n)
			}
		}
	}
	return d, d.finish()
}

// declare reads a state line of a mermaid state diagram.
func (d *stateDiagram) declare(s string, n int) bool {
	if m := stateDescRE.FindStringSubmatch(s); m != nil {
		st := d.state(m[2])
		st.desc = append(st.desc, m[1])
		if m[3] != "" {
			d.open(st, n)
		}
		return true
	}
	if m := stateSpecialRE.FindStringSubmatch(s); m != nil {
		d.state(m[1]).kind = specialKinds[m[2]+m[3]]
		return true
	}
	if m := stateDeclareRE.FindStringSubmatch(s); m != nil {
		st := d.state(m[1])
		if m[2] != "" {
			d.open(st, n)
		}
		return true
	}
	return false
}

// transitionOrDescription reads a transition, an "id : text" description
// or a bare id.
func (d *stateDiagram) transitionOrDescription(s string) bool {
	head, label, hasLabel := strings.Cut(s, ":")
	if from, to, ok := strings.Cut(head, "-->"); ok {
		from, to = strings.TrimSpace(from), strings.TrimSpace(to)
		if !stateIDRE.MatchString(from) || !stateIDRE.MatchString(to) {
			return false
		}
		d.transitions = append(d.transitions, &transition{
			from:  d.end(from, true),
			to:    d.end(to, false),
			label: strings.TrimSpace(label),
		})
		return true
	}
	id := strings.TrimSpace(head)
	if !stateIDRE.MatchString(id) || id == "[*]" {
		return false
	}
	st := d.state(id)
	if text := strings.TrimSpace(label); hasLabel && text != "" {
		st.desc = append(st.desc, text)
	}
	return true
}

// noteText returns the text of a note: the text after its colon, or the
// lines after the line i up to "end note" joined with spaces, with the index
// of the last line read. ok is false when "end note" is missing.
func noteText(inline string, hasColon bool, lines []string, i int) (text string, last int, ok bool) {
	if hasColon {
		return strings.TrimSpace(inline), i, true
	}
	var parts []string
	for j := i + 1; j < len(lines); j++ {
		l := strings.TrimSpace(lines[j])
		if strings.EqualFold(l, "end note") {
			return strings.Join(parts, " "), j, true
		}
		if l != "" {
			parts = append(parts, l)
		}
	}
	return "", i, false
}

// renderState draws a mermaid state diagram.
func renderState(src string, width int, g *glyphs) (string, error) {
	d, err := parseMermaidState(src)
	if err != nil {
		return "", err
	}
	return d.draw(width, g)
}

// draw draws the diagram for an output width: the title, the top diagram,
// and a section for every composite state with states inside it.
func (d *stateDiagram) draw(width int, g *glyphs) (string, error) {
	var out []string
	if d.title != "" {
		out = append(out, d.title, "")
	}
	out, err := d.section(out, d.root, width, g)
	return strings.Join(out, "\n"), err
}

// section appends to out the diagrams of the regions of s with their notes,
// under the text of s unless s is the root, and then the sections of the
// composite states inside s.
func (d *stateDiagram) section(out []string, s *state, width int, g *glyphs) ([]string, error) {
	if s != d.root {
		out = append(out, "")
		out = append(out, wrapHanging(stateText(s, g), width)...)
		out = append(out, "")
	}
	var parts [][]string
	widest := 0
	for _, r := range s.regions {
		if r.empty() {
			continue
		}
		fc, err := renderFlowchart(d.flowchart(r, g), width, g)
		if err != nil {
			return nil, err
		}
		widest = max(widest, blockWidth(fc))
		parts = append(parts, append(strings.Split(fc, "\n"), d.noteLines(r, width, g)...))
	}
	for i, p := range parts {
		if i > 0 {
			out = append(out, "", strings.Repeat(g.region, widest), "")
		}
		out = append(out, p...)
	}
	for _, r := range s.regions {
		for _, c := range r.states {
			if c.hasInside() {
				var err error
				if out, err = d.section(out, c, width, g); err != nil {
					return nil, err
				}
			}
		}
	}
	return out, nil
}

// empty reports whether a region has nothing to draw.
func (r *region) empty() bool {
	return len(r.states) == 0 && r.start == nil && r.end == nil
}

// hasInside reports whether s is a composite state with states inside it.
func (s *state) hasInside() bool {
	for _, r := range s.regions {
		if !r.empty() {
			return true
		}
	}
	return false
}

// chain returns the regions from the top of the diagram down to the region
// of s.
func chain(s *state) []*region {
	var c []*region
	for r := s.parent; r != nil; r = r.owner.parent {
		c = append(c, r)
	}
	slices.Reverse(c)
	return c
}

// ancestorIn returns s, or the composite state that contains s, in r.
func ancestorIn(s *state, r *region) *state {
	for s.parent != r {
		s = s.parent.owner
	}
	return s
}

// innermost returns the innermost region that has both ends of t, as states
// of it or inside its composite states, and the states of that region that
// are or contain the ends.
func innermost(t *transition) (*region, *state, *state) {
	a, b := chain(t.from), chain(t.to)
	i := 1
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	r := a[i-1]
	return r, ancestorIn(t.from, r), ancestorIn(t.to, r)
}

// noteLines returns the notes of the diagram of r, after a blank line.
func (d *stateDiagram) noteLines(r *region, width int, g *glyphs) []string {
	var out []string
	for _, n := range d.notes {
		var prefix string
		if n.of != nil {
			if n.of.parent != r {
				continue
			}
			prefix = stateText(n.of, g)
		} else {
			if in, _, _ := innermost(n.link); in != r {
				continue
			}
			prefix = stateText(n.link.from, g) + " --> " + stateText(n.link.to, g)
		}
		if len(out) == 0 {
			out = append(out, "")
		}
		out = append(out, wrapHanging(prefix+": "+brRE.ReplaceAllString(n.text, " "), width)...)
	}
	return out
}

// stateText returns the text of a state on one line.
func stateText(s *state, g *glyphs) string {
	return strings.Join(strings.Fields(brRE.ReplaceAllString(strings.Join(nodeLines(s, g), " "), " ")), " ")
}

// wrapHanging wraps s at whitespace to width; the lines after the first
// start with 2 spaces. Words are never split.
func wrapHanging(s string, width int) []string {
	var out []string
	line := ""
	for _, w := range strings.Fields(s) {
		next := w
		if line != "" {
			next = line + " " + w
		}
		if line != "" && text.Width(next) > width {
			out = append(out, line)
			next = "  " + w
		}
		line = next
	}
	return append(out, line)
}

// flowchart returns the mermaid flowchart of a region: a node for its start,
// its states and its end, and an edge for every transition of which this is
// the innermost region. Nodes get generated ids, so that no state id is read
// as flowchart syntax.
func (d *stateDiagram) flowchart(r *region, g *glyphs) string {
	var b strings.Builder
	fmt.Fprintf(&b, "flowchart %s\n", d.direction(r))
	ids := map[*state]string{}
	node := func(s *state) {
		if s == nil {
			return
		}
		ids[s] = fmt.Sprintf("n%d", len(ids)+1)
		fmt.Fprintf(&b, "    %s[\"%s\"]\n", ids[s], breakLinks(strings.Join(nodeLines(s, g), "<br>"), g.ascii))
	}
	node(r.start)
	for _, s := range r.states {
		node(s)
	}
	node(r.end)
	for _, t := range d.transitions {
		in, a, z := innermost(t)
		if in != r {
			continue
		}
		from, to := ids[a], ids[z]
		if label := strings.TrimSpace(brRE.ReplaceAllString(t.label, " ")); label != "" {
			fmt.Fprintf(&b, "    %s -->|%s| %s\n", from, label, to)
		} else {
			fmt.Fprintf(&b, "    %s --> %s\n", from, to)
		}
	}
	return b.String()
}

// direction returns the direction of the diagram of a region: the direction
// of its composite state, or else of the nearest composite state around it
// that has one, or TB.
func (d *stateDiagram) direction(r *region) string {
	for s := r.owner; ; s = s.parent.owner {
		if s.dir != "" {
			return s.dir
		}
		if s.parent == nil {
			return "TB"
		}
	}
}

// symbols are the texts of the states drawn as a symbol.
var symbols = map[stateKind]string{stateStart: "●", stateEnd: "◉", stateChoice: "◇", stateBar: "━━━"}

// nodeLines returns the lines of the text of a state: its symbol, the lines
// of its description, or its id.
func nodeLines(s *state, g *glyphs) []string {
	if sym, ok := g.symbols[s.kind]; ok {
		return []string{sym}
	}
	if len(s.desc) > 0 {
		return s.desc
	}
	return []string{s.id}
}

// breakLinks replaces the characters of flowchart links in a node text with
// look-alikes of the same width: a - after a - with a non-breaking hyphen,
// and a > after a - or a = with a single right-pointing angle quotation
// mark. In ASCII mode the characters are kept with a space before them.
func breakLinks(s string, ascii bool) string {
	r := []rune(s)
	var b strings.Builder
	for i, c := range r {
		link := i > 0 && (c == '-' && r[i-1] == '-' || c == '>' && (r[i-1] == '-' || r[i-1] == '='))
		switch {
		case !link:
			b.WriteRune(c)
		case ascii:
			b.WriteString(" " + string(c))
		case c == '-':
			b.WriteRune('‑')
		default:
			b.WriteRune('›')
		}
	}
	return b.String()
}
