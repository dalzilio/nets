// Copyright 2026. Silvano DAL ZILIO. All rights reserved.
// Use of this source code is governed by the AGPL license
// that can be found in the LICENSE file.

package nets

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
)

//
// Parsing and evaluation of atomic propositions on markings.
//
// form	::=	( form )  (* parenthesized expression *)
// 	|	- form		  (* negation *)
// 	|	form op form  (* conjunction (/\) or disjunction (\/), => and <=> *)
// 	|	rexp cp rexp  (* where cp is one of >, >=, <, <=, = *)
// rexp ::= id        (* name of place *)
//  |   constant      (* positive integer constant *)
//  |   rexp + rexp   (* addition, note that expressions cannot contain parentheses *)
// id	::=	aname | qname

// Formula is the interface (AST) for atomic state propositions on markings. The
// verdict of a formula is computed with respect to a marking using the
// Eval method.
type Formula interface {
	// Eval checks if the formula is true for the marking.
	Eval(Marking) bool
	fmt.Stringer
}

// formOp is for binary operations between formulas: and, or, imply, equiv
type formOp struct {
	op          formTokenKind
	left, right Formula
}

func (f formOp) Eval(m Marking) bool {
	switch f.op {
	case formulaAND:
		return (f.left).Eval(m) && (f.right).Eval(m)
	case formulaOR:
		return (f.left).Eval(m) || (f.right).Eval(m)
	case formulaIMPLY:
		return !(f.left).Eval(m) || (f.right).Eval(m)
	case formulaEQUIV:
		return ((f.left).Eval(m) && (f.right).Eval(m)) || (!(f.left).Eval(m) && !(f.right).Eval(m))
	}
	log.Fatalf("unauthorized operation in formula (%d)", f.op)
	return false
}

func (f formOp) String() string {
	switch f.op {
	case formulaAND:
		return fmt.Sprintf("(%s ∧ %s)", f.left, f.right)
	case formulaOR:
		return fmt.Sprintf("(%s ∨ %s)", f.left, f.right)
	case formulaIMPLY:
		return fmt.Sprintf("(%s ⇒ %s)", f.left, f.right)
	case formulaEQUIV:
		return fmt.Sprintf("(%s ⇔ %s)", f.left, f.right)
	}
	return fmt.Sprintf("<not an operator>(%d)", f.op)
}

// formAtom, atomic formula, can only be a comparaison (using >, >=, <, <=, or
// =) between two expressions.
type formAtom struct {
	op          formTokenKind
	left, right expression
}

func (f formAtom) Eval(m Marking) bool {
	switch f.op {
	case formulaEQ:
		return (f.left).value(m) == (f.right).value(m)
	case formulaLTHAN:
		return (f.left).value(m) < (f.right).value(m)
	case formulaLEQ:
		return (f.left).value(m) <= (f.right).value(m)
	case formulaGTHAN:
		return (f.left).value(m) > (f.right).value(m)
	case formulaGEQ:
		return (f.left).value(m) >= (f.right).value(m)
	}
	log.Fatalf("unauthorized operation between expressions (%d)", f.op)
	return false
}

func (f formAtom) String() string {
	switch f.op {
	case formulaEQ:
		return fmt.Sprintf("%s = %s", f.left, f.right)
	case formulaLTHAN:
		return fmt.Sprintf("%s < %s", f.left, f.right)
	case formulaLEQ:
		return fmt.Sprintf("%s ≤ %s", f.left, f.right)
	case formulaGTHAN:
		return fmt.Sprintf("%s > %s", f.left, f.right)
	case formulaGEQ:
		return fmt.Sprintf("%s ≥ %s", f.left, f.right)
	}
	return fmt.Sprintf("<not a comparaison>(%d)", f.op)
}

// formNeg is for negation of a formula
type formNeg struct {
	exp Formula
}

func (f formNeg) Eval(m Marking) bool {
	return !(f.exp.Eval(m))
}

