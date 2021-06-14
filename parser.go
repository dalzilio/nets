// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

//
// code inspired by: http://blog.gopheracademy.com/advent-2014/parsers-lexers/
//

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// parser represents a net parser.
type parser struct {
	s      *scanner
	net    *Net           // top-level net (head of the stack)
	pl, tr map[string]int // list of place and trans. identifiers
	tok    token          // last read token
	ahead  bool           // true if there is a token stored in tok
}

// Parse returns a pointer to a Net structure from a textual representation of a
// TPN. We return a nil pointer and an error if there was a problem while
// reading the specification.
func Parse(r io.Reader) (*Net, error) {
	p := &parser{
		s:     &scanner{r: bufio.NewReader(r), pos: &textPos{}},
		net:   &Net{},
		pl:    make(map[string]int),
		tr:    make(map[string]int),
		ahead: false,
	}
	if err := p.parse(); err != nil {
		return nil, fmt.Errorf("error parsing net: %s", err)
	}
	return p.net, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *parser) scan() token {
	// If we have a token on the buffer, then return it.
	// Otherwise read the next token from the scanner.
	// and save it to the buffer in case we unscan later.
	if p.ahead {
		p.ahead = false
	} else {
		p.tok = p.s.scan()
	}
	return p.tok
}

// unscan backtrack the currently  read token.
func (p *parser) unscan() {
	p.ahead = true
}

// checkPL returns the index of a place in the net and creates one if necessary.
// We do not support placer labels at the moment.
func (p *parser) checkPL(s string) int {
	n, ok := p.pl[s]
	if !ok {
		n = len(p.pl)
		p.pl[s] = n
		p.net.Pl = append(p.net.Pl, s)
		p.net.Plabel = append(p.net.Plabel, "")
	}
	return n
}

// checkTR returns the index of a transition in the net and creates one if necessary
func (p *parser) checkTR(s string) int {
	n, ok := p.tr[s]
	if !ok {
		n = len(p.tr)
		p.tr[s] = n
		p.net.Tr = append(p.net.Tr, s)
		p.net.Tlabel = append(p.net.Tlabel, "")
		p.net.Time = append(p.net.Time, TimeInterval{})
		p.net.Cond = append(p.net.Cond, nil)
		p.net.Inhib = append(p.net.Inhib, nil)
		p.net.Pre = append(p.net.Pre, nil)
		p.net.Delta = append(p.net.Delta, nil)
		p.net.Prio = append(p.net.Prio, nil)
	}
	return n
}

func (p *parser) parse() error {
	for {
		switch tok := p.scan(); tok.tok {
		case tokEOF:
			return nil
		case tokNET:
			tok = p.scan()
			if tok.tok != tokIDENT {
				return fmt.Errorf(" found %q; expected identifier after NET at %s", tok.s, tok.pos.String())
			}
			p.net.Name = tok.s
		case tokTR:
			if e := p.parseTR(); e != nil {
				return e
			}
		case tokPL:
			if e := p.parsePL(); e != nil {
				return e
			}
		case tokPRIO:
			if e := p.parsePRIO(); e != nil {
				return e
			}
		case tokNOTE:
			if e := p.parseNOTE(); e != nil {
				return e
			}
		default:
			return fmt.Errorf(" found %q; expected keywords, %s",
				tok.s, tok.pos.String())
		}
	}
}

