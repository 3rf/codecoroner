package main

import (
	"flag"
	"fmt"
	"github.com/3rf/codecoroner/unused"
	"os"
)

func Two() int {
	return unused.Nothing
}

func main() {
	uff := unused.NewUnusedFunctionFinder()
	flag.BoolVar(&(uff.Verbose), "v", false,
		"prints extra information during execution to stderr")
	flag.BoolVar(&(uff.IncludeAll), "all", false,
		"includes all found packages in analysis, not just main packages")
	flag.StringVar(&(uff.Ignore), "ignore", "",
		"don't read files that match the given string (use to avoid /testdata, etc) ")
	flag.StringVar(&(uff.CallgraphJSON), "callgraphjson", "",
		"pass in a callgraph in json format instead of computing one")
	flag.BoolVar(&(uff.Idents), "idents", false, "")
	flag.BoolVar(&(uff.ExportedOnly), "exported", false, "")
	flag.BoolVar(&(uff.SkipMethods), "skipmethods", false, "")
	flag.Parse()

	unusedFuncs, err := uff.Run(flag.Args())
	if err != nil {
		os.Exit(unused.NICE)
	}

	for _, f := range unusedFuncs {
		fmt.Printf("%s\n", f)
	}
}
