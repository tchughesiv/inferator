package main

import (
	"flag"
	"fmt"

	"github.com/tchughesiv/inferator/version"
)

var (
	operator = flag.Bool("operator", false, "get current operator version")
)

func main() {
	flag.Parse()
	if !*operator {
		fmt.Println("Operator version is " + version.Version)
	}
	if *operator {
		fmt.Println(version.Version)
	}
}
