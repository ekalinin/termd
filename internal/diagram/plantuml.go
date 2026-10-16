package diagram

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	mdiagram "github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// errNotSupported is the reason of a PlantUML type that termd does not draw.
var errNotSupported = errors.New("not supported")

// renderPlantUML draws a PlantUML block of the type typ.
func renderPlantUML(typ, src string, width int) (string, error) {
	switch typ {
	case "sequence":
		m, err := translateSequence(src)
		if err != nil {
			return "", err
		}
		return safeRender(m, mdiagram.DefaultConfig())
	case "mindmap":
		title, root, err := parsePlantMindmap(src)
		if err != nil {
			return "", err
		}
		return drawTree(title, root, width), nil
	}
	return "", errNotSupported
}

var startRE = regexp.MustCompile(`^@start([a-z]+)`)

// participantKinds are the keywords that declare a participant of a PlantUML
// sequence diagram.
var participantKinds = []string{"participant", "actor", "boundary", "control", "entity", "database", "collections", "queue"}

// plantumlType returns the type of a PlantUML block: the keyword of its
// @start line, or, for @startuml and a block without one, the type its
// keywords name, as PlantUML picks it. A block with participants or messages
// is a "sequence" diagram; any other block has no type.
func plantumlType(src string) string {
	lines := strings.Split(src, "\n")
	for _, l := range lines {
		if m := startRE.FindStringSubmatch(strings.TrimSpace(l)); m != nil {
			if m[1] != "uml" {
				return m[1]
			}
			break
		}
	}
	seq := false
	for _, l := range lines {
		s := strings.TrimSpace(l)
		switch {
		case hasKeyword(s, "class", "interface", "enum", "abstract"):
			return "class"
		case hasKeyword(s, "usecase"):
			return "usecase"
		case hasKeyword(s, "state") || strings.HasPrefix(s, "[*]"):
			return "state"
		case hasKeyword(s, "component"):
			return "component"
		case hasKeyword(s, "object"):
			return "object"
		case hasKeyword(s, "node"):
			return "deployment"
		case s == "start" || len(s) > 1 && strings.HasPrefix(s, ":") && strings.HasSuffix(s, ";"):
			return "activity"
		}
		if _, ok := parseMessage(s); ok || hasKeyword(s, participantKinds...) {
			seq = true
		}
	}
	if seq {
		return "sequence"
	}
	return ""
}

// hasKeyword reports whether the line s is one of the keywords or starts with
// one of them followed by a space.
func hasKeyword(s string, keywords ...string) bool {
	for _, k := range keywords {
		if s == k || strings.HasPrefix(s, k+" ") {
			return true
		}
	}
	return false
}

// plantMessage is a message line of a PlantUML sequence diagram.
type plantMessage struct {
	from, to, arrow string
	// activation is the mark after the receiver: "++", "--", "**", "!!" or "".
	activation string
	text       string
	hasText    bool
}

// parseMessage reads a line as a message: a sender, an arrow, a receiver,
// optional activation marks and an optional ": text".
func parseMessage(s string) (plantMessage, bool) {
	var m plantMessage
	head := s
	if i := colonOutsideQuotes(s); i >= 0 {
		head, m.text, m.hasText = s[:i], strings.TrimSpace(s[i+1:]), true
	}
	from, rest, ok := cutName(strings.TrimSpace(head))
	if !ok {
		return m, false
	}
	arrow, rest := cutArrow(strings.TrimLeft(rest, " \t"))
	if !strings.Contains(arrow, "-") {
		return m, false
	}
	rest = strings.TrimSpace(rest)
	for _, mark := range []string{"++", "--", "**", "!!"} {
		if r, ok := strings.CutSuffix(rest, mark); ok {
			rest, m.activation = strings.TrimSpace(r), mark
			break
		}
	}
	to, tail, ok := cutName(rest)
	if !ok || strings.TrimSpace(tail) != "" {
		return m, false
	}
	m.from, m.to, m.arrow = from, to, arrow
	return m, true
}

// colonOutsideQuotes returns the index of the first colon of s that is not
// inside double quotes, or -1.
func colonOutsideQuotes(s string) int {
	quoted := false
	for i, r := range s {
		switch {
		case r == '"':
			quoted = !quoted
		case r == ':' && !quoted:
			return i
		}
	}
	return -1
}

// arrowChars are the characters of an arrow besides its x and o
// decorations and its [color].
const arrowChars = "<>-/\\"

