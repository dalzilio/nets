// Copyright 2025. Silvano DAL ZILIO. All rights reserved.
// Use of this source code is governed by the AGPL license
// that can be found in the LICENSE file.

package nets

// AddToPlace returns a new Marking obtained from m by adding mult tokens to
// place pl.
func (m Marking) AddToPlace(pl int, mult int) Marking {
	if mult == 0 {
		return m
	}
	if m == nil {
		return Marking{Atom{pl, mult}}
	}
	for i := range m {
		if m[i].Pl == pl {
			m[i].Mult += mult
			if m[i].Mult == 0 {
				return append(m[:i], m[i+1:]...)
			}
			return m
		}
		if m[i].Pl > pl {
			return append(m[:i], append(Marking{Atom{pl, mult}}, m[i:]...)...)
		}
	}
	return append(m, Atom{pl, mult})
}

// Add returns the pointwise sum of two markings, m and m2.
func (m Marking) Add(m2 Marking) Marking {
	res := []Atom{}
	k1, k2 := 0, 0
	for {
		switch {
		case k1 == len(m):
			res = append(res, m2[k2:]...)
			return res
		case k2 == len(m2):
			res = append(res, m[k1:]...)
			return res
		case m[k1].Pl == m2[k2].Pl:
			if mult := m[k1].Mult + m2[k2].Mult; mult != 0 {
				res = append(res, Atom{Pl: m[k1].Pl, Mult: mult})
			}
			k1++
			k2++
		case m[k1].Pl < m2[k2].Pl:
			res = append(res, m[k1])
			k1++
		default:
			res = append(res, m2[k2])
			k2++
		}
	}
}

// IsEnabled checks if transition t in the net is enabled for marking m, meaning
// m is greater than the precondition for t (in net.Cond) and also less than the
// inhibition/capacity constraints given in net.Inhib.
func (net *Net) IsEnabled(m Marking, t int) bool {
	for _, v := range net.Cond[t] {
		if m.Get(v.Pl) < v.Mult {
			return false
		}
	}
	for _, v := range net.Inhib[t] {
		if m.Get(v.Pl) >= v.Mult {
			return false
		}
	}
	return true
}

// AllEnabled returns the set of transitions (as an ordered slice of transition index) enabled for marking m.
func (net *Net) AllEnabled(m Marking) []int {
	enabled := []int{}
	for t := range net.Tr {
		if net.IsEnabled(m, t) {
			enabled = append(enabled, t)
		}
	}
	return enabled
}

// Get returns the multiplicity associated with place pl. The returned value is
// 0 if pl is not in m.
func (m *Marking) Get(pl int) int {
	if m == nil {
		return 0
	}
	for _, a := range *m {
		if a.Pl == pl {
			return a.Mult
		}
		if a.Pl > pl {
			return 0
		}
	}
	return 0
}

// updateIfGreater returns the marking obtained from m by setting the
// multiplicity of place pl to mul, but only if mul is greater than the marking
// of pl in m. This is the least upper bound of m and the marking {pl : mul}
func (m Marking) updateIfGreater(pl int, mul int) Marking {
	if m == nil {
		return Marking{Atom{pl, mul}}
	}
	for i := range m {
		switch {
		case m[i].Pl == pl:
			if m[i].Mult < mul {
				m[i].Mult = mul
			}
			return m
		case m[i].Pl > pl:
			return append(m[:i], append(Marking{Atom{pl, mul}}, m[i:]...)...)
		}
	}
	return append(m, Atom{pl, mul})
}

// updateIfLess returns the marking obtained from m by setting the multiplicity
// of place pl to mul, but only if mul is less than the marking of pl in m. This
// is the greatest lower bound of m and the marking {pl : mul}
func (m Marking) updateIfLess(pl int, mul int) Marking {
	if m == nil {
		return Marking{Atom{pl, mul}}
	}
	for i := range m {
		switch {
		case m[i].Pl == pl:
			if m[i].Mult > mul {
				m[i].Mult = mul
			}
			return m
		case m[i].Pl > pl:
			return append(m[:i], append(Marking{Atom{pl, mul}}, m[i:]...)...)
		}
	}
	return append(m, Atom{pl, mul})
}

// Clone returns a copy of Marking  m.
func (m *Marking) Clone() Marking {
	mc := make(Marking, len(*m))
	copy(mc, *m)
	return mc
}

// Equal reports whether Marking m2 is equal to m.
func (m Marking) Equal(m2 Marking) bool {
	if len(m) != len(m2) {
		return false
	}
	for k := range m {
		if m[k] != m2[k] {
			return false
		}
	}
	return true
}
