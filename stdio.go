// Copyright 2025. Silvano DAL ZILIO. All rights reserved.
// Use of this source code is governed by the AGPL license
// that can be found in the LICENSE file.

package nets

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// Mtoa converts a marking into a string
func (net *Net) Mtoa(m Marking) string {
	var buf bytes.Buffer
	for k, v := range m {
		if k > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(net.Pl[v.Pl])
		if v.Mult != 1 {
			buf.WriteRune('*')
			buf.WriteString(strconv.Itoa(int(v.Mult)))
		}
	}
	return buf.String()
}

func (net *Net) printTransition(cond, inhibcond, inpt, delta Marking) string {
	var left, right bytes.Buffer
	for p, pname := range net.Pl {
		inp := inpt.Get(p)
		outp := delta.Get(p) - inp
		if inp == -1 {
			fmt.Fprintf(&left, " %s", pname)
		}
		if inp < -1 {
			fmt.Fprintf(&left, " %s*%d", pname, -inp)
		}
		if outp == 1 {
			fmt.Fprintf(&right, " %s", pname)
		}
		if outp > 1 {
			fmt.Fprintf(&right, " %s*%d", pname, outp)
		}
		if inhibp := inhibcond.Get(p); inhibp != 0 {
			fmt.Fprintf(&left, " %s?-%d", pname, inhibp)
		}
		if readp := cond.Get(p) + inp; readp != 0 {
			fmt.Fprintf(&left, " %s?%d", pname, readp)
		}
	}
	return fmt.Sprintf("%s ->%s\n", left.String(), right.String())
}

// FPrint formats the net structure and writes it to w.
func (net *Net) Fprint(w io.Writer) {
	fmt.Fprintf(w, "#\n# net %s\n", net.Name)
	fmt.Fprintf(w, "# %d places, %d transitions\n#\n\n", len(net.Pl), len(net.Tr))

	for k, v := range net.Pl {
		fmt.Fprintf(w, "pl %s", v)
		if net.Plabel[k] != "" {
			fmt.Fprintf(w, " : %s", net.Plabel[k])
		}
		if p := net.Initial.Get(k); p != 0 {
			fmt.Fprintf(w, " (%d)", p)
		}
		fmt.Fprint(w, "\n")
	}
	for k, v := range net.Tr {
		fmt.Fprintf(w, "tr %s ", v)
		if net.Tlabel[k] != "" {
			fmt.Fprintf(w, ": %s ", net.Tlabel[k])
		}
		if !net.Time[k].Trivial() {
			fmt.Fprint(w, net.Time[k].String())
		}
		fmt.Fprint(w, net.printTransition(net.Cond[k],
			net.Inhib[k],
			net.Pre[k],
			net.Delta[k]))
	}
	for k, v := range net.Prio {
		if len(v) != 0 {
			fmt.Fprintf(w, "pr %s >", net.Tr[k])
			for _, t := range v {
				fmt.Fprintf(w, " %s", net.Tr[t])
			}
			fmt.Fprintf(w, "\n")
		}
	}
}

// String returns a textual representation of the net structure.
func (net *Net) String() string {
	var buf bytes.Buffer
	net.Fprint(&buf)
	return buf.String()
}
