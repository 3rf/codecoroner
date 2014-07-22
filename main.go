package main

import (
	"flag"
	"fmt"
	"github.com/3rf/go-unused-funcs/unused"
)

func Two() int {
	return 2
}

func main() {
	flag.Parse()
	fmt.Println(Two())
	uff := unused.NewUnusedFunctionFinder()
	uff.Verbose = true
	uff.Run(flag.Args())
}
