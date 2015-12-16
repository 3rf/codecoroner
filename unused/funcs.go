package unused

import (
	"fmt"
	"go/token"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"golang.org/x/tools/go/types"
	"strings"
)

// main method for running callgraph-based unused code analysis
func (ucf *UnusedCodeFinder) findUnusedFuncs() ([]UnusedObject, error) {
	ucf.buildFuncUses()
	// get the callgraph if we are doing this the hard way
	ucf.Logf("Running callgraph analysis on following packages: \n\t%v",
		strings.Join(ucf.pkgsAsArray(), "\n\t"))
	if err := ucf.getCallgraph(); err != nil {
		ucf.Errorf("Error running callgraph analysis: %v", err.Error())
		return nil, err
	}

	// finally, figure out which functions are not in the graph
	ucf.Logf("Scanning callgraph for unused functions")
	unusedFuncs := ucf.computeUnusedFuncs()
	return unusedFuncs, nil
}

func (ucf *UnusedCodeFinder) buildFuncUses() {
	ucf.Logf("Building a map for used functions")
	ucf.funcDefs = map[token.Pos]types.Object{} //XXX
	for _, info := range ucf.program.Imported {
		// find all declared funcs
		for _, obj := range info.Info.Defs {
			if obj != nil && obj.Pkg() != nil {
				if _, ok := obj.(*types.Func); ok {
					ucf.funcDefs[obj.Pos()] = obj
				}
			}
		}
	}

}

func (ucf *UnusedCodeFinder) getCallgraph() error {
	var buildMode ssa.BuilderMode
	if ucf.Verbose {
		buildMode = ssa.GlobalDebug
	}
	ssaP := ssautil.CreateProgram(ucf.program, buildMode)
	ssaP.Build()
	roots, err := ucf.getRoots(ssaP)
	if err != nil {
		return fmt.Errorf("error finding roots for callgraph analysis: %v", err)
	}
	res := rta.Analyze(roots, true)

	// build a simplified callgraph map for name->filenames
	ucf.funcUses = map[token.Pos]bool{} //XXX
	for node, _ := range res.Reachable {
		ucf.funcUses[node.Pos()] = true
	}
	return nil
}

func (ucf *UnusedCodeFinder) computeUnusedFuncs() []UnusedObject {
	unused := []UnusedObject{}
	for pos, fobj := range ucf.funcDefs {
		if !ucf.funcUses[pos] {
			unused = append(unused, UnusedObject{fobj.Name(), ucf.program.Fset.Position(pos)})
		}
	}
	return unused
}

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
