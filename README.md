# Nets

**Nets** is a Go library for parsing Petri nets, and Time Petri nets, written
using the textual description format of the [Tina
toolbox](http://projects.laas.fr/tina/). The format is defined in the section on
[the .net format](http://projects.laas.fr/tina/manuals/formats.html#2) described
in the [manual pages for
Tina](http://projects.laas.fr/tina/manuals/formats.html). The library also
defines an exported struct (a concrete type) for Petri nets.

[![Go Report Card](https://goreportcard.com/badge/github.com/dalzilio/nets)](https://goreportcard.com/report/github.com/dalzilio/nets)
[![GoDoc](https://godoc.org/github.com/dalzilio/mcc?status.svg)](https://godoc.org/github.com/dalzilio/nets)
[![Release](https://img.shields.io/github/v/release/dalzilio/rudd)](https://github.com/dalzilio/nets/releases)

## Installation

```bash
go get github.com/dalzilio/nets
```

## Usage

You can find some examples of code in the `*_test.go` files and some example of
`.net` files in directory `testdata`. The main function, `Parse`, returns a
`Net` struct from an `io.Reader`.

```go
package main

import (
  "fmt"
  "os"
  
  "github.com/dalzilio/nets"
)

func main() {
	file, _ := os.Open("testdata/sokoban_3.net")
	net, err := nets.Parse(file)
	if err != nil {
		log.Fatal("parsing error: ", err)
	}
	fmt.Printf("net %s has %d transitions\n", net.Name, len(net.Tr))
	// Output:
	// net Sokoban has 452 transitions
}
```

## Dependencies

The library has no dependencies outside of the standard Go library. It uses Go
modules and has been tested with Go 1.16.

## License

This software is distributed under the [GNU Affero GPL
v3](https://www.gnu.org/licenses/agpl-3.0.en.html). A copy of the license
agreement is found in the [LICENSE](./LICENSEmd) file.

## Authors

* **Silvano DAL ZILIO** -  [LAAS/CNRS](https://www.laas.fr/)