func (f formNeg) String() string {
	return fmt.Sprintf("- (%s)", f.exp)
}

// *************************************************************

// expression stands for a sum of places and constants used inside state
// formulas. We keep a (multi)set of places and a constant, but also the text
// associated with the initial expression (used for the Stringer method). We
// support repetitions of the same place, and constant are added, since
// expressions of the form (p + 1 + q + 3 + p) are accepted.
type expression struct {
	names    []string
	constant int
	places   []int
}

func (e *expression) value(m Marking) int {
	r := e.constant
	for _, v := range e.places {
		r += m.Get(v)
	}
	return r
}

func (e expression) String() string {
	if len(e.places) == 0 {
		return strconv.Itoa(e.constant)
	}
	r := e.names[0]
	for i := 1; i < len(e.places); i++ {
		r += " + " + e.names[i]
	}
	if e.constant != 0 {
		r += " + " + strconv.Itoa(e.constant)
	}
	return r
}

// *************************************************************

type formTokenKind int

const (
	formulaAND    formTokenKind = iota // is '/\'
	formulaOR                          // '\/'
	formulaNEG                         // '-'
	formulaID                          // identifier
	formulaCST                         // positive integer constant
	formulaIMPLY                       // '=>'
	formulaEQUIV                       // '<=>'
	formulaPLUS                        // '+'
	formulaLTHAN                       // '<'
	formulaLEQ                         // '<='
	formulaGTHAN                       // '>'
	formulaGEQ                         // '>='
	formulaEQ                          // '='
	formulaEOF                         // end of formula
	formulaLPAREN                      // '('
	formulaRPAREN                      // ')'
	formulaERROR                       // to report illegal tokens
)

type formToken struct {
	tok formTokenKind
	s   string
}

type formParser struct {
	net      *Net            // net used to find the names of places
	soff     int             // starting offset in the Reader (for reporting errors)
	eoff     int             // ending offset in the Reader
	form     []byte          // the string we try to parse
	reader   *strings.Reader // a Reader to the formula we need to parse
	buff     strings.Builder // a buffer where to read identifiers
	pardepth int             // depth of parentheses
	tok      formToken       // last read token
	ahead    bool            // true if there is a token stored in tok
}

// ParseFormula returns a Formula from a string representing a (marking) state
// formulas.
func (net *Net) ParseFormula(in string) (Formula, error) {
	p := &formParser{
		net:    net,
		form:   []byte(in),
		reader: strings.NewReader(in),
		ahead:  false,
	}
	f, err := p.parseFormula(0)
	if (f == nil) && (err == nil) {
		return f, fmt.Errorf("empty formula")
	}
	return f, err
}

