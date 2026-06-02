package nets

import (
	"testing"
)

func TestParseFormula(t *testing.T) {
	// We generate a new net that is needed only for its place names.
	net := Net{Pl: []string{"a", "b", "c", "d"}}

	testCases := []struct {
		input    string
		expected bool
	}{
		// Correct expressions for stress testing
		{"- (a >= b + c)", true},
		{"a + 3 = 10 /\\ b + c = d", true},
		{"(a + 3 <= 10 /\\ b + 1 + c + 5 = d) <=> (c + 1 + c + 4 <= d + d)", true},
		{"a + 3 <= 10 => - (b + c = d <=> c + 1 + 4 <= d)", true},
		{"(a = 1 /\\ - (b =2 \\/ (b =3))) /\\ (c = d + 0)", true},
		// Problematic expressions that should raise an error
		{"(a + 3 <= 10 /\\ b ) + c = d", false},
	}

	for _, tc := range testCases {
		f, err := net.ParseFormula(tc.input)
		t.Logf("\nin : %s\nout: %s\n\n", tc.input, f)
		if (err == nil) != tc.expected {
			t.Errorf("Expected %t, got %s", tc.expected, err)
		}
	}
}

func TestParseFormulaInvalid(t *testing.T) {
	// We generate a new net that is needed only for its place names.
	net := Net{Pl: []string{"a", "b", "c", "d", "e", "f", "g"}}

	testInvalid := []string{
		// Incomplete or malformed expressions
		"",
		" ",
		"a + b + c +",
		">",
		"=",
		"\\/",
		"=> a + b",
		"a + b \\/ c < d =>",
		// Parentheses issues
		")",
		"(",
		"()",
		"(a + b",
		"(a > b) )",
		"((a > b)) \\/ (c < d",
		"a + (b + c)",
		"(a + b) + c",
		// Undefined indentifiers and operators
		"(a > b) \\/ (c < d) \\/ (e = f) \\/ (g >= h)",
		"(a > b * 3) \\/ (c < d)",
		// Invalid operators or combinations
		"a + b > c < d", // Chained comparisons not allowed
		"a + b = c = d", // Chained equality not allowed
		// Comparaison missing or parentheses in expressions
		"(a)",
		"(a + b)",
		"a + (b + c)",
		"a + -b",
		"a++b",
		"a + +b",
		"a + b+",
		"1 + 2 +",
		// Invalid identifiers or constants
		"",
		"123a",
		"a b",
		"a+b",
		"a + b + c + d + e + f +", // Trailing operator
		"a + b > (c + d)",         // Parentheses in REXP
		// Nested or unbalanced logical operators
		// Mixed invalid operators and syntax
		"a + b > c \\/ d < e => f = g \\/ h != i", // `!=` not allowed, mixed operators
		// Empty or whitespace-only
		"   ",
		"\t",
		"\n",
		"\r\n",
	}

	for _, tc := range testInvalid {
		f, err := net.ParseFormula(tc)
		t.Logf("\nInp: \"%s\"\nErr: %s\n\n", tc, err)
		if err == nil {
			t.Errorf("Expected an error for \"%s\"\nfound: %s\n\n", tc, f)
		}
	}
}

func TestParseFormulaValid(t *testing.T) {
	// We generate a new net that is needed only for its place names.
	net := Net{Pl: []string{"a", "b", "c", "d", "e", "f", "g", "h"}}

	testInvalid := []string{
		// Correct expressions for stress testing
		"- (a >= b + c)",
		"a + 3 = 10 /\\ b + c = d",
		"(a + 3 <= 10 /\\ b + 1 + c + 5 = d) <=> (c + 1 + c + 4 <= d + d)",
		"a + 3 <= 10 => - (b + c = d <=> c + 1 + 4 <= d)",
		"(a = 1 /\\ - (b =2 \\/ (b =3))) /\\ (c = d + 0)",
		// Basic Forms (Comparisons)
		"a > b",
		"a >= b",
		"a < b",
		"a <= b",
		"a = b",
		"a + b > c",
		"a + b >= c",
		"a + 1 < b + 2",
		"a + b <= 3",
		"1 + 2 = 3",
		// negations
		"-(a > b)",
		"-(a >= b)",
		"-(a < b)",
		"-(a <= b)",
		"-(a = b)",
		"-(a + b >= c)",
		"-(a + 1 = b + 2)",
		"-(1 + 2 > 3)",
		// Conjunctions
		"(a > b) /\\ (c < d)",
		"(a >= b)  /\\ (c <= d)",
		"(a = b)  /\\ (c = d)",
		"(a + b > c)  /\\ (d < e)",
		"(a > b)  /\\ (c < d)  /\\ (e = f)",
		// Disjunctions
		"(a > b) \\/ (c < d)",
		"(a >= b) \\/ (c <= d)",
		"(a = b) \\/ (c = d)",
		"(a + b > c) \\/ (d < e)",
		"(a > b) \\/ (c < d) \\/ (e = f)",
		// Implications
		"(a > b) => (c < d)",
		"(a >= b) => (c <= d)",
		"(a = b) => (c = d)",
		"(a + b > c) => (d < e)",
		"-(a > b) => (c < d)",
		// Equivalences
		"(a > b) <=> (c < d)",
		"(a >= b) <=> (c <= d)",
		"(a = b) <=> (c = d)",
		"(a + b > c) <=> (d < e)",
		"-(a > b) <=> (c < d)",
		// Complex Combinations
		"((a > b) \\/ (c < d)) \\/ (e = f)",
		"((a + b >= c) \\/ (d < e)) => (f = g)",
		"-(a > b) \\/ (c < d)",
		"((a > b) \\/ (c < d)) <=> (e = f)",
		"((a + b >= c) \\/ (d < e)) => (f = g)",
		"(a > b) \\/ (c < d) \\/ (e = f) \\/ (g >= h)",
		"((a > b) => (c < d)) \\/ (e = f)",
		"((a > b) <=> (c < d)) \\/ (e = f)",
		// Nested Parentheses
		"((a > b))",
		"(((a > b)) \\/ (c < d))",
		"((a + b >= c) \\/ (d < e))",
		"((a > b) \\/ (c < d)) \\/ (e = f)",
		// Mixed Operators with Parentheses
		"((a + b >= c) => (d < e)) \\/ (f = g)",
		"((a > b) <=> (c < d)) => (e = f)",
		"-( (a > b) \\/ (c < d) )",
		"-( (a > b) \\/ (c < d) )",
	}

	for k, tc := range testInvalid {
		f, err := net.ParseFormula(tc)
		// t.Logf("\ntest[%d] = \"%s\" -- out: %s", k, tc, f)
		if err != nil {
			t.Errorf("\nExpected no errors for test[%d] = \"%s\"\nfound: %s\n\n", k, tc, f)
		}
	}
}

// func TestParseFormulaOne(t *testing.T) {
// 	// We generate a new net that is needed only for its place names.
// 	net := Net{Pl: []string{"a", "b", "c", "d", "e", "f", "g"}}
// 	tc := "(a > b) \\/ (c < d) \\/ (e = f) \\/ (g >= h)"
// 	_, err := net.ParseFormula(tc)
// 	if err != nil {
// 		t.Errorf("\nExpected no errors for \"%s\"\nFound: %s\n", tc, err)
// 	}
// }
