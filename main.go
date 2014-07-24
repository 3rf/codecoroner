package main

import (
	"flag"
	"github.com/3rf/go-unused-funcs/unused"
)

func Two() int {
	return 2
}

func main() {
	flag.Parse()
	uff := unused.NewUnusedFunctionFinder()
	uff.Verbose = true
	//uff.IncludeAll = true
	uff.Run(flag.Args())
}
