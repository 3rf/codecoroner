package main

import (
	"flag"
	"github.com/3rf/go-unused-funcs/unused"
)

func main() {
	flag.Parse()

	uff := unused.NewUnusedFunctionFinder()
	uff.Run(flag.Args())
}