// cutName cuts a participant name from the start of s: a quoted string or a
// run of characters that are neither spaces nor arrow characters.
func cutName(s string) (name, rest string, ok bool) {
	if strings.HasPrefix(s, `"`) {
		end := strings.Index(s[1:], `"`)
		if end < 0 {
			return "", s, false
		}
		return s[1 : end+1], s[end+2:], true
	}
	i := strings.IndexFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '[' || r == '"' || strings.ContainsRune(arrowChars, r)
	})
	if i < 0 {
		i = len(s)
	}
	if i == 0 {
		return "", s, false
	}
	return s[:i], s[i:], true
}

// cutArrow cuts an arrow from the start of s. An x or o is part of the arrow
// only next to an arrow character: before one at the start, and after one at
// the end when a space or the end of s follows.
func cutArrow(s string) (arrow, rest string) {
	i := 0
	if len(s) > 1 && (s[0] == 'x' || s[0] == 'o') && strings.ContainsRune(arrowChars, rune(s[1])) {
		i = 1
	}
	for i < len(s) {
		switch {
		case strings.ContainsRune(arrowChars, rune(s[i])):
			i++
		case s[i] == '[':
			end := strings.IndexByte(s[i:], ']')
			if end < 0 {
				return s[:i], s[i:]
			}
			i += end + 1
		default:
			if (s[i] == 'x' || s[i] == 'o') && i > 0 && (i+1 == len(s) || s[i+1] == ' ' || s[i+1] == '\t') {
				i++
			}
			return s[:i], s[i:]
		}
	}
	return s, ""
}

// plantParticipant is a participant of a PlantUML sequence diagram.
type plantParticipant struct {
	// id is the mermaid id; label the name shown when it differs from id.
	id, label string
	actor     bool
	// box is the label of the box around the participant, or "".
	box   string
	inBox bool
}

// sequence translates a PlantUML sequence diagram into mermaid text.
type sequence struct {
	parts  []*plantParticipant
	byName map[string]*plantParticipant
	title  string
	// lastFrom and lastTo are the sender and receiver of the latest message,
	// for notes attached to it.
	lastFrom, lastTo string
	// blocks are the open blocks: "group" or the mermaid keyword.
	blocks []string
	// active are the activations: the activated participant and the sender
	// of the message that activated it.
	active [][2]string
	out    []string
}

var (
	mermaidIDRE  = regexp.MustCompile(`^[\p{L}\p{N}_]+$`)
	dividerRE    = regexp.MustCompile(`^==+\s*(.*?)\s*==+$`)
	refRE        = regexp.MustCompile(`^ref\s+over\s+(.+?)\s*:\s*(.*)$`)
	noteRE       = regexp.MustCompile(`^[hr]?note\s+(left of|right of|over|left|right|across)\b\s*([^:]*?)\s*(?::\s*(.*))?$`)
	blockRE      = regexp.MustCompile(`^(alt|loop|opt|break|critical|par|group)\b\s*(.*)$`)
	elseRE       = regexp.MustCompile(`^else\b\s*(.*)$`)
	activationRE = regexp.MustCompile(`^(activate|deactivate)\s+(.+)$`)
	endNoteRE    = regexp.MustCompile(`^end\s*[hr]?note$`)
)

// translateSequence returns the mermaid sequenceDiagram of a PlantUML
// sequence diagram: the participants in the order of their first
// appearance, declared first, then the translated lines. The first line it
// cannot read is the error "unsupported line N".
func translateSequence(src string) (string, error) {
	sq := &sequence{byName: map[string]*plantParticipant{}}
	lines := strings.Split(src, "\n")
	if err := sq.walk(lines, true); err != nil {
		return "", err
	}
	if err := sq.walk(lines, false); err != nil {
		return "", err
	}
	var b strings.Builder
	if sq.title != "" {
		b.WriteString("---\ntitle: " + strconv.Quote(sq.title) + "\n---\n")
	}
	b.WriteString("sequenceDiagram\n")
	for i, p := range sq.parts {
		if p.inBox && (i == 0 || !sq.parts[i-1].inBox || sq.parts[i-1].box != p.box) {
			b.WriteString("box transparent " + p.box + "\n")
		}
		kind := "participant"
		if p.actor {
			kind = "actor"
		}
		b.WriteString(kind + " " + p.id)
		if p.label != "" {
			b.WriteString(" as " + p.label)
		}
		b.WriteString("\n")
		if p.inBox && (i == len(sq.parts)-1 || !sq.parts[i+1].inBox || sq.parts[i+1].box != p.box) {
			b.WriteString("end\n")
		}
	}
	for _, l := range sq.out {
		b.WriteString(l + "\n")
	}
	return b.String(), nil
}

