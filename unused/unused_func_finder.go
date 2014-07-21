// The "unused" package wraps the go 'oracle' tool and provides
// hooks for finding unused functions in a codebase
package unused

import (
	//"code.google.com/p/go.tools/oracle"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type CGEntry struct {
	Name string `json:"name"`
	Pos  string `json:"pos"`
}
type Callgraph struct {
	Callgraph []CGEntry `json:"callgraph"`
}

type FoundFunc struct {
	Name string
	File string
}

type UnusedFuncFinder struct {
	CallgraphJSON []byte //ugh I hate this
	filesByCaller map[string][]string

	Verbose      bool
	ExcludeTests bool

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

func (uff *UnusedFuncFinder) ReadFuncsAndImportsFromFile(filename string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}

	// first, update the set of used packages
	for _, i := range f.Imports {
		uff.pkgs[i.Path.Value] = struct{}{}
	}

	// iterate over the AST, tracking dound functions
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
				// skip other cases
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

func (uff *UnusedFuncFinder) canReadSourceFile(filename string) bool {
	if uff.ExcludeTests && strings.HasSuffix(filename, "_test.go") {
		return false
	}
	if !strings.HasSuffix(filename, ".go") {
		return false
	}
	return true
}

func (uff *UnusedFuncFinder) ReadDir(dirname string) error {
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && uff.canReadSourceFile(path) {
			err = uff.ReadFuncsAndImportsFromFile(path)
		}
		return err
	})
	return err
}

func (uff *UnusedFuncFinder) Run(fileArgs []string) error {
	for _, filename := range fileArgs {
		if isDir(filename) {
			err := uff.ReadDir(filename)
			if err != nil {
				uff.Errorf("Error reading '%v' directory: %v", filename, err.Error())
				uff.Errorf("continuing...")
			}
		} else {
			if uff.canReadSourceFile(filename) {
				err := uff.ReadFuncsAndImportsFromFile(filename)
				if err != nil {
					uff.Errorf("Error reading '%v' file: %v", filename, err.Error())
					uff.Errorf("continuing...")
				}
			}
		}
	}
	fmt.Printf("%+v\n\n%+v", uff.pkgs, uff.funcs)
	return nil
}
