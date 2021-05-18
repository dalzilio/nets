// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

//go:generate stringer -type=tokenKind

import "fmt"

type textPos struct {
	line  int
	col   int
	ahead int
}

func (t *textPos) String() string {
	return fmt.Sprintf("line: %d column: %d", t.line+1, t.col-t.ahead)
}

type tokenKind int

// tokenKind is an enumeration describing possible tokens in a net file. tokTR is
// the token for transitions 'tr' in the net format
const (
	tokTR        tokenKind = iota // 'tr'
	tokEOF                        // '\0'
	tokPL                         // 'pl'
	tokNET                        // 'net'
	tokARROW                      // '->'
	tokIDENT                      // identifier [a-Z]([a-Z0-9_'])*
	tokTIMINGC                    // '[a,b]'
	tokINHIBITOR                  // inhibitor arc: '?-1'
	tokREAD                       // read arc: '?1'
	tokLABEL                      // ':'
	tokILLEGAL                    // used to report errors
	tokMARKING                    // initial marking ([0-9]*)
	tokPRIO                       // 'pr'
	tokGT                         // '>' used in priorities
	tokLT                         // '<' used in priorities
	tokSTAR                       // arc multiplicity: '*'
	tokINT                        // integer value, could occur in tpn instruction
	tokNOTE                       // notes can appear when translating from TINA
)

type token struct {
	tok tokenKind
	pos textPos
	s   string
}

func (tok token) String() string {
	return "token (" + fmt.Sprintf("%d", tok.tok) +
		") " + tok.s + fmt.Sprintf(" %v \n", tok.pos)
}

var eof = rune(0)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch == '{') || (ch == '}')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

func isIdentChar(ch rune) bool {
	return (ch == '_') || (ch == '\'') || (ch == '.')
}
