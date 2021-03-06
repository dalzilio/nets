// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets_test

import (
	"fmt"
	"log"
	"os"

	"github.com/dalzilio/nets"
)

// This example shows the basic usage of the package: Parse a .net file and
// output the result on the standard output. Since the net has priorities, we
// also compute the transitive closure of its priority relation (a call to
// PrioClosure). Note that we print the number of places and transitions of the
// net, as a comment, but that we strip the original comments found in the file.
func Example_basic() {
	file, err := os.Open("testdata/demo.net")
	if err != nil {
		log.Fatal("error opening file:", err)
	}
	defer file.Close()

	net, err := nets.Parse(file)
	if err != nil {
		log.Fatal("parsing error: ", err)
	}
	if err := net.PrioClosure(); err != nil {
		log.Fatal("error with priorities: ", err)
	}
	fmt.Printf("%s", net)
	// Output:
	// #
	// # net demo
	// # 4 places, 7 transitions
	// #
	//
	// pl p0
	// pl p1
	// pl p4 : b
	// pl p2 (1)
	// tr t1 [0,1] p0 -> p1
	// tr t0 : a ]2,3[ p0*3 -> p1 p4
	// tr t3  p2 ->
	// tr t5 : {\{a\}}  p4 -> p0
	// tr t4  -> p4
	// tr t6  p4?1 ->
	// tr t2 : {b s} [0,0] p1?-4000 ->
	// pr t1 > t0
	// pr t3 > t1 t0 t2
	// pr t6 > t1 t0 t2
}

// This example shows how to use the result of parsing a .net file to find the
// number of transitions in a net.
func Example_countTransitions() {
	file, _ := os.Open("testdata/sokoban_3.net")
	net, err := nets.Parse(file)
	if err != nil {
		log.Fatal("parsing error: ", err)
	}
	fmt.Printf("net %s has %d transitions\n", net.Name, len(net.Tr))
	// Output:
	// net Sokoban has 452 transitions
}

// This example shows how to output a Net into a PNML Place/Transition file.
func Example_pnml() {
	file, _ := os.Open("testdata/abp.net")
	net, err := nets.Parse(file)
	if err != nil {
		log.Fatal("parsing error: ", err)
	}
	_ = net.Pnml(os.Stdout)
}
