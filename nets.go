// Copyright 2025. Silvano DAL ZILIO. All rights reserved.
// Use of this source code is governed by the AGPL license
// that can be found in the LICENSE file.

package nets

import "fmt"

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

// Marking is the type of Petri net markings. It is a slice of Atoms (places index
// and multiplicities) sorted in increasing order by places. We may use negative
// multiplicities, for instance to encode the Delta in a Petri net transition.
//
// Conventions
//
//   - Items are of the form {key, multiplicity}
//   - Items with weight 0 do not appear in multisets (default weight);
//   - Items are ordered in increasing order of keys.
type Marking []Atom

// Atom is a pair of a place index (an index in slice Pl) and a multiplicity (we
// never store places with a null multiplicity). We assume that markings and arc
// weights should fit into a 32 bits integer and we make no attempt to check if
// these values overflow.
type Atom struct {
	Pl   int
	Mult int
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
