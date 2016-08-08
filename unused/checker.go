package unused

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func (ucf *UnusedCodeFinder) findUnusedObjects() []Object {
	unused := []Object{}
	ucf.Logf("Checking against a list of defined functions and variables")
	for _, info := range ucf.program.Imported {
		// find all declared funcs
		for _, obj := range info.Info.Defs {
			if obj != nil && obj.Pkg() != nil {
				name := obj.Name()
				if name == "_" || name == "main" || name == "init" || name == "." {
					continue
				}
				// We check a separate map for funcs and for everything else, as funcs are
				// checked using callgraph analysis and variables use a simple lookup.
				// See package notes for an explanation.
				if !ucf.funcUses[obj.Pos()] && !ucf.varUses[obj.Pos()] {
					unused = append(unused, ToObject(ucf.program, obj))
				}
			}
		}
	}
	return unused
}

func (ucf *UnusedCodeFinder) findFuncUses() error {
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
	for node, _ := range res.Reachable {
		ucf.funcUses[node.Pos()] = true
	}
	return nil
}

func (ucf *UnusedCodeFinder) findVarUses() {
	for _, info := range ucf.program.Imported {
		// check all used vars
		for _, obj := range info.Info.Uses {
			if obj.Pkg() != nil {
				if _, ok := obj.(*types.Func); !ok {
					ucf.varUses[obj.Pos()] = true
				}
			}
		}
	}
}