func (p *parser) parseTR() error {
	var err error
	tok := p.scan()
	if tok.tok != tokIDENT {
		return fmt.Errorf(" found %q, expected valid transition name at %s", tok.s, tok.pos.String())
	}
	index := p.checkTR(tok.s)
	// we shouldcheck for an (optional) label then (also optional) time
	// interval, in this order.
	//    ’tr’ <transition> {":" <label>} {<interval>} {<tinput> -> <toutput>}
	afterArrow := false
	haslabel := false
	hastinterval := false
	hasarcs := false
	for {
		switch tok := p.scan(); tok.tok {
		case tokLABEL:
			if haslabel || hastinterval || hasarcs {
				return fmt.Errorf(" bad label declaration, at %s", tok.pos.String())
			}
			haslabel = true // to avoid double label decl
			p.net.Tlabel[index] = tok.s
		case tokTIMINGC:
			if hastinterval || hasarcs {
				return fmt.Errorf(" bad time interval declaration, at %s", tok.pos.String())
			}
			hastinterval = true // to avoid double time interval decl
			tgc := TimeInterval{}
			arr := strings.Fields(tok.s)
			if len(arr) != 4 {
				return fmt.Errorf(" bad time interval declaration, %s at %s", tok.s, tok.pos.String())
			}
			if arr[0] == "[" {
				tgc.Left.Bkind = BCLOSE
			} else {
				tgc.Left.Bkind = BOPEN
			}
			tgc.Left.Value, err = strconv.Atoi(arr[1])
			if err != nil {
				return fmt.Errorf(" in timing interval, %s at %s", tok.s, tok.pos.String())
			}
			if arr[2] == "w" {
				tgc.Right.Bkind = BINFTY
			} else {
				tgc.Right.Value, err = strconv.Atoi(arr[2])
				if (err != nil) || (tgc.Right.Value < tgc.Left.Value) {
					return fmt.Errorf(" in timing interval, %s at %s", tok.s, tok.pos.String())
				}
				if arr[3] == "[" {
					tgc.Right.Bkind = BOPEN
				} else {
					tgc.Right.Bkind = BCLOSE
				}
			}
			if err := p.net.Time[index].intersectWith(tgc); err != nil {
				return fmt.Errorf(" %s: for transition %s, at %s", err, p.net.Tr[index], tok.pos.String())
			}
		case tokARROW:
			if afterArrow {
				return fmt.Errorf(" cannot have two arrows (->) in tr declaration at %s", tok.pos.String())
			}
			hasarcs = true // to avoid label and time interval decl after declaring arcs
			afterArrow = true
		case tokIDENT:
			// tinput  ::= <place>{<arc>}
			// toutput ::= <place>{<normal_arc>}
			pindex := p.checkPL(tok.s)
			hasarcs = true
			tok = p.scan()
			mult := 1
			ok := false
			switch tok.tok {
			case tokREAD:
				if afterArrow {
					return fmt.Errorf(" read arcs in outputs of transition at %s", tok.pos.String())
				}
				mult, err = mconvert(tok.s)
				if err != nil {
					return fmt.Errorf(" in multiplicity, %s (%s) at %s", tok.s, err, tok.pos.String())
				}
				p.net.Cond[index] = p.net.Cond[index].setifbigger(pindex, mult)
			case tokINHIBITOR:
				if afterArrow {
					return fmt.Errorf(" inhibitor arcs in outputs of transition at %s", tok.pos.String())
				}
				mult, err = mconvert(tok.s)
				if err != nil {
					return fmt.Errorf(" in multiplicity, %s (%s) at %s", tok.s, err, tok.pos.String())
				}
				p.net.Inhib[index] = p.net.Inhib[index].setiflower(pindex, mult)
			case tokSTAR:
				mult, err = mconvert(tok.s)
				if err != nil {
					return fmt.Errorf(" in multiplicity, %s (%s) at %s", tok.s, err, tok.pos.String())
				}
				ok = true
				fallthrough
			default:
				if !ok {
					// it means that we did not fallthrough the previous case
					// and we need to pop back the extra token that we scanned
					// looking for an arc description
					p.unscan()
				}
				if afterArrow {
					p.net.Delta[index] = p.net.Delta[index].add(pindex, mult)
				} else {
					p.net.Delta[index] = p.net.Delta[index].add(pindex, -mult)
					p.net.Pre[index] = p.net.Pre[index].add(pindex, -mult)
					p.net.Cond[index] = p.net.Cond[index].add(pindex, mult)
				}
			}
		default:
			p.unscan()
			return nil
		}
	}
}

func (p *parser) parsePL() error {
	//   pldesc ::= ’pl’ <place> {":" <label>} {(<marking>)} {<pinput> -> <poutput>}
	var err error
	tok := p.scan()
	if tok.tok != tokIDENT {
		return fmt.Errorf(" found %q, expected valid place name at %s", tok.s, tok.pos.String())
	}
	index := p.checkPL(tok.s)
	afterArrow := false // in case we have tr declarations
	haslabel := false
	hasinitm := false
	hasarcs := false
	for {
		switch tok := p.scan(); tok.tok {
		case tokLABEL:
			if haslabel || hasinitm || hasarcs {
				return fmt.Errorf(" bad label declaration, at %s", tok.pos.String())
			}
			haslabel = true
			p.net.Plabel[index] = tok.s
		case tokMARKING:
			if hasinitm || hasarcs {
				return fmt.Errorf(" bad marking declaration, at %s", tok.pos.String())
			}
			plm, err := mconvert(tok.s)
			if err != nil {
				return fmt.Errorf(" in marking, %s (%s) at %s", tok.s, err, tok.pos.String())
			}
			hasinitm = true
			p.net.Initial = p.net.Initial.add(index, plm)
		case tokARROW:
			if afterArrow {
				return fmt.Errorf(" cannot have two arrows (->) in pl declaration at %s", tok.pos.String())
			}
			hasarcs = true // to avoid label and time interval decl after declaring arcs
			afterArrow = true
		case tokIDENT:
			// then tok.s is the name of a transition
			//    pinput  ::= <transition>{<normal_arc>}
			//    poutput ::= <transition>{arc}
			tindex := p.checkTR(tok.s)
			hasarcs = true
			tok = p.scan()
			mult := 1
			ok := false
			switch tok.tok {
			case tokREAD:
				if !afterArrow {
					return fmt.Errorf(" read arcs in inputs of place, at %s", tok.pos.String())
				}
				mult, err = mconvert(tok.s)
				if err != nil {
					return fmt.Errorf(" in multiplicity, %s (%s) at %s", tok.s, err, tok.pos.String())
				}
				p.net.Cond[tindex] = p.net.Cond[tindex].setifbigger(index, mult)
			case tokINHIBITOR:
				if !afterArrow {
					return fmt.Errorf(" inhibitor arcs in inputs of place at %s", tok.pos.String())
				}
				mult, err = mconvert(tok.s)
				if err != nil {
					return fmt.Errorf(" in multiplicity, %s (%s) at %s", tok.s, err, tok.pos.String())
				}
				p.net.Inhib[tindex] = p.net.Inhib[tindex].setiflower(index, mult)
			case tokSTAR:
				mult, err = mconvert(tok.s)
				if err != nil {
					return fmt.Errorf(" in multiplicity, %s (%s) at %s", tok.s, err, tok.pos.String())
				}
				ok = true
				fallthrough
			default:
				if !ok {
					// it means that we did not fallthrough the previous case
					// (we have a normal arc, witout a '?' or '*' decoration)
					// and we need to pop back the extra token that we scanned
					p.unscan()
				}
				if afterArrow {
					p.net.Delta[tindex] = p.net.Delta[tindex].add(index, -mult)
					p.net.Pre[tindex] = p.net.Pre[tindex].add(index, -mult)
					p.net.Cond[tindex] = p.net.Cond[tindex].add(index, mult)
				} else {
					p.net.Delta[tindex] = p.net.Delta[tindex].add(index, mult)
				}
			}
		default:
			p.unscan()
			return nil
		}
	}
}

