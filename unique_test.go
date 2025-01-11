// Copyright 2025. Silvano DAL ZILIO. All rights reserved.
// Use of this source code is governed by the AGPL license
// that can be found in the LICENSE file.

package nets

import "testing"

func TestMarking(t *testing.T) {
	// Marking and Unique rely on the fact that places are listed in
	// order and that multiplicities are positive
	tables := []Marking{
		{},
		{{Pl: 3, Mult: 4}},
		{{Pl: 0, Mult: 3}, {Pl: 5, Mult: 4}},
		{{Pl: 6, Mult: 7}, {Pl: 8, Mult: 7}, {Pl: 10, Mult: 4}},
	}
	for _, input := range tables {
		k, _ := input.Unique()
		m := k.Marking()
		if !m.Equal(input) {
			t.Errorf("Equal(%v, %v) false", input, m)
		}
	}
}
