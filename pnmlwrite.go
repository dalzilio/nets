// Copyright 2025. Silvano DAL ZILIO. All rights reserved.
// Use of this source code is governed by the AGPL license
// that can be found in the LICENSE file.

package nets

import (
	"fmt"
	"io"

	"github.com/dalzilio/nets/internal/pnml"
)

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
			Init:  int(net.Initial.Get(k)),
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
			trans[k].In = append(trans[k].In, pnml.Arc{Place: &places[m.Pl], Mult: int(m.Mult)})
		}
		post := pre.Add(net.Delta[k])
		for _, m := range post {
			trans[k].Out = append(trans[k].Out, pnml.Arc{Place: &places[m.Pl], Mult: int(m.Mult)})
		}
	}
	return pnml.Write(w, net.Name, places, trans)
}