func (p *parser) parseNOTE() error {
	tok := p.scan()
	if tok.tok != tokIDENT {
		return fmt.Errorf(" found %q, expected a note identifier at %s", tok.s, tok.pos.String())
	}
	tok = p.scan()
	if tok.tok != tokINT {
		return fmt.Errorf(" found %q, expected a note index at %s", tok.s, tok.pos.String())
	}
	tok = p.scan()
	if tok.tok != tokIDENT {
		return fmt.Errorf(" found %q, expected a note body at %s", tok.s, tok.pos.String())
	}
	return nil
}

func (p *parser) parsePRIO() error {
	pre, post := []int{}, []int{}
	isgt := false
	var tok token
	for {
		tok = p.scan()
		if tok.tok != tokIDENT {
			break
		}
		n := p.checkTR(tok.s)
		pre = setAdd(pre, n)
	}
	if tok.tok != tokGT && tok.tok != tokLT {
		return fmt.Errorf("found %q, expected priority > or < at %s", tok.s, tok.pos.String())
	}
	if tok.tok == tokGT {
		isgt = true
	}
	for {
		tok = p.scan()
		if tok.tok != tokIDENT {
			// if we found GT, we add pre > post
			if isgt {
				for _, t := range pre {
					p.net.Prio[t] = setUnion(p.net.Prio[t], post)
				}
			} else {
				for _, t := range post {
					p.net.Prio[t] = setUnion(p.net.Prio[t], pre)
				}

			}
			p.unscan()
			return nil
		}
		n := p.checkTR(tok.s)
		post = setAdd(post, n)
	}
}

// setAdd takes a sorted list of integers (here transitions index), s, and adds
// v to it.
func setAdd(s []int, v int) []int {
	if len(s) == 0 {
		return []int{v}
	}
	for i := range s {
		if s[i] == v {
			return s
		}
		if s[i] > v {
			res := make([]int, len(s)+1)
			copy(res[:i], s[:i])
			copy(res[i+1:], s[i:])
			res[i] = v
			return res
		}
	}
	res := make([]int, len(s))
	copy(res, s)
	res = append(res, v)
	return res
}

// setUnion does set union between two slices of sorted integers, s1 and s2.
func setUnion(s1, s2 []int) []int {
	res := make([]int, len(s1))
	copy(res, s1)
	for _, v := range s2 {
		res = setAdd(res, v)
	}
	return res
}

// mconvert is used to convert values found on markings and weights into
// integers. We take into account the possibility that s ends with a
// "multiplier", such as `3K` (3000), which is valid in Tina.
func mconvert(s string) (int, error) {
	if len(s) == 0 {
		return 0, errors.New("empty value in weights or marking")
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		if ch := s[len(s)-1]; ch == 'K' || ch == 'M' || ch == 'G' || ch == 'T' || ch == 'P' || ch == 'E' {
			v, err = strconv.Atoi(s[:len(s)-1])
			if err != nil {
				return 0, fmt.Errorf("not a valid weight or marking; %s", err)
			}
			switch ch {
			case 'K':
				return v * 1000, nil
			case 'M':
				return v * 1000000, nil
			case 'G':
				return v * 1000000000, nil
			case 'T':
				return v * 1000000000000, nil
			case 'P':
				return v * 1000000000000000, nil
			case 'E':
				return v * 1000000000000000000, nil
			default:
				return v, fmt.Errorf("not a valid multiplier in weight or marking; %v", ch)
			}
		}
	}
	return v, nil
}
