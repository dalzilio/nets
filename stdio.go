// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

import (
	"bytes"
	"fmt"
)

func (net *Net) PrintTransition(cond, inhibcond, inpt, delta Marking) string {
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

func (net *Net) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "#\n# net %s\n", net.Name)
	fmt.Fprintf(&buf, "# %d places, %d transitions\n#\n\n", len(net.Pl), len(net.Tr))

	for k, v := range net.Pl {
		fmt.Fprintf(&buf, "pl %s", v)
		if net.Plabel[k] != "" {
			fmt.Fprintf(&buf, " : %s", net.Plabel[k])
		}
		if p := net.Initial.Get(k); p != 0 {
			fmt.Fprintf(&buf, " (%d)", p)
		}
		fmt.Fprintf(&buf, "\n")
	}
	for k, v := range net.Tr {
		fmt.Fprintf(&buf, "tr %s ", v)
		if net.Tlabel[k] != "" {
			fmt.Fprintf(&buf, ": %s ", net.Tlabel[k])
		}
		if !net.Time[k].trivial() {
			buf.WriteString(net.Time[k].String())
		}
		buf.WriteString(net.PrintTransition(net.Cond[k],
			net.Inhib[k],
			net.Pre[k],
			net.Delta[k]))
	}
	for k, v := range net.Prio {
		if len(v) != 0 {
			fmt.Fprintf(&buf, "pr %s >", net.Tr[k])
			for _, t := range v {
				fmt.Fprintf(&buf, " %s", net.Tr[t])
			}
			fmt.Fprintf(&buf, "\n")
		}
	}
	return buf.String()
}
