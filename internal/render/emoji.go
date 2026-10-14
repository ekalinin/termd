package render

import (
	"unicode"

	emast "github.com/yuin/goldmark-emoji/ast"
	"github.com/yuin/goldmark-emoji/definition"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	gtext "github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// emojiParser turns a GitHub emoji shortcode, such as :rocket:, into an
// emoji node of goldmark-emoji. Unlike the parser of goldmark-emoji, it keeps
// a shortcode right after a letter or a digit as text, as GitHub does, so
// a:b:c does not become a🅱️c.
type emojiParser struct {
	emojis definition.Emojis
}

func (p emojiParser) Trigger() []byte {
	return []byte{':'}
}

func (p emojiParser) Parse(parent ast.Node, block gtext.Reader, pc parser.Context) ast.Node {
	if prev := block.PrecendingCharacter(); unicode.IsLetter(prev) || unicode.IsDigit(prev) {
		return nil
	}
	line, _ := block.PeekLine()
	// The characters of a name, as goldmark-emoji reads them.
	i := 1
	for i < len(line) && (util.IsAlphaNumeric(line[i]) || line[i] == '_' || line[i] == '-' || line[i] == '+') {
		i++
	}
	if i == 1 || i >= len(line) || line[i] != ':' {
		return nil
	}
	e, ok := p.emojis.Get(string(line[1:i]))
	if !ok {
		return nil
	}
	block.Advance(i + 1)
	return emast.NewEmoji(line[1:i], e)
}
