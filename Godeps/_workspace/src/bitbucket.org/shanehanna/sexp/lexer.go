package sexp

import (
	"fmt"
	"regexp"
	"strconv"
)

type Item struct {
	Type     ItemType
	Position int
	Value    []byte
}

type ItemType int

const (
	ItemError        ItemType = iota
	ItemBracketLeft           // (
	ItemBracketRight          // )
	ItemToken                 // abc        Token.
	ItemQuote                 // "abc"      Quoted string. May also include length 3"abc"
	ItemVerbatim              // 3:abc      Length prefixed "verbatim" encoding.
	// ItemHex                // #616263#   Hexidecimal string.
	// ItemBase64             // {MzphYmM=} Base64 of the verbatim encoding "3:abc"
	// ItemBase64Octet        // |YWJj|     Base64 encoding of the octet-string "abc"
	ItemEOF
)

var (
	reBracketLeft  = regexp.MustCompile(`^\(`)
	reBracketRight = regexp.MustCompile(`^\)`)
	reWhitespace   = regexp.MustCompile(`^\s+`)
	reVerbatim     = regexp.MustCompile(`^(\d+):`)
	reQuote        = regexp.MustCompile(`^(\d+)?"((?:[^\\"]|\\.)*)"`)

	// Strict(er) R.Rivset 1997 draft token + unicode letter support (hello 1997).
	// reToken     = regexp.MustCompile(`^[\p{L}][\p{L}\p{N}\-./_:*+=]+`)
	// Instead a token can be anything including '(', ')' and ' ' so long as you escape them:
	reToken = regexp.MustCompile(`^(?:[^\\ ()]|\\.)+`)
)

type stateFn func(*lexer) stateFn

type lexer struct {
	input   []byte
	items   chan Item
	start   int
	pos     int
	state   stateFn
	matches [][]byte
}

func (l *lexer) emit(t ItemType) {
	l.items <- Item{t, l.start, l.input[l.start:l.pos]}
}

func (l *lexer) Next() Item {
	item := <-l.items
	return item
}

func (l *lexer) scan(re *regexp.Regexp) bool {
	if l.match(re) {
		l.start = l.pos
		l.pos += len(l.matches[0])
		return true
	}
	return false
}

func (l *lexer) match(re *regexp.Regexp) bool {
	if l.matches = re.FindSubmatch(l.input[l.pos:]); l.matches != nil {
		return true
	}
	return false
}

func (l *lexer) run() {
	for l.state = lex; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{ItemError, l.start, []byte(fmt.Sprintf(format, args...))}
	return nil
}

func lex(l *lexer) stateFn {
	// The order is important here, reToken must come last because it'll match reVerbatim and
	// reQuote atoms as well.
	switch {
	case l.pos >= len(l.input):
		l.emit(ItemEOF)
		return nil
	case l.scan(reWhitespace):
		return lex
	case l.scan(reBracketLeft):
		l.emit(ItemBracketLeft)
		return lex
	case l.scan(reBracketRight):
		l.emit(ItemBracketRight)
		return lex
	case l.scan(reQuote):
		// TODO: errorf if length exists and doesn't line up with quote length.
		// Don't include quotes in Value.
		l.items <- Item{ItemQuote, l.start, []byte(l.matches[2])}
		return lex
	case l.scan(reVerbatim):
		bytes, _ := strconv.ParseInt(string(l.matches[1]), 10, 64)
		l.start = l.pos
		l.pos += int(bytes)
		l.emit(ItemVerbatim)
		return lex
	case l.scan(reToken):
		l.emit(ItemToken)
		return lex
	}

	// TODO: Read number of runes. Reading 10 bytes may leave the last unprintable.
	near := l.input[l.pos:]
	if len(near) < 10 {
		near = near[:len(near)]
	} else {
		near = near[:10]
	}
	return l.errorf("Unexpected byte at %d near '%s'.", l.pos, near)
}

/*
  Lex S-Expressions.

  See http://people.csail.mit.edu/rivest/Sexp.txt

  * Unlike the R.Rivest 1997 draft tokens will match any unicode letters.
  * Canonical S-Expressions may have spaces between atoms which isn't strictly correct.
*/
func NewLexer(input []byte) *lexer {
	l := &lexer{input: input, items: make(chan Item)}
	go l.run()
	return l
}
