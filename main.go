package main

import (
	"flag"
	"fmt"
	"github.com/3rf/codecoroner/unused"
	"go/build"
	"golang.org/x/tools/go/buildutil"
	"os"
	"sort"
	"strings"
)

func main() {
	var ignoreList string
	ucf := unused.NewUnusedCodeFinder()
	flag.BoolVar(&(ucf.Verbose), "v", false,
		"prints extra information during execution to stderr")
	flag.BoolVar(&(ucf.IncludeTests), "tests", false, "include tests in the analysis")
	flag.StringVar(&(ignoreList), "ignore", "",
		"don't read files that contain the given comma-separated strings (use to avoid /testdata, etc) ")
	// hack for testing code with build flags
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", "a list of build tags")
	flag.Parse()
	// handle ignore list
	ucf.Ignore = strings.Split(ignoreList, ",")
	if len(ucf.Ignore) > 0 && ucf.Ignore[0] == "" {
		ucf.Ignore = nil
	}

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"./..."} //TODO warn? Let's copy go vet...
	}

	unusedObjects, err := ucf.Run(args)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}
	ucf.Logf("") // ensure a newline before printing results if -v is on

	sort.Sort(unused.ByPosition(unusedObjects))
	for _, o := range unusedObjects {
		fmt.Printf("%s\n", o)
	}
}
