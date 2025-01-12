// Copyright (c) 2024 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"unique"
)

// Handle is a unique identifier for a Marking. Basically this is the canonical,
// interned version (using go unique package) of a string representation of a
// Marking.
type Handle unique.Handle[string]

// Value returns a copy of the string value that produced the Handle.
func (h Handle) Value() string {
	return unique.Handle[string](h).Value()
}

// Unique returns a unique Handle from a marking. It only accepts positive
// markings where multiplicities can be cast into a uint32 value.
func (m Marking) Unique() (Handle, error) {
	var buf bytes.Buffer
	buf.Grow(8 * len(m))
	arr := make([]byte, 4)
	//
	// we assume that a place index is never greater than MaxInt32, which means
	// more than 2.147 billion places in the net !
	//
	for _, v := range m {
		if v.Mult < 0 {
			return Handle(unique.Make("")), fmt.Errorf("negative multiplicity")
		}
		if v.Mult >= math.MaxInt32 {
			return Handle(unique.Make("")), fmt.Errorf("multiplicity over MaxInt32")
		}
		binary.BigEndian.PutUint32(arr, uint32(v.Pl))
		buf.Write(arr)
		binary.BigEndian.PutUint32(arr, uint32(v.Mult))
		buf.Write(arr)
	}
	return Handle(unique.Make(buf.String())), nil
}

// Marking returns the marking associated with a marking Handle
func (mk Handle) Marking() Marking {
	m := Marking{}
	// We use the fact that places occuring in markings are in increasing
	// order
	s := []byte(mk.Value())
	a := Atom{}
	i := 0
	for i < len(s) {
		a.Pl = int(binary.BigEndian.Uint32(s[i : i+4]))
		a.Mult = int(binary.BigEndian.Uint32(s[i+4 : i+8]))
		m = append(m, a)
		i += 8
	}
	return m
}