// walk reads the lines of the diagram. The first pass only collects the
// participants and the title; the second one writes the mermaid lines.
func (sq *sequence) walk(lines []string, collect bool) error {
	box, inBox := "", false
	for i := 0; i < len(lines); i++ {
		s := strings.TrimSpace(lines[i])
		unsupported := fmt.Errorf("unsupported line %d", i+1)
		switch {
		case hasKeyword(s, "legend"), s == "header", s == "footer":
			// A block up to endlegend, endheader or endfooter.
			end := "end" + strings.Fields(s)[0]
			for i < len(lines) && !hasKeyword(strings.ReplaceAll(strings.TrimSpace(lines[i]), "end ", "end"), end) {
				i++
			}
		case s == "" || strings.HasPrefix(s, "'") || strings.HasPrefix(s, "@start") || strings.HasPrefix(s, "@end"),
			strings.HasPrefix(s, "..."), strings.HasPrefix(s, "||"), strings.HasPrefix(s, "!"),
			hasKeyword(s, "hide", "show", "header", "footer", "caption", "newpage", "create", "destroy"):
		case strings.HasPrefix(s, "/'"):
			for i < len(lines) && !strings.HasSuffix(strings.TrimSpace(lines[i]), "'/") {
				i++
			}
		case hasKeyword(s, "skinparam"):
			if strings.HasSuffix(s, "{") {
				for i < len(lines) && strings.TrimSpace(lines[i]) != "}" {
					i++
				}
			}
		case hasKeyword(s, "title"):
			if collect {
				sq.title = sq.text(strings.TrimSpace(strings.TrimPrefix(s, "title")))
			}
		case hasKeyword(s, "box"):
			box, inBox = sq.boxLabel(strings.TrimSpace(strings.TrimPrefix(s, "box"))), true
		case s == "end box":
			box, inBox = "", false
		case hasKeyword(s, participantKinds...):
			if !sq.declare(s, box, inBox) {
				return unsupported
			}
		case hasKeyword(s, "autonumber"):
			if !collect && !strings.HasPrefix(s, "autonumber stop") && !strings.HasPrefix(s, "autonumber resume") {
				sq.out = append(sq.out, "autonumber")
			}
		default:
			if collect {
				sq.collect(s)
				continue
			}
			end, ok := sq.line(lines, i)
			if !ok {
				return unsupported
			}
			i = end
		}
	}
	return nil
}

// collect records the participants that a line names.
func (sq *sequence) collect(s string) {
	if m, ok := parseMessage(s); ok {
		sq.participant(m.from)
		sq.participant(m.to)
		return
	}
	var names string
	if m := refRE.FindStringSubmatch(s); m != nil {
		names = m[1]
	} else if m := noteRE.FindStringSubmatch(s); m != nil {
		names = m[2]
	} else if m := activationRE.FindStringSubmatch(s); m != nil {
		names = m[2]
	}
	for n := range strings.SplitSeq(names, ",") {
		if n = strings.Trim(strings.TrimSpace(n), `"`); n != "" {
			sq.participant(n)
		}
	}
}

// participant returns the participant named name, adding it when it is new.
func (sq *sequence) participant(name string) *plantParticipant {
	if p, ok := sq.byName[name]; ok {
		return p
	}
	p := &plantParticipant{id: name}
	if !mermaidIDRE.MatchString(name) {
		p.id, p.label = "p"+strconv.Itoa(len(sq.parts)+1), name
	}
	sq.parts = append(sq.parts, p)
	sq.byName[name] = p
	return p
}

// declare reads a participant declaration: a kind, a name or a quoted
// label, and an optional alias after "as". Colors, stereotypes and "order"
// are ignored.
func (sq *sequence) declare(s, box string, inBox bool) bool {
	kind, rest, _ := strings.Cut(s, " ")
	first, rest, ok := cutName(strings.TrimSpace(rest))
	if !ok {
		return false
	}
	quoted := strings.HasPrefix(strings.TrimSpace(s[len(kind):]), `"`)
	name, label := first, ""
	if after, ok := strings.CutPrefix(strings.TrimSpace(rest), "as "); ok {
		alias := strings.TrimSpace(after)
		aliasQuoted := strings.HasPrefix(alias, `"`)
		second, _, ok := cutName(alias)
		if !ok {
			return false
		}
		if aliasQuoted && !quoted {
			name, label = first, second
		} else {
			name, label = second, first
		}
	}
	p := sq.participant(name)
	if label != "" && label != p.id {
		p.label = label
	}
	p.actor = kind == "actor"
	p.box, p.inBox = box, inBox
	return true
}

