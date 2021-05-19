// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

import (
	"fmt"
	"io"

	"github.com/dalzilio/nets/internal/pnml"
)

// madd returns a new Marking obtained by adding the multiplicities of the
// element in m1 and m2.
func madd(m1, m2 Marking) Marking {
	i1, i2 := 0, 0
	m := Marking{}
	for {
		switch {
		case i1 == len(m1):
			return append(m, m2[i2:]...)
		case i2 == len(m2):
			return append(m, m1[i1:]...)
		case m1[i1].Pl == m2[i2].Pl:
			if sum := m1[i1].Mult + m2[i2].Mult; sum != 0 {
				m = append(m, Atom{m1[i1].Pl, sum})
			}
			i1++
			i2++
		case m1[i1].Pl < m2[i2].Pl:
			m = append(m, m1[i1])
			i1++
		case m2[i2].Pl < m1[i1].Pl:
			m = append(m, m2[i2])
			i2++
		}
	}
}

// Pnml marshall a Net into a P/T net in PNML format and writes the output on an
// io.Writer. Because of limitations in the PNML format, we return an error if
// the net has inhibitor arcs. We also drop timing information on transitions
// and replace read arcs with "tests"; meaning a pair of input/output arcs.
//
// This method is only useful if you create or modify an object of type Net. It
// is preferable to use the `ndrio` program to transform a .net file into a PNML
// P/T file.
//
// We combine names and labels for the naming of places and transitions in the
// PNML file but we build the id by adding a prefix ('pl_' for places and 'tr_'
// for transitions), because it is possible to use the same name as a place and
// as a transition in a .net file.
func (net *Net) Pnml(w io.Writer) error {
	for k, v := range net.Inhib {
		if len(v) != 0 {
			return fmt.Errorf("cannot marshal net with inhibitor arcs; see transition %s", net.Tr[k])
		}
	}
	places := make([]pnml.Place, len(net.Pl))
	trans := make([]pnml.Trans, len(net.Tr))
	for k, v := range net.Pl {
		places[k] = pnml.Place{
			Name:  v,
			Label: net.Plabel[k],
			Init:  net.Initial.Get(k),
		}
	}
	for k, v := range net.Tr {
		trans[k] = pnml.Trans{
			Name:  v,
			Label: net.Tlabel[k],
			In:    []pnml.Arc{},
			Out:   []pnml.Arc{},
		}
		pre := net.Cond[k]
		for _, m := range pre {
			trans[k].In = append(trans[k].In, pnml.Arc{Place: &places[m.Pl], Mult: m.Mult})
		}
		post := madd(pre, net.Delta[k])
		for _, m := range post {
			trans[k].Out = append(trans[k].Out, pnml.Arc{Place: &places[m.Pl], Mult: m.Mult})
		}
	}
	return pnml.Write(w, net.Name, places, trans)
}
