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
	flag.BoolVar(&(uff.ExportedOnly), "exported", false, "")
	flag.BoolVar(&(uff.SkipMethodsAndFields), "skipmembers", false, "")
	flag.BoolVar(&(uff.IncludeTests), "tests", false, "")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("Must specify either 'funcs' or 'idents' command. Run with -help for more info.")
		os.Exit(2)
	}
	command := flag.Arg(0)
	switch command {
	case "funcs", "functions":
		uff.Idents = false
	case "idents", "identifiers":
		uff.Idents = true
	default:
		fmt.Println("Must specify either 'funcs' or 'idents' command. Run with -help for more info.")
		os.Exit(2)
	}

	unusedFuncs, err := uff.Run(flag.Args()[1:])
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	for _, f := range unusedFuncs {
		fmt.Printf("%s\n", f)
	}
}
