package unused

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// main method for running callgraph-based unused code analysis
func (ucf *UnusedCodeFinder) findUnusedFuncs() ([]UnusedObject, error) {
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

func (ucf *UnusedCodeFinder) getCallgraph() error {
	var conf loader.Config
	_, err := conf.FromArgs(ucf.pkgsAsArray(), ucf.IncludeTests)
	if err != nil {
		return fmt.Errorf("error loading program data: %v", err)
	}
	conf.AllowErrors = true
	ucf.Logf("Running loader")
	p, err := conf.Load()
	if err != nil {
		return fmt.Errorf("error loading program data: %v", err)
	}
	var buildMode ssa.BuilderMode
	if ucf.Verbose {
		buildMode = ssa.GlobalDebug
	}
	ssaP := ssautil.CreateProgram(p, buildMode)
	ssaP.Build()
	roots, err := ucf.getRoots(ssaP)
	if err != nil {
		return fmt.Errorf("error finding roots for callgraph analysis: %v", err)
	}
	res := rta.Analyze(roots, true)

	// build a simplified callgraph map for name->filenames
	for node, _ := range res.Reachable {
		position := ssaP.Fset.Position(node.Pos())
		ucf.filesByCaller[node.Name()] = append(ucf.filesByCaller[node.Name()], position)
	}
	return nil
}

func (ucf *UnusedCodeFinder) isInCG(f UnusedObject) bool {
	for _, pos := range ucf.filesByCaller[f.Name] {
		if strings.Contains(pos.Filename, f.Position.Filename) {
			return true
		}
	}
	return false
}

func (ucf *UnusedCodeFinder) computeUnusedFuncs() []UnusedObject {
	unused := []UnusedObject{}
	for _, f := range ucf.funcs {
		if !ucf.isInCG(f) {
			unused = append(unused, f)
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
			for _, pkg := range pkgs {
				if main := prog.CreateTestMainPackage(pkg); main != nil {
					mains = append(mains, main)
				} else {
					ucf.Logf("WARNING: -tests flag specified, but no test files found for %s", pkg)
				}
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
