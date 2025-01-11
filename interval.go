// Copyright 2025. Silvano DAL ZILIO. All rights reserved.
// Use of this source code is governed by the AGPL license
// that can be found in the LICENSE file.

package nets

import (
	"bytes"
	"fmt"
	"strconv"
)

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

func (b Bound) String() string {
	switch b.Bkind {
	case BINFTY:
		return "w"
	case BCLOSE:
		return fmt.Sprintf("=%d", b.Value)
	default:
		return fmt.Sprintf("x%d", b.Value)
	}
}

// PrintLowerBound returns a textual representation of a time interval bound
// used as a lower bound constraint, such as "4 <" or "5 ≤". We return the
// string "∞" if b is infinite (which should not happen in practice).
func (b Bound) PrintLowerBound() string {
	switch b.Bkind {
	case BINFTY:
		return "∞"
	case BCLOSE:
		return fmt.Sprintf("%d ≤", b.Value)
	default:
		return fmt.Sprintf("%d <", b.Value)
	}
}

// PrintUpperBound is the dual  of PrintLowerBound and returns a representation
// of a time interval bound used as a lower bound constraint, such as "< 4" or
// "≤ 5". We return the string "< ∞" if b is infinite.
func (b Bound) PrintUpperBound() string {
	switch b.Bkind {
	case BINFTY:
		return "< ∞"
	case BCLOSE:
		return fmt.Sprintf("≤ %d", b.Value)
	default:
		return fmt.Sprintf("< %d", b.Value)
	}
}

// TimeInterval is the type of time intervals.
type TimeInterval struct {
	Left, Right Bound
}

func (i *TimeInterval) String() string {
	if i.Left.Bkind == BINFTY {
		// it means interval was never set
		return "[0,w["
	}
	var buf bytes.Buffer
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

/*****************************************************************************/

// BSubstract computes the diference, b1 - b2, between its time bounds
// parameters. We return an infinite bound when b2 is infinite.
func BSubstract(b1, b2 Bound) Bound {
	if b1.Bkind == BINFTY {
		return b1
	}
	if b2.Bkind == BINFTY {
		return b2
	}
	diff := b1.Value - b2.Value
	if b1.Bkind == BOPEN || b2.Bkind == BOPEN {
		return Bound{BOPEN, diff}
	}
	return Bound{BCLOSE, diff}
}

// BAdd returns the sum of two time bounds.
func BAdd(b1, b2 Bound) Bound {
	if b1.Bkind == BINFTY {
		return b1
	}
	if b2.Bkind == BINFTY {
		return b1
	}
	add := b1.Value + b2.Value
	if b1.Bkind == BOPEN || b2.Bkind == BOPEN {
		return Bound{BOPEN, add}
	}
	return Bound{BCLOSE, add}
}

// BCompare returns an integer comparing two bounds. The result will be 0 if a
// and b are equal, negative if a < b, and positive otherwise. We return the
// difference between the bounds values, with some exceptions. We always return
// -1 when b is infinite or when a and b have same values, but a is open whereas
// b is closed. For intance, the bound [1,.. is considered strictly greater than
// ]1,.. with our choice. We return -1 in the symetric cases.
func BCompare(a, b Bound) int {
	if b.Bkind == BINFTY {
		return -1
	}
	if a.Bkind == BINFTY {
		return +1
	}
	if a.Value != b.Value {
		return a.Value - b.Value
	}
	if a.Bkind == b.Bkind {
		return 0
	}
	if a.Bkind == BOPEN && b.Bkind == BCLOSE {
		return -1
	}
	return +1
}

// BIsPositive returns true if b1 is greater or equal to 0.
func BIsPositive(b Bound) bool {
	if b.Value > 0 || b.Bkind == BINFTY {
		return true
	}
	if b.Value == 0 && b.Bkind == BCLOSE {
		return true
	}
	return false
}

// BMax returns the max of a and b.
func BMax(a, b Bound) Bound {
	if BCompare(a, b) <= 0 {
		return b
	}
	return a
}

// BMin returns the min of a and b.
func BMin(a, b Bound) Bound {
	if BCompare(a, b) <= 0 {
		return a
	}
	return b
}

/*****************************************************************************/

// Trivial is true if the time interval i is of the form [0, w[ or if the
// interval is un-initialized (meaning the left part of the interval is of kind
// BINFTY)
func (i *TimeInterval) Trivial() bool {
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

// intersectWith sets interval i to the intersection of i and j. We return an
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
