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

var validKinds = map[string]struct{}{
	unused.Miscs:      struct{}{},
	unused.Functions:  struct{}{},
	unused.Variables:  struct{}{},
	unused.Parameters: struct{}{},
	unused.Fields:     struct{}{},
}

func main() {
	var ignoreList, kindList string
	var ungrouped, shortNames bool
	ucf := unused.NewUnusedCodeFinder()
	flag.BoolVar(&(ucf.Verbose), "v", false,
		"prints extra information during execution to stderr")
	flag.BoolVar(&(ucf.IncludeTests), "tests", false, "include tests in the analysis")
	flag.BoolVar(&ungrouped, "ungrouped", false, "disable output type grouping")
	flag.BoolVar(&shortNames, "shortnames", false, "display identifiers with simpler syntax")
	flag.StringVar(&ignoreList, "ignore", "",
		"don't read files that contain these comma-separated strings (use to avoid /testdata, etc)")
	flag.StringVar(&kindList, "kinds", "",
		"comma-separated string of kinds of unused code to report (funcs,vars,fields,params,misc)")
	// hack for testing code with build flags
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", "a list of build tags")
	flag.Parse()

	// handle ignore list
	ucf.Ignore = strings.Split(ignoreList, ",")
	if len(ucf.Ignore) > 0 && ucf.Ignore[0] == "" {
		ucf.Ignore = nil
	}

	kinds := []string{unused.Functions, unused.Variables, unused.Fields, unused.Parameters}
	// parse and validate kind list
	if kindList != "" {
		kinds = strings.Split(kindList, ",")
		for _, k := range kinds {
			if _, ok := validKinds[k]; !ok {
				fmt.Printf(
					"ERROR: '%v' is not a valid kind. "+
						"Valid kinds are funcs, vars, fields, params, or misc.\n", k)
				os.Exit(2)
			}
		}
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

	ucf.Logf("Sorting and grouping %v results for types %v", len(unusedObjs), strings.Join(kinds, ","))
	grouping := unused.GroupObjects(unusedObjs)
	if ungrouped {
		ucf.Logf("Printing results withobjs grouping\n", len(unusedObjs))
		objs := []unused.Object{}
		for _, k := range kinds {
			objs = append(objs, grouping[k]...)
		}
		sort.Sort(unused.ByPosition(objs))
		printObjs(shortNames, objs)
	} else {
		ucf.Logf("Printing results with grouping\n")
		for _, k := range kinds {
			fmt.Println("[", strings.ToUpper(k), "]")
			objs := grouping[k]
			sort.Sort(unused.ByPosition(objs))
			printObjs(shortNames, objs)
			fmt.Println()
		}
	}
}

func printObjs(shortNames bool, objs []unused.Object) {
	for _, o := range objs {
		if shortNames {
			fmt.Println(unused.ObjectString(o))
		} else {
			fmt.Println(unused.ObjectFullString(o))
		}
	}
}
