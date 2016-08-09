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
	var ungrouped, shortNames bool
	ucf := unused.NewUnusedCodeFinder()
	flag.BoolVar(&(ucf.Verbose), "v", false,
		"prints extra information during execution to stderr")
	flag.BoolVar(&(ucf.IncludeTests), "tests", false, "include tests in the analysis")
	flag.BoolVar(&ungrouped, "ungrouped", false, "disable output type grouping")
	flag.BoolVar(&shortNames, "shortnames", false, "display identifiers with simpler syntax")
	flag.StringVar(&ignoreList, "ignore", "",
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

	unusedObjs, err := ucf.Run(args)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	sort.Sort(unused.ByPosition(unusedObjs))
	if ungrouped {
		// TODO types should work without grouping
		ucf.Logf("Printing %v results, ungrouped\n", len(unusedObjs)) // ensure a newline before printing results if -v is on
		for _, o := range unusedObjs {
			fmt.Println(unused.ObjectFullString(o))
		}
	} else {
		grouping := []string{unused.Functions, unused.Variables, unused.Fields, unused.Parameters}
		ucf.Logf("Printing %v results for types %v\n",
			len(unusedObjs), strings.Join(grouping, ","))
		printGrouped(
			shortNames,
			grouping, //FIXME
			unusedObjs)
	}
}

func printGrouped(shortNames bool, types []string, objs []unused.Object) {
	grouped := unused.GroupObjects(objs)
	// todo use list
	for _, t := range types {
		fmt.Println("[", strings.ToUpper(t), "]")
		for _, o := range grouped[t] {
			if shortNames {
				fmt.Println(unused.ObjectString(o))
			} else {
				fmt.Println(unused.ObjectFullString(o))
			}
		}
		fmt.Println()
	}
}
