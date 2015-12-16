package unused

import (
	"fmt"
	"golang.org/x/tools/go/ssa"
)

// grab the callgraph roots from the passed in files. This is based on adonovan's
// code from https://github.com/golang/tools/blob/master/cmd/callgraph/main.go
func (ucf *UnusedCodeFinder) getRoots(prog *ssa.Program) ([]*ssa.Function, error) {
	pkgs := prog.AllPackages()
	mains := []*ssa.Package{}

	// create a test main if the user requests it
	if ucf.IncludeTests {
		if len(pkgs) > 0 {
			ucf.Logf("Building a test main for analysis")
			if main := prog.CreateTestMainPackage(pkgs...); main != nil {
				mains = append(mains, main)
			} else {
				ucf.Logf("WARNING: -tests flag specified, but no test files were located")
			}
		} else {
			return nil, fmt.Errorf("no packages specified")
		}
	}

	// then find *all* main packages
	for _, pkg := range pkgs {
		if pkg.Pkg.Name() == "main" {
			if pkg.Func("main") == nil {
				return nil, fmt.Errorf("no func main() in main package")
			}
			mains = append(mains, pkg)
		}
	}
	if len(mains) == 0 {
		return nil, fmt.Errorf("no main packages found")
	}

	roots := []*ssa.Function{}
	for _, root := range mains {
		roots = append(roots, root.Func("init"), root.Func("main"))
	}

	return roots, nil
}
