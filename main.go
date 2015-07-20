package main

import (
	"flag"
	"fmt"
	"github.com/3rf/codecoroner/unused"
	"go/build"
	"golang.org/x/tools/go/buildutil"
	"os"
	"strings"
)

func main() {
	var ignoreList string
	ucf := unused.NewUnusedCodeFinder()
	flag.BoolVar(&(ucf.Verbose), "v", false,
		"prints extra information during execution to stderr")
	flag.BoolVar(&(ucf.IncludeTests), "tests", false, "include tests in the analysis")
	flag.StringVar(&(ignoreList), "ignore", "",
		"don't read files that contain the given comma-seperated strings (use to avoid /testdata, etc) ")
	flag.BoolVar(&(ucf.SkipMethodsAndFields), "skipmembers", false, "ignore unused struct fields and methods (idents only)")
	// hack for testing code with build flags
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", "a list of build tags")
	flag.Parse()
	// handle ignore list
	ucf.Ignore = strings.Split(ignoreList, ",")
	if len(ucf.Ignore) > 0 && ucf.Ignore[0] == "" {
		ucf.Ignore = nil
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Must specify either 'funcs' or 'idents' command. Run with -help for more info.")
		os.Exit(2)
	}
	command := flag.Arg(0)
	switch command {
	case "funcs", "functions":
		ucf.Idents = false
	case "idents", "identifiers":
		ucf.Idents = true
	default:
		fmt.Println("Must specify either 'funcs' or 'idents' command. Run with -help for more info.")
		os.Exit(2)
	}

	unusedFuncs, err := ucf.Run(flag.Args()[1:])
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	for _, f := range unusedFuncs {
		fmt.Printf("%s\n", f)
	}
}
