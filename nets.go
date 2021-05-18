// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

import (
	"bytes"
	"strconv"
)

// Net is the concrete type of Time Petri Nets. We support labels on both
// transitions and places.
//
// Conventions
//
//      - In a condition, an element (v, m) means p_v >= m
//      - In a delta, an element (v, m) means p_v += m
type Net struct {
	Name    string         // Name of the net.
	Pl      []string       // List of places names.
	Tr      []string       // List of transitions names.
	Tlabel  []string       // List of transition labels. We use the empty string when no labels.
	Plabel  []string       // List of place labels.
	Time    []TimeInterval // List of (static) timing constraints for each transition.
	Cond    []Marking      // Each transition has a list of conditions (for read arcs).
	Inhib   []Marking      // Each transition has inhibition conditions (possibly with capacities).
	Pre     []Marking      // The Pre (input places) condition for each transition.
	Delta   []Marking      // The delta (Post - Pre) for each transition.
	Initial Marking        // Initial marking of places.
	Prio    [][]int        // the slice Prio[i] lists all transitions with less priority than Tr[i] (the slice is sorted).
}

// Marking is the type of Petri net markings. It is a set of Atoms (places index
// and multiplicities) sorted in increasing order by places. We use negative
// multiplicities to encode the Delta in a Petri net transition.
//
// Conventions
//
//      - Items are of the from {key, multiplicity}
//      - Items with weight 0 do not appear in multisets (default weight);
//      - Items are ordered in increasing order of keys.
//
type Marking []Atom

// Atom is a pair of a place index (an index in slice Pl) and a multiplicity (we
// never store places with a null multiplicity). The value of arc weights and
// the initial marking of places may not fit into a 32 bits integer (but should
// fit in a 64 bits one). We make no attempt to check if these values overflow.
type Atom struct{ Pl, Mult int }

// Bkind is the type of possible time constraints bounds
type Bkind uint8

// Bkind is an enumeration describing the three different types of (time)
// interval bounds. BINFTY, as a right bound, is used for infinite intervals. As
// a left bound, it is used to denote empty intervals (errors).
const (
	BINFTY Bkind = iota // ..,w[
	BCLOSE              // [a,..
	BOPEN               // ]a,..
)

// Bound is the type of bounds in a time interval.
type Bound struct {
	Bkind
	Value int
}

// TimeInterval is the type of time intervals.
type TimeInterval struct {
	Left, Right Bound
}

// trivial is true if the time interval i is of the form [0, w[.
func (i *TimeInterval) trivial() bool {
	if i.Right.Bkind != BINFTY {
		return false
	}
	if i.Left.Bkind != BCLOSE {
		return false
	}
	if i.Left.Value != 0 {
		return false
	}
	return true
}

func (i *TimeInterval) String() string {
	var buf bytes.Buffer
	if i.Left.Bkind == BINFTY {
		return "[void]"
	}
	if i.Left.Bkind == BCLOSE {
		buf.WriteRune('[')
	} else {
		buf.WriteRune(']')
	}
	buf.WriteString(strconv.Itoa(int(i.Left.Value)))
	buf.WriteRune(',')
	if i.Right.Bkind == BINFTY {
		buf.WriteString("w[")
	} else {
		buf.WriteString(strconv.Itoa(int(i.Right.Value)))
		if i.Right.Bkind == BCLOSE {
			buf.WriteRune(']')

		} else {
			buf.WriteRune('[')
		}
	}
	return buf.String()
}

// add updates a marking by adding the value v with multiplicity k
// to m.
func (m Marking) add(val int, mul int) Marking {
	if mul == 0 {
		return m
	}
	if m == nil {
		return Marking{Atom{val, mul}}
	}
	for i := range m {
		if m[i].Pl == val {
			m[i].Mult += mul
			if m[i].Mult == 0 {
				return append(m[:i], m[i+1:]...)
			}
			return m
		}
		if m[i].Pl > val {
			return append(m[:i], append(Marking{Atom{val, mul}}, m[i:]...)...)
		}
	}
	return append(m, Atom{val, mul})
}

// Get returns the multiplicity associated with value v. The returned
// value is 0 if v is not in m.
func (m Marking) Get(v int) int {
	if m == nil {
		return 0
	}
	for _, a := range m {
		if a.Pl == v {
			return a.Mult
		}
		if a.Pl > v {
			return 0
		}
	}
	return 0
}

// setifbigger updates a marking by setting the multiplicity of val to mul in m,
// but only if mul is bigger than the existing value.
func (m Marking) setifbigger(val int, mul int) Marking {
	if m == nil {
		return Marking{Atom{val, mul}}
	}
	for i := range m {
		switch {
		case m[i].Pl == val:
			if m[i].Mult < mul {
				m[i].Mult = mul
			}
			return m
		case m[i].Pl > val:
			return append(m[:i], append(Marking{Atom{val, mul}}, m[i:]...)...)
		}
	}
	return append(m, Atom{val, mul})
}

// setiflower updates a marking by setting the multiplicity of val to mul in m,
// but only if mul if lower than the existing value.
func (m Marking) setiflower(val int, mul int) Marking {
	if m == nil {
		return Marking{Atom{val, mul}}
	}
	for i := range m {
		switch {
		case m[i].Pl == val:
			if m[i].Mult > mul {
				m[i].Mult = mul
			}
			return m
		case m[i].Pl > val:
			return append(m[:i], append(Marking{Atom{val, mul}}, m[i:]...)...)
		}
	}
	return append(m, Atom{val, mul})
}
