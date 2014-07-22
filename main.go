package main

import (
	"flag"
	"github.com/3rf/go-unused-funcs/unused"
)

func main() {
	flag.Parse()

	uff := unused.NewUnusedFunctionFinder()
	uff.Verbose = true
	uff.Run(flag.Args())
}