// parseFormula parses until we find a matching, closing parentheses (using the
// value of the depth parameter), or we find EOF (and we started parsing at
// depth zero).
func (s *formParser) parseFormula(depth int) (Formula, error) {
	var f Formula
	var err error
	for {
		tok := s.scan()
		switch tok.tok {
		case formulaEOF:
			if s.pardepth != 0 {
				return f, fmt.Errorf("mismatched parentheses")
			}
			return f, nil
		case formulaERROR:
			// We receive an error from the lexer
			return f, fmt.Errorf("%s at %s", tok.s, s.problems())
		case formulaIMPLY, formulaEQUIV, formulaAND, formulaOR:
			// We find a connector, so there must be a left hand-side.
			if f == nil {
				return f, fmt.Errorf("left-hand side of connector missing: %s", s.problems())
			}
			fleft := f
			s.soff = s.eoff
			fright, err := s.parseFormula(s.pardepth)
			if err != nil {
				return f, err
			}
			s.soff = s.eoff
			f = formOp{op: tok.tok, left: fleft, right: fright}
		case formulaNEG:
			// A negation should only occur at the beginning of a new formula
			// context, e.g. after a connector or a left parenthesis.
			if f != nil {
				return f, fmt.Errorf("wrong left-hand side for negation: %s", s.problems())
			}
			f, err = s.parseFormula(s.pardepth)
			if err != nil {
				return f, err
			}
			s.soff = s.eoff
			f = formNeg{exp: f}
		case formulaEQ, formulaGEQ, formulaGTHAN, formulaLEQ, formulaLTHAN:
			return f, fmt.Errorf("comparaison operator outside expression: %s", s.problems())
		case formulaID, formulaCST:
			// Parse the expression starting from here and add it has an atomic
			// formula (formAtom). We check if we are higher than the current
			// depth in order to catch mismatched parentheses.
			s.unscan()
			f, err = s.parseExpression()
			if err != nil {
				return f, err
			}
		case formulaLPAREN:
			if f != nil {
				return f, fmt.Errorf("Unexpected token before '(': %s", s.problems())
			}
			s.pardepth++
			s.soff = s.eoff
			f, err = s.parseFormula(s.pardepth)
			if f == nil {
				return f, fmt.Errorf("empty formula")
			}
			if err != nil {
				return f, err
			}
		case formulaRPAREN:
			if s.pardepth == 0 {
				return f, fmt.Errorf("mismatched parenthesis %s", s.problems())
			}
			s.pardepth--
			s.soff = s.eoff
			if s.pardepth < depth {
				return f, nil
			}
		default:
			return f, fmt.Errorf("error in formula")
		}
	}
}

// parseExpression parses expressions of the form E1 CMP E2, where CMP is one of
// the supported comparaison operators and E1 and E2 are expressions. When
// called, we known that the next token was unscanned and is a constant or place
// identifier.
func (s *formParser) parseExpression() (Formula, error) {
	var err error
	r := formAtom{}
	r.left, err = s.parseSum()
	if err != nil {
		return nil, err
	}

	tok := s.scan()
	switch tok.tok {
	case formulaEQ, formulaGEQ, formulaGTHAN, formulaLEQ, formulaLTHAN:
		r.op = tok.tok
	default:
		return nil, fmt.Errorf("Expected a comparaison operator: %s", s.problems())
	}

	r.right, err = s.parseSum()
	return r, err
}

// parseSum parses expressions of the form A1 + ... + An, where the Ai are
// either places or constants. The parameter is the first place, or constant, we
// matched. We unscan the last read token.
func (s *formParser) parseSum() (expression, error) {
	e := expression{}
	for {
	Start:
		tok := s.scan()
		switch tok.tok {
		case formulaCST:
			i, err := strconv.Atoi(tok.s)
			if err != nil {
				return e, fmt.Errorf("illegal constant %s", tok.s)
			}
			e.constant += i
			// check if there is a following + or return
			tok = s.scan()
			switch tok.tok {
			case formulaPLUS:
				goto Start
			default:
				s.unscan()
				return e, nil
			}
		case formulaID:
			i := slices.Index(s.net.Pl, tok.s)
			if i < 0 {
				return e, fmt.Errorf("unknown place identifier %s", tok.s)
			}
			e.places = append(e.places, i)
			e.names = append(e.names, tok.s)
			// check if there is a following + or return
			tok = s.scan()
			switch tok.tok {
			case formulaPLUS:
				goto Start
			default:
				s.unscan()
				return e, nil
			}
		default:
			return e, fmt.Errorf("expected a name or constant but found %s", tok.s)
		}
	}
}

// problems returns a substring of the formula where we have identified a problem
func (s *formParser) problems() string {
	return string(s.form[:s.soff]) + " ‗ " + string(s.form[s.soff:])
	// return string(s.form)
}

// read reads the next rune from the Reader. It may return eof if we reach the
// end of the formula
func (s *formParser) read() rune {
	ch, size, err := s.reader.ReadRune()
	if err != nil {
		return eof
	}
	s.eoff += size
	return ch
}

// unread places the previously read rune back on the reader. We assume that all
// runes were of size 1 and not an EOF.
func (s *formParser) unread() {
	_ = s.reader.UnreadRune()
	s.eoff--
}

