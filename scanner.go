// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

import (
	"bufio"
	"bytes"
	"strings"
)

// scanner adds a position field for easy error reporting
type scanner struct {
	r   *bufio.Reader
	pos *textPos
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	if s.pos.ahead != 0 {
		s.pos.ahead--
	} else {
		if ch == '\n' {
			s.pos.line++
			s.pos.col = 0
		} else {
			s.pos.col++
		}
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *scanner) unread() {
	_ = s.r.UnreadRune()
	s.pos.ahead++
}

// returns a token with the current position in the file
func (s *scanner) position(t tokenKind, lit string) token {
	return token{tok: t, pos: *s.pos, s: lit}
}

// scan returns the next token and literal value.
// We always skip whitespaces and EOL
func (s *scanner) scan() token {
	// Read the next non whitespace rune.
	ch := s.read()
	for isWhitespace(ch) {
		ch = s.read()
	}

	switch {
	case isLetter(ch):
		s.unread()
		return s.scanIdent()
	case isDigit(ch):
		value := s.scanNumber(ch)
		return s.position(tokINT, value)
	case ch == eof:
		return s.position(tokEOF, "EOF")
	case ch == ':':
		return s.scanLabel()
	case ch == '?' || ch == '*':
		return s.scanArc(ch)
	case ch == '-':
		if ch1 := s.read(); ch1 == '>' {
			return s.position(tokARROW, "->")
		}
		return s.position(tokILLEGAL, string(ch))
	case ch == '(':
		return s.scanMarking()
	case (ch == '[') || (ch == ']'):
		s.unread()
		return s.scanTimingConstraint()
	case ch == '>':
		return s.position(tokGT, string(ch))
	case ch == '<':
		return s.position(tokLT, string(ch))
	case ch == '#':
		// this is a comment, we skip until '\n'
		for {
			ch = s.read()
			if ch == eof || ch == '\n' || ch == '\r' {
				s.unread()
				return s.scan()
			}
		}
	default:
		return s.position(tokILLEGAL, string(ch))
	}
}

func (s *scanner) scanTimingConstraint() token {
	// Skip every character until a closing bracket
	// and returns a white-space separated list of Bounds
	ch := s.read()
	var buf bytes.Buffer
	buf.WriteRune(ch)
	buf.WriteRune(' ')
	for {
		ch = s.read()
		switch {
		case (ch == '[') || (ch == ']'):
			buf.WriteRune(' ')
			buf.WriteRune(ch)
			return s.position(tokTIMINGC, buf.String())
		case ch == ',':
			buf.WriteRune(' ')
		case isDigit(ch) || (ch == 'w'):
			buf.WriteRune(ch)
		case isWhitespace(ch):
		default:
			return s.position(tokILLEGAL, string(ch))
		}
	}
}

func (s *scanner) scanArc(r rune) token {
	ch := s.read()
	switch {
	case (r == '?'):
		switch {
		case isDigit(ch):
			weight := s.scanNumber(ch)
			return s.position(tokREAD, weight)
		case ch == '-':
			weight := s.scanNumber(0)
			return s.position(tokINHIBITOR, weight)
		default:
			return s.position(tokILLEGAL, string(ch))
		}
	case (r == '*'):
		switch {
		case isDigit(ch):
			weight := s.scanNumber(ch)
			return s.position(tokSTAR, weight)
		default:
			return s.position(tokILLEGAL, string(ch))
		}
	default:
		return s.position(tokILLEGAL, string(ch))
	}
}

func (s *scanner) scanLabel() token {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer

	ch := s.read()
	for isWhitespace(ch) {
		ch = s.read()
	}

	if ch == eof || ch == '}' || ch == '\\' {
		return s.position(tokILLEGAL, string(ch))
	}

	if ch == '{' {
		// we accept any chain between braces, and in which characters {, }, and \ are prefixed by \
		buf.WriteRune('{')
		for ch != '}' {
			ch = s.read()
			if ch == eof || ch == '\n' || ch == '\r' {
				return s.position(tokILLEGAL, buf.String())
			}
			if ch == '\\' {
				buf.WriteRune(ch)
				// we possibly have an escaped character
				ch = s.read()
				buf.WriteRune(ch)
				if ch != '{' && ch != '}' && ch != '\\' {
					return s.position(tokILLEGAL, buf.String())
				}
				ch = s.read()
			}
			buf.WriteRune(ch)
		}
		return s.position(tokLABEL, buf.String())
	}

	// Read every subsequent ident character into the buffer.
	// Until the first Non-ident characters
	// We do not accept "escaped" label names at the moment
	for {
		switch {
		case isWhitespace(ch):
			s.unread()
			return s.position(tokLABEL, buf.String())
		case ch == eof:
			return s.position(tokILLEGAL, "EOF")
		default:
			buf.WriteRune(ch)
		}
		ch = s.read()
	}
}

func (s *scanner) scanMarking() token {
	value := s.scanNumber(0)
	ch := s.read()
	switch {
	case ch == ')':
		return s.position(tokMARKING, value)
	default:
		return s.position(tokILLEGAL, string(ch))
	}
}

func (s *scanner) scanIdent() token {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	ch := s.read()

	// Read every subsequent ident character into the buffer. Non-ident
	// characters, like EOF, will cause the loop to exit. If escaped we return
	// the identfier until the closing '}'

	if ch == '}' {
		buf.WriteRune(ch)
		return s.position(tokILLEGAL, buf.String())
	}

	if ch == '{' {
		// we accept any chain between braces, and in which characters {, }, and \ are prefixed by \
		buf.WriteRune('{')
		for ch != '}' {
			ch = s.read()
			if ch == eof || ch == '\n' || ch == '\r' {
				return s.position(tokILLEGAL, buf.String())
			}
			if ch == '\\' {
				buf.WriteRune(ch)
				// we possibly have an escaped character
				ch = s.read()
				buf.WriteRune(ch)
				if ch != '{' && ch != '}' && ch != '\\' {
					return s.position(tokILLEGAL, buf.String())
				}
				ch = s.read()
			}
			buf.WriteRune(ch)
		}
		return s.position(tokIDENT, buf.String())
	}

	// otherwise read the identifier and match it against reserved word
	for isLetter(ch) || isDigit(ch) || isIdentChar(ch) {
		buf.WriteRune(ch)
		ch = s.read()
	}
	s.unread()
	switch strings.ToUpper(buf.String()) {
	case "TR":
		return s.position(tokTR, "tr")
	case "NET":
		return s.position(tokNET, "net")
	case "PL":
		return s.position(tokPL, "pl")
	case "PR":
		return s.position(tokPRIO, "pr")
	case "NT":
		return s.position(tokNOTE, "nt")
	}

	// If not reserved then return as a regular identifier.
	return s.position(tokIDENT, buf.String())
}

// scanNumber scan the input for digits and return the resulting number as a
// string. The value of c is either 0 or the first digit of the result
func (s *scanner) scanNumber(c rune) string {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	if c != 0 {
		buf.WriteRune(c)
	}
	ch := s.read()
	for isDigit(ch) {
		buf.WriteRune(ch)
		ch = s.read()
	}
	if ch == 'K' || ch == 'M' || ch == 'G' || ch == 'T' || ch == 'P' || ch == 'E' {
		buf.WriteRune(ch)
		return buf.String()
	}
	s.unread()
	return buf.String()
}
