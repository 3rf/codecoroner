package main

import (
	"flag"
	"fmt"
	"github.com/3rf/go-unused-funcs/unused"
	"os"
)

func Two() int {
	return 2
}

func main() {
	flag.Parse()
	uff := unused.NewUnusedFunctionFinder()
	uff.Verbose = true
	//uff.IncludeAll = true
	unusedFuncs, err := uff.Run(flag.Args())
	if err != nil {
		os.Exit(1)
	}

	for _, f := range unusedFuncs {
		fmt.Printf("%v in '%v'\n", f.Name, f.File)
	}
}