// returns a token and its textual content (literal) at the current position into the input reader
func (s *formParser) position(t formTokenKind, lit string) formToken {
	return formToken{tok: t, s: lit}
}

// unscan backtrack the currently  read token.
func (s *formParser) unscan() {
	s.ahead = true
}

// scan returns the next token. We always skip whitespaces and EOL If a token
// has been unscanned then read that instead.
func (s *formParser) scan() formToken {
	if s.ahead {
		s.ahead = false
		return s.tok
	}

	// Read the next non whitespace rune.
	ch := s.read()
	for isWhitespace(ch) {
		ch = s.read()
	}

	switch {
	case isLetter(ch):
		s.unread()
		s.tok = s.scanIdent()
	case isDigit(ch):
		s.unread()
		s.tok = s.scanConstant()
	case ch == eof:
		s.tok = s.position(formulaEOF, "EOF")
	case ch == '-':
		s.tok = s.position(formulaNEG, "-")
	case ch == '(':
		s.tok = s.position(formulaLPAREN, "(")
	case ch == ')':
		s.tok = s.position(formulaRPAREN, ")")
	case ch == '+':
		s.tok = s.position(formulaPLUS, "+")
	case ch == '=':
		ch = s.read()
		if ch == '>' {
			s.tok = s.position(formulaIMPLY, "=>")
			break
		}
		s.unread()
		s.tok = s.position(formulaEQ, "=")
	case ch == '>':
		ch = s.read()
		if ch == '=' {
			s.tok = s.position(formulaGEQ, ">=")
			break
		}
		s.unread()
		s.tok = s.position(formulaGTHAN, ">")
	case ch == '<':
		ch = s.read()
		if ch == '=' {
			ch = s.read()
			if ch == '>' {
				s.tok = s.position(formulaEQUIV, "<=>")
				break
			}
			s.unread()
			s.tok = s.position(formulaLEQ, "<=")
			break
		}
		s.tok = s.position(formulaLTHAN, "<")
	case ch == '/':
		ch = s.read()
		if ch != '\\' {
			s.tok = s.position(formulaERROR, "expecting /\\")
			break
		}
		s.tok = s.position(formulaAND, "/\\")
	case ch == '\\':
		ch = s.read()
		if ch != '/' {
			s.tok = s.position(formulaERROR, "expecting \\/")
			break
		}
		s.tok = s.position(formulaOR, "\\/")
	default:
		s.tok = s.position(formulaERROR, s.buff.String())
	}

	return s.tok
}

func (s *formParser) scanIdent() formToken {
	// Create a buffer and read the current character into it.
	s.soff = s.eoff
	ch := s.read()
	s.buff.Reset()
	// Read every subsequent ident character into the buffer.
	// Non-ident characters, like EOF, will cause the loop to exit.
	// if escaped, then return the identfier until the closing '}'
	if ch == '{' {
		s.buff.WriteRune('{')
		for {
			ch = s.read()
			if ch == '}' {
				s.buff.WriteRune('}')
				break
			}
			if ch == eof {
				return s.position(formulaERROR, "closing '}' missing")
			}
			s.buff.WriteRune(ch)
		}
		return s.position(formulaID, s.buff.String())
	}

	// otherwise read the identifier and match it against reserved word
	for isLetter(ch) || isDigit(ch) || isIdentChar(ch) {
		s.buff.WriteRune(ch)
		ch = s.read()
	}
	s.unread()
	return s.position(formulaID, s.buff.String())
}

func (s *formParser) scanConstant() formToken {
	// Create a buffer and read the current character into it.
	s.soff = s.eoff
	ch := s.read()
	s.buff.Reset()

	for isDigit(ch) {
		s.buff.WriteRune(ch)
		ch = s.read()
	}
	s.unread()

	// If not a number then it is an identifier.
	return s.position(formulaCST, s.buff.String())
}
