// The "unused" package wraps the go 'oracle' tool and provides
// hooks for finding unused functions in a codebase
package unused

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/oracle"
	"golang.org/x/tools/oracle/serial"
)

//TODO rename to FuncEntry or something
type FoundFunc struct {
	Name string
	File string
}

type UnusedFuncFinder struct {
	Callgraph []serial.CallGraph

	Ignore        string
	Verbose       bool
	IncludeAll    bool
	LogWriter     io.Writer
	CallgraphJSON string // for setting user json input (hack?)

	filesByCaller map[string][]string
	pkgs          map[string]struct{}
	funcs         []FoundFunc
	numFilesRead  int
}

func NewUnusedFunctionFinder() *UnusedFuncFinder {
	return &UnusedFuncFinder{
		// init private storage
		pkgs:          map[string]struct{}{},
		filesByCaller: map[string][]string{},
		funcs:         []FoundFunc{},
		// default to stderr; this can be overwritten before Run() is called
		LogWriter: os.Stderr,
	}
}

// TODO: move this log stuff to the bottom
// Logf is a one-off function for writing any verbose log output to
// stderr. There might be a more idiomatic way to do this in go...
func (uff *UnusedFuncFinder) Logf(format string, v ...interface{}) {
	if uff.Verbose {
		//ignore any errors in Fprintf for now
		fmt.Fprintf(uff.LogWriter, format+"\n", v...)
	}
}

// Errorf is a one-off function for writing any error output to
// stderr. There might be a more idiomatic way to do this in go...
func (uff *UnusedFuncFinder) Errorf(format string, v ...interface{}) {
	fmt.Fprintf(uff.LogWriter, format+"\n", v...)
}

// AddPkg sets the package name as an entry in the package map,
// here the map holds no values and functions as a hash set
func (uff *UnusedFuncFinder) AddPkg(pkgName string) {
	uff.pkgs[pkgName] = struct{}{}
}

func (uff *UnusedFuncFinder) pkgsAsArray() []string {
	packages := make([]string, 0, len(uff.pkgs))
	for pkg, _ := range uff.pkgs {
		packages = append(packages, pkg)
	}
	return packages
}

func (uff *UnusedFuncFinder) getCallgraphFromOracle() error {
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

func (uff *UnusedFuncFinder) getCallgraphFromJSON(input io.Reader) error {
	serialRes := serial.Result{}
	dec := json.NewDecoder(input)
	err := dec.Decode(&serialRes)
	if err != nil {
		return fmt.Errorf("error reading json: %v", err)
	}
	if len(serialRes.Callgraph) == 0 {
		return fmt.Errorf("no callgraph found in json input")
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

	uff.numFilesRead++
	return nil
}

func (uff *UnusedFuncFinder) computeUnusedFuncs() []FoundFunc {
	unused := []FoundFunc{}
	for _, f := range uff.funcs {
		if !uff.isInCG(f) {
			unused = append(unused, f)
		}
	}
	return unused
}

func (uff *UnusedFuncFinder) isInCG(f FoundFunc) bool {
	files, ok := uff.filesByCaller[f.Name]
	if !ok {
		return false
	}
	for _, path := range files {
		if strings.Contains(path, f.File) {
			return true
		}
	}
	return false
}

func (uff *UnusedFuncFinder) buildFileMap() {
	for _, entry := range uff.Callgraph {
		//strip off the package name for simplicity
		//TODO, can this be left on? Try prepending func names with package?
		idx := strings.LastIndex(entry.Name, ".") + 1
		if idx != 0 {
			uff.filesByCaller[entry.Name[idx:]] = append(uff.filesByCaller[entry.Name[idx:]], entry.Pos)
		}
	}
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
	//TODO, make this nicer and add a test
	strippedGopath := strings.TrimPrefix(abs, filepath.Join(gopath, "/src/")+"/")
	return filepath.Dir(strippedGopath), nil
}

func (uff *UnusedFuncFinder) canReadSourceFile(filename string) bool {
	if uff.Ignore != "" && strings.Contains(filename, uff.Ignore) { //TODO regex
		uff.Logf("Ignoring path '%v'", filename)
		return false
	}
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

func (uff *UnusedFuncFinder) Run(fileArgs []string) ([]FoundFunc, error) {

	// do some basic sanity checks on system configuration
	if len(fileArgs) == 0 {
		uff.Errorf("Must supply at least one file as an argument")
		return nil, fmt.Errorf("no files supplied as arguments")
	}
	if os.Getenv("GOPATH") == "" {
		uff.Errorf("GOPATH environment varaible is not set")
		return nil, fmt.Errorf("GOPATH not set")
	}

	// first, get all the file names and package imports
	uff.Logf("Collecting func declarations from source files")
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
	uff.Logf("Parsed %v source files", uff.numFilesRead)

	// then get the callgraph from json or the oracle
	if uff.CallgraphJSON == "" {
		uff.Logf("Running callgraph analysis on following packages: \n\t%v",
			strings.Join(uff.pkgsAsArray(), "\n\t"))
		if err := uff.getCallgraphFromOracle(); err != nil {
			uff.Errorf("Error getting results from oracle: %v", err.Error())
			return nil, err
		}
	} else {
		uff.Logf("Reading callgraph from '%v'", uff.CallgraphJSON)
		jsonFile, err := os.Open(uff.CallgraphJSON)
		if err != nil {
			uff.Errorf("Error opening json file: %v", err.Error())
			return nil, err
		}
		err = uff.getCallgraphFromJSON(jsonFile)
		if err != nil {
			uff.Errorf("Error reading callgrapg from json file: %v", err.Error())
			return nil, err
		}
	}

	// use that callgraph to build a callgraph->file map
	uff.buildFileMap()

	// finally, figure out which functions are not in the graph
	uff.Logf("Scanning callgraph for unused functions")
	unusedFuncs := uff.computeUnusedFuncs()

	uff.Logf("") // assure space between log output and results
	return unusedFuncs, nil
}
