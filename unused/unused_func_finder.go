// The "unused" package wraps the go 'oracle' tool and provides
// hooks for finding unused functions in a codebase
package unused

import (
	"code.google.com/p/go.tools/oracle"
	"code.google.com/p/go.tools/oracle/serial"
	//"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type FoundFunc struct {
	Name string
	File string
}

type UnusedFuncFinder struct {
	Callgraph []serial.CallGraph

	filesByCaller map[string][]string

	Verbose       bool
	IncludeAll    bool
	CallgraphJSON []byte // for setting user json input (hack?)

	pkgs  map[string]struct{}
	funcs []FoundFunc
}

func NewUnusedFunctionFinder() *UnusedFuncFinder {
	return &UnusedFuncFinder{
		pkgs:          map[string]struct{}{},
		filesByCaller: map[string][]string{},
		funcs:         []FoundFunc{},
	}
}

// TODO: move this log stuff to the bottom
// Logf is a one-off function for writing any verbose log output to
// stderr. There might be a more idiomatic way to do this in go...
func (uff *UnusedFuncFinder) Logf(format string, v ...interface{}) {
	if uff.Verbose {
		//ignore any errors in Fprintf for now
		fmt.Fprintf(os.Stderr, format+"\n", v...)
	}
}

// Errorf is a one-off function for writing any error output to
// stderr. There might be a more idiomatic way to do this in go...
func (uff *UnusedFuncFinder) Errorf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}

// AddPkg sets the package name as an entry in the package map,
// here the map holds no values and functions as a hash set
func (uff *UnusedFuncFinder) AddPkg(pkgName string) {
	uff.pkgs[pkgName] = struct{}{}
}

func (uff *UnusedFuncFinder) pkgsAsArray() []string {
	packages := make([]string, 0, len(uff.pkgs))
	for pkg, _ := range uff.pkgs {
		if isNotStandardLibrary(pkg) {
			packages = append(packages, pkg)
		}
	}
	return packages
}

func (uff *UnusedFuncFinder) getCallgraphJSONFromOracle() error {
	res, err := oracle.Query(uff.pkgsAsArray(), "callgraph", "", nil, &build.Default, true)
	if err != nil {
		return err
	}
	serialRes := res.Serial()
	if serialRes.Callgraph == nil {
		return fmt.Errorf("no callgraph present in oracle results")
	}
	uff.Callgraph = serialRes.Callgraph
	return nil
}

func (uff *UnusedFuncFinder) readFuncsAndImportsFromFile(filename string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}

	// check if this is a main packages or
	// if we want to analyze everything
	if f.Name.Name == "main" || uff.IncludeAll {
		pkgName, err := getFullPkgName(filename)
		if err != nil {
			return fmt.Errorf("error getting main package path: %v", err)
		}
		uff.AddPkg(pkgName)
	}

	// update the set of used packages if we want to throw them all in
	// FIXME: do we want to instead just grab the pkg path of all files???
	if uff.IncludeAll {
		for _, i := range f.Imports {
			// strip quotes from package string for safe passing to the oracle
			pkg := strings.Trim(i.Path.Value, "\"`'")
			uff.AddPkg(pkg)
		}
	}

	// iterate over the AST, tracking found functions
	ast.Inspect(f, func(n ast.Node) bool {
		var s string
		switch n.(type) {
		case *ast.FuncDecl:
			asFunc := n.(*ast.FuncDecl)
			s = asFunc.Name.String()
		}
		if s != "" {
			switch {
			case strings.Contains(s, "Test"):
			case s == "main":
			case s == "init":
			case s == "test":
			default:
				uff.funcs = append(uff.funcs, FoundFunc{s, filename})
			}
		}
		return true
	})

	return nil
}

// helper for directory traversal
func isDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

// helper for grabbing package name from its folder
func getFullPkgName(filename string) (string, error) {
	gopath := os.Getenv("GOPATH")
	abs, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	strippedGopath := strings.TrimPrefix(abs, gopath+"/src/")
	return filepath.Dir(strippedGopath), nil
}

func (uff *UnusedFuncFinder) canReadSourceFile(filename string) bool {
	if !strings.HasSuffix(filename, ".go") {
		return false
	}
	return true
}

func isNotStandardLibrary(pkg string) bool {
	// THIS IS WRONG
	// FIXME HACK HACK HACK
	return strings.ContainsRune(pkg, '.')
}

func (uff *UnusedFuncFinder) readDir(dirname string) error {
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && uff.canReadSourceFile(path) {
			err = uff.readFuncsAndImportsFromFile(path)
		}
		return err
	})
	return err
}

func (uff *UnusedFuncFinder) Run(fileArgs []string) error {
	// first, get all the file names and package imports
	for _, filename := range fileArgs {
		if isDir(filename) {
			if err := uff.readDir(filename); err != nil {
				uff.Errorf("Error reading '%v' directory: %v", filename, err.Error())
				uff.Errorf("Continuing...")
			}
		} else {
			if uff.canReadSourceFile(filename) {
				if err := uff.readFuncsAndImportsFromFile(filename); err != nil {
					uff.Errorf("Error reading '%v' file: %v", filename, err.Error())
					uff.Errorf("Continuing...")
				}
			}
		}
	}

	// then get the callgraph from json or the oracle
	if uff.CallgraphJSON == nil {
		uff.Logf("Running callgraph analysis on following packages: \n\t%v",
			strings.Join(uff.pkgsAsArray(), "\n\t"))
		if err := uff.getCallgraphJSONFromOracle(); err != nil {
			uff.Errorf("Error getting results from oracle: %v", err.Error())
			return err
		}
	}

	fmt.Printf("%v", uff.Callgraph)
	return nil
}
