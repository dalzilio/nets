// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

import (
	"bytes"
	"fmt"
	"strconv"
)

// Net is the concrete type of Time Petri Nets. We support labels on both
// transitions and places. The semantics of nets is as follows. Our choice
// differs from the traditional semantics of TPN (based on Pre- and
// Post-conditions), because we want to uniformly support inhibitor-arcs and
// capacities.
//
// • COND: In a condition, Cond[k], an atom (p, m) entails that if transition
// Tr[k] is enabled at marking M then M.Get(p) >= m. Therefore Tr[k] is enabled
// at M when Compare(M, Cond[k]) >= 0 (for the pointwise comparison).
//
// • INHIB: In a dual way, in the inhibition condition Inhib[k], if transition
// Tr[k] is enabled at marking M then M.Get(p) < Inhib[k].Get(p) or
// Inhib[k].Get(p) == 0.
//
// • PRE: The value of Pre[k] models the arcs from an "input" place to
// transition Tr[k]. In a TPN, the value of Tr[k].Get(p) gives the tokens
// "losts" from the input plac p. This is useful when we need to check if a
// (timed) transition is re-initialized. Transition Tr[k2] is not re-initialized
// at marking M, after firing Tr[k1], if Compare(Add(M, Pre[k2]), Cond[k1]) >= 0
// (using pointwise operations).
//
// • DELTA: An atom (p, m) in Delta[k] indicates that if Tr[k] fires then the
// marking of place p must increase by m (in this case m can be negative). Hence
// if we fire Tr[k] at marking M, the result is Add(M, Delta[k]).
type Net struct {
	Name    string         // Name of the net.
	Pl      []string       // List of places names.
	Tr      []string       // List of transitions names.
	Tlabel  []string       // List of transition labels. We use the empty string when no labels.
	Plabel  []string       // List of place labels.
	Time    []TimeInterval // List of (static) timing constraints for each transition.
	Cond    []Marking      // Each transition has a list of conditions.
	Inhib   []Marking      // Each transition has inhibition conditions (possibly with capacities).
	Pre     []Marking      // The Pre (input places) condition for each transition (only useful with read arcs in TPN).
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

// trivial is true if the time interval i is of the form [0, w[ or if the
// interval is un-initialized (meaning the left part of the interval is of kind
// BINFTY)
func (i *TimeInterval) trivial() bool {
	if i.Left.Bkind == BINFTY {
		return true
	}
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

// intersectWith sets interval i to the intersection of  i and j. We return an
// error if the intersection is empty.
func (i *TimeInterval) intersectWith(j TimeInterval) error {
	if i.Left.Bkind == BINFTY {
		// it means we are initializing the interval
		i.Left.Bkind = j.Left.Bkind
		i.Left.Value = j.Left.Value
		i.Right.Bkind = j.Right.Bkind
		i.Right.Value = j.Right.Value
	}
	if j.Left.Bkind == BINFTY {
		return fmt.Errorf("bad time interval when computing intersection")
	}
	// we compute the max of the left parts
	if j.Left.Value >= i.Left.Value {
		if j.Left.Value > i.Left.Value || (j.Left.Value == i.Left.Value && j.Left.Bkind == BOPEN) {
			// we update the left part
			i.Left.Bkind = j.Left.Bkind
			i.Left.Value = j.Left.Value
		}
	}
	if j.Right.Bkind == BINFTY {
		// we do not need to update the right part
		return nil
	}
	if i.Right.Bkind == BINFTY {
		i.Right.Bkind = j.Right.Bkind
		i.Right.Value = j.Right.Value
		return nil
	}
	// when both intervals are right-bounded we take the min of their right parts
	if j.Right.Value <= i.Right.Value {
		if j.Right.Value < i.Right.Value || (j.Right.Value == i.Right.Value && j.Right.Bkind == BOPEN) {
			i.Right.Bkind = j.Right.Bkind
			i.Right.Value = j.Right.Value
		}
	}
	// we need to test if the result is empty
	if i.Right.Value < i.Left.Value || (i.Right.Value == i.Left.Value && (i.Left.Bkind == BOPEN || i.Right.Bkind == BOPEN)) {
		return fmt.Errorf("empty time interval when computing intersection")
	}
	return nil
}

func (i *TimeInterval) String() string {
	var buf bytes.Buffer
	if i.Left.Bkind == BINFTY {
		// it means interval was never set
		return "[0,w["
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

// Add returns the pointwise sum of m1 and m2.
func Add(m1, m2 Marking) Marking {
	res := []Atom{}
	k1, k2 := 0, 0
	for {
		switch {
		case k1 == len(m1):
			res = append(res, m2[k2:]...)
			return res
		case k2 == len(m2):
			res = append(res, m1[k1:]...)
			return res
		case m1[k1].Pl == m2[k2].Pl:
			if mult := m1[k1].Mult + m2[k2].Mult; mult != 0 {
				res = append(res, Atom{Pl: m1[k1].Pl, Mult: mult})
			}
			k1++
			k2++
		case m1[k1].Pl < m2[k2].Pl:
			res = append(res, m1[k1])
			k1++
		default:
			res = append(res, m2[k2])
			k2++
		}
	}
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

// PrioClosure updates the priority relation by computing its transitive
// closure. We return an error if we have circular dependencies between
// transitions.
func (net *Net) PrioClosure() error {
	// We keep a list/set of the transitions for which we have computed the
	// closure and a work list of transitions we need to work with. Initially we
	// start with all the transitions ti that do not appear in a relation ti >
	// tj. Then we iterate going backward from this list, adding the transitions
	// that have all their "successors" in the done list.
	done := []int{}
	work := []int{}
	for k, v := range net.Prio {
		if len(v) == 0 {
			done = setAdd(done, k)
		} else {
			work = setAdd(work, k)
		}
	}
	if len(done) == len(net.Tr) {
		// the priority list of all transitions is empty; so we have no
		// priorities at all
		return nil
	}
	if len(done) == 0 {
		return fmt.Errorf("problem with priorities, no minimal elements")
	}
	for {
		if len(work) == 0 {
			return nil
		}
		workn := []int{}
		donen := make([]int, len(done))
		copy(donen, done)
		for _, t := range work {
			if setIncluded(net.Prio[t], done) {
				for _, v := range net.Prio[t] {
					net.Prio[t] = setUnion(net.Prio[t], net.Prio[v])
				}
				donen = setAdd(donen, t)
			} else {
				workn = setAdd(workn, t)
			}
		}
		// The length of work should decrease  at each loop, otherwise it means
		// we have a circular dependency
		if len(workn) == len(work) {
			for _, t := range work {
				if setMember(net.Prio[t], t) >= 0 {
					return fmt.Errorf("cyclic dependencies in priority for %s", net.Tr[t])
				}
			}
			return fmt.Errorf("cyclic dependencies between priorities")
		}
		work = workn
		done = donen
	}
}
