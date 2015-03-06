// The "unused" package wraps the go 'oracle' and 'loader' tools and provides
// hooks for finding unused functions and identifiers in a codebase
package unused

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/oracle/serial"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var NICE = 2

func Three() int {
	return 3
}

// UnusedThing represents a found unused function or identifier
type UnusedThing struct {
	Name string
	File string
}

func (ut UnusedThing) String() string {
	if ut.File != "" {
		return fmt.Sprintf("%s in '%s'", ut.Name, ut.File)
	}
	return ut.Name
}

type UnusedCodeFinder struct {
	// universal config options
	Ignore     []string
	Verbose    bool
	IncludeAll bool
	LogWriter  io.Writer

	Idents               bool
	ExportedOnly         bool
	SkipMethodsAndFields bool
	IncludeTests         bool

	filesByCaller map[string][]string
	pkgs          map[string]struct{}
	funcs         []UnusedThing
	numFilesRead  int
	Callgraph     []serial.CallGraph
}

func NewUnusedCodeFinder() *UnusedCodeFinder {
	return &UnusedCodeFinder{
		// init private storage
		pkgs:          map[string]struct{}{},
		filesByCaller: map[string][]string{},
		funcs:         []UnusedThing{},
		// default to stderr; this can be overwritten before Run() is called
		LogWriter: os.Stderr,
	}
}

// TODO: move this log stuff to the bottom
// Logf is a one-off function for writing any verbose log output to
// stderr. There might be a more idiomatic way to do this in go...
func (ucf *UnusedCodeFinder) Logf(format string, v ...interface{}) {
	if ucf.Verbose {
		//ignore any errors in Fprintf for now
		fmt.Fprintf(ucf.LogWriter, format+"\n", v...)
	}
}

// Errorf is a one-off function for writing any error output to
// stderr. There might be a more idiomatic way to do this in go...
func (ucf *UnusedCodeFinder) Errorf(format string, v ...interface{}) {
	fmt.Fprintf(ucf.LogWriter, format+"\n", v...)
}

// AddPkg sets the package name as an entry in the package map,
// here the map holds no values and functions as a hash set
func (ucf *UnusedCodeFinder) AddPkg(pkgName string) {
	ucf.pkgs[pkgName] = struct{}{}
	ucf.Logf("Found pkg %v", pkgName)
}

func (ucf *UnusedCodeFinder) pkgsAsArray() []string {
	packages := make([]string, 0, len(ucf.pkgs))
	for pkg, _ := range ucf.pkgs {
		packages = append(packages, pkg)
	}
	return packages
}

func (ucf *UnusedCodeFinder) readFuncsAndImportsFromFile(filename string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}

	// check if this is a main packages or
	// if we want to analyze everything
	if f.Name.Name == "main" || ucf.IncludeAll || ucf.Idents {
		pkgName, err := getFullPkgName(filename)
		if err != nil {
			return fmt.Errorf("error getting main package path: %v", err)
		}
		ucf.AddPkg(pkgName)
	}

	// iterate over the AST, tracking found functions
	ast.Inspect(f, func(n ast.Node) bool {
		var s string
		switch node := n.(type) {
		case *ast.FuncDecl:
			s = node.Name.String()
		}
		if s != "" {
			switch {
			//TODO make this a helper
			case strings.Contains(s, "Test"):
			case s == "main":
			case s == "init":
			case s == "test":
			default:
				ucf.funcs = append(ucf.funcs, UnusedThing{s, filename})
			}
		}
		return true
	})

	ucf.numFilesRead++
	return nil
}

// helper for directory traversal
func isDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

// helper for grabbing package name from its folder
func getFullPkgName(filename string) (string, error) {
	abs, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	goPaths := filepath.SplitList(os.Getenv("GOPATH"))
	for _, p := range goPaths {
		p = filepath.Join(p, "src") + string(filepath.Separator)
		if !strings.HasPrefix(abs, p) {
			continue
		}
		stripped := strings.TrimPrefix(abs, p)
		return filepath.Dir(stripped), nil
	}
	// a check during initialization ensures that GOPATH != "" so this should be safe
	return "", fmt.Errorf("cd %q and try again", goPaths[len(goPaths)-1])
}

func (ucf *UnusedCodeFinder) canReadSourceFile(filename string) bool {
	if ucf.shouldIgnorePath(filename) {
		ucf.Logf("Ignoring path '%v'", filename)
		return false
	}
	if !strings.HasSuffix(filename, ".go") {
		return false
	}
	return true
}

func isNotStandardLibrary(pkg string) bool {
	// THIS IS WRONG I AM LEAVING IT IN AS A TEST
	return strings.ContainsRune(pkg, '.')
}

func (ucf *UnusedCodeFinder) shouldIgnorePath(path string) bool {
	for _, ignoreToken := range ucf.Ignore {
		if strings.Contains(path, ignoreToken) {
			return true
		}
	}
	return false
}

func (ucf *UnusedCodeFinder) readDir(dirname string) error {
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && ucf.canReadSourceFile(path) {
			err = ucf.readFuncsAndImportsFromFile(path)
		}
		return err
	})
	return err
}

func (ucf *UnusedCodeFinder) Run(fileArgs []string) ([]UnusedThing, error) {

	// do some basic sanity checks on system configuration
	if len(fileArgs) == 0 {
		return nil, fmt.Errorf(
			"no files supplied as arguments; must supply at least one file or directory")
	}
	if os.Getenv("GOPATH") == "" {
		return nil, fmt.Errorf("GOPATH not set")
	}

	// first, get all the file names and package imports
	ucf.Logf("Collecting declarations from source files")
	for _, filename := range fileArgs {
		if isDir(filename) {
			if err := ucf.readDir(filename); err != nil {
				ucf.Errorf("Error reading '%v' directory: %v", filename, err.Error())
				ucf.Errorf("Continuing...")
			}
		} else {
			if ucf.canReadSourceFile(filename) {
				if err := ucf.readFuncsAndImportsFromFile(filename); err != nil {
					ucf.Errorf("Error reading '%v' file: %v", filename, err.Error())
					ucf.Errorf("Continuing...")
				}
			}
		}
	}
	ucf.Logf("Parsed %v source files", ucf.numFilesRead)

	if ucf.Idents {
		return ucf.findUnusedIdents()
	}

	// then get the callgraph from the oracle
	ucf.Logf("Running callgraph analysis on following packages: \n\t%v",
		strings.Join(ucf.pkgsAsArray(), "\n\t"))
	if err := ucf.getCallgraphFromOracle(); err != nil {
		ucf.Errorf("Error getting results from oracle: %v", err.Error())
		return nil, err
	}

	// use that callgraph to build a callgraph->file map
	ucf.buildFileMap()

	// finally, figure out which functions are not in the graph
	ucf.Logf("Scanning callgraph for unused functions")
	unusedFuncs := ucf.computeUnusedFuncs()

	ucf.Logf("") // assure space between log output and results
	return unusedFuncs, nil
}