// boxLabel returns the label of a box line: its quoted or plain text without
// a #color.
func (sq *sequence) boxLabel(s string) string {
	if label, _, ok := cutName(s); ok && strings.HasPrefix(s, `"`) {
		return label
	}
	var words []string
	for w := range strings.FieldsSeq(s) {
		if !strings.HasPrefix(w, "#") {
			words = append(words, w)
		}
	}
	return strings.Join(words, " ")
}

// text returns a text of the diagram for mermaid: a \n of PlantUML is a
// space.
func (sq *sequence) text(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, `\n`, " "))
}

// id returns the mermaid id of the participant named name.
func (sq *sequence) id(name string) string {
	return sq.participant(strings.Trim(strings.TrimSpace(name), `"`)).id
}

// overAll returns the participants of a note over all participants.
func (sq *sequence) overAll() string {
	if len(sq.parts) == 0 {
		return ""
	}
	first, last := sq.parts[0].id, sq.parts[len(sq.parts)-1].id
	if first == last {
		return first
	}
	return first + "," + last
}

// note adds a note over all participants, when there are any.
func (sq *sequence) noteOverAll(text string) {
	if all := sq.overAll(); all != "" && text != "" {
		sq.out = append(sq.out, "note over "+all+": "+text)
	}
}

// line translates the line i and returns the index of its last line: a note
// can continue on the next lines.
func (sq *sequence) line(lines []string, i int) (int, bool) {
	s := strings.TrimSpace(lines[i])
	if m, ok := parseMessage(s); ok {
		sq.message(m)
		return i, true
	}
	switch {
	case s == "end":
		if len(sq.blocks) == 0 {
			return i, false
		}
		top := sq.blocks[len(sq.blocks)-1]
		sq.blocks = sq.blocks[:len(sq.blocks)-1]
		if top != "group" {
			sq.out = append(sq.out, "end")
		}
	case hasKeyword(s, "return"):
		if len(sq.active) == 0 {
			return i, false
		}
		a := sq.active[len(sq.active)-1]
		sq.active = sq.active[:len(sq.active)-1]
		sq.out = append(sq.out, a[0]+"-->>-"+a[1]+": "+sq.text(strings.TrimPrefix(s, "return")))
	case dividerRE.MatchString(s):
		sq.noteOverAll(sq.text(dividerRE.FindStringSubmatch(s)[1]))
	case refRE.MatchString(s):
		m := refRE.FindStringSubmatch(s)
		sq.out = append(sq.out, "note over "+sq.ids(m[1])+": ref "+sq.text(m[2]))
	case noteRE.MatchString(s):
		return sq.note(lines, i, noteRE.FindStringSubmatch(s))
	case blockRE.MatchString(s):
		m := blockRE.FindStringSubmatch(s)
		sq.blocks = append(sq.blocks, m[1])
		if m[1] == "group" {
			sq.noteOverAll(sq.text(m[2]))
		} else {
			sq.out = append(sq.out, strings.TrimSpace(m[1]+" "+sq.text(m[2])))
		}
	case elseRE.MatchString(s):
		if len(sq.blocks) == 0 {
			return i, false
		}
		kw := "else"
		if sq.blocks[len(sq.blocks)-1] == "par" {
			kw = "and"
		}
		sq.out = append(sq.out, strings.TrimSpace(kw+" "+sq.text(elseRE.FindStringSubmatch(s)[1])))
	case activationRE.MatchString(s):
		m := activationRE.FindStringSubmatch(s)
		id := sq.id(m[2])
		if m[1] == "activate" {
			sq.active = append(sq.active, [2]string{id, sq.lastFrom})
		} else if n := len(sq.active); n > 0 && sq.active[n-1][0] == id {
			sq.active = sq.active[:n-1]
		}
		sq.out = append(sq.out, m[1]+" "+id)
	default:
		return i, false
	}
	return i, true
}

// ids returns the mermaid ids of a comma-separated list of names.
func (sq *sequence) ids(names string) string {
	var ids []string
	for n := range strings.SplitSeq(names, ",") {
		ids = append(ids, sq.id(n))
	}
	return strings.Join(ids, ",")
}

