package main

import (
	"flag"
	"github.com/3rf/go-oracle-tools/unused"
)

func main() {
	flag.Parse()

	uff := unused.NewUnusedFunctionFinder()
	uff.Run(flag.Args())
}