// message translates a message. An arrow that points only to the left swaps
// the participants; -- makes it dotted and >> or << asynchronous.
func (sq *sequence) message(m plantMessage) {
	from, to := sq.id(m.from), sq.id(m.to)
	arrow := strings.Trim(m.arrow, "xo")
	if i := strings.IndexByte(arrow, '['); i >= 0 {
		if j := strings.IndexByte(arrow[i:], ']'); j >= 0 {
			arrow = arrow[:i] + arrow[i+j+1:]
		}
	}
	if strings.HasPrefix(arrow, "<") && !strings.HasSuffix(arrow, ">") {
		from, to = to, from
	}
	head := ">>"
	if strings.Contains(arrow, ">>") || strings.Contains(arrow, "<<") {
		head = ")"
	}
	line := "-" + head
	if strings.Contains(arrow, "--") {
		line = "--" + head
	}
	switch m.activation {
	case "++":
		line += "+"
		sq.active = append(sq.active, [2]string{to, from})
	case "--":
		line += "-"
		if n := len(sq.active); n > 0 && sq.active[n-1][0] == from {
			sq.active = sq.active[:n-1]
		}
	}
	sq.out = append(sq.out, from+line+to+": "+sq.text(m.text))
	sq.lastFrom, sq.lastTo = from, to
}

// note translates a note; one without ": text" continues up to "end note",
// and its lines are joined with spaces.
func (sq *sequence) note(lines []string, i int, m []string) (int, bool) {
	where, names, text := m[1], m[2], m[3]
	if !strings.Contains(lines[i], ":") {
		var parts []string
		for i++; i < len(lines) && !endNoteRE.MatchString(strings.TrimSpace(lines[i])); i++ {
			parts = append(parts, strings.TrimSpace(lines[i]))
		}
		if i == len(lines) {
			return i, false
		}
		text = strings.Join(parts, " ")
	}
	text = sq.text(text)
	switch where {
	case "left of", "right of", "over":
		if names == "" {
			return i, false
		}
		sq.out = append(sq.out, "note "+where+" "+sq.ids(names)+": "+text)
	case "left", "right":
		target := sq.lastFrom
		if where == "right" {
			target = sq.lastTo
		}
		if target == "" {
			return i, false
		}
		sq.out = append(sq.out, "note "+where+" of "+target+": "+text)
	case "across":
		sq.noteOverAll(text)
	}
	return i, true
}

// plantNodeRE matches a node line of a PlantUML mindmap: the depth
// characters, an optional [#color], an optional _ for a node without a box,
// and the text.
var plantNodeRE = regexp.MustCompile(`^(\*+|\++|-+)(\[[^\]]*\])?(_)?\s*(.*)$`)

// parsePlantMindmap returns the title and the root of a PlantUML mindmap.
// The depth of a node is the number of its *, + or - characters, and a node
// is a child of the nearest node before it with a smaller depth.
func parsePlantMindmap(src string) (string, *mindNode, error) {
	lines := strings.Split(src, "\n")
	type open struct {
		node  *mindNode
		depth int
	}
	var title string
	var root *mindNode
	var stack []open
	for i := 0; i < len(lines); i++ {
		s := strings.TrimSpace(lines[i])
		switch {
		case s == "" || strings.HasPrefix(s, "'") || strings.HasPrefix(s, "@start") || strings.HasPrefix(s, "@end") || s == "left side" || s == "right side":
			continue
		case hasKeyword(s, "title"):
			title = strings.TrimSpace(strings.TrimPrefix(s, "title"))
			continue
		}
		m := plantNodeRE.FindStringSubmatch(s)
		if m == nil {
			return "", nil, fmt.Errorf("unsupported line %d", i+1)
		}
		depth, txt := len(m[1]), m[4]
		if body, ok := strings.CutPrefix(txt, ":"); ok {
			// A multi-line text, up to a line that ends with ";".
			for !strings.HasSuffix(body, ";") {
				if i++; i >= len(lines) {
					return "", nil, errors.New("unclosed node")
				}
				body += "\n" + strings.TrimSpace(lines[i])
			}
			txt = strings.TrimSuffix(body, ";")
		}
		parts := strings.Split(strings.ReplaceAll(txt, `\n`, "\n"), "\n")
		for j, p := range parts {
			parts[j] = strings.TrimSpace(p)
		}
		for len(stack) > 0 && stack[len(stack)-1].depth >= depth {
			stack = stack[:len(stack)-1]
		}
		n := &mindNode{text: parts}
		switch {
		case root == nil:
			root = n
		case len(stack) == 0:
			return "", nil, errors.New("more than one root")
		default:
			parent := stack[len(stack)-1].node
			parent.children = append(parent.children, n)
		}
		stack = append(stack, open{n, depth})
	}
	if root == nil {
		return "", nil, errors.New("no nodes")
	}
	return title, root, nil
}
