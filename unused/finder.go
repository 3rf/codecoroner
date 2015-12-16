// The "unused" package wraps the go's static anaylsis packages and provides
// hooks for finding unused functions and identifiers in a codebase
package unused

import (
	"fmt"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type UnusedCodeFinder struct {
	// universal config options
	Idents    bool
	Ignore    []string
	Verbose   bool
	LogWriter io.Writer

	IncludeTests bool

	filesByCaller map[string][]token.Position
	pkgs          map[string]struct{}
	funcDefs      map[token.Pos]types.Object
	funcUses      map[token.Pos]bool
	funcs         []UnusedObject
	numFilesRead  int
	program       *loader.Program
}

func NewUnusedCodeFinder() *UnusedCodeFinder {
	return &UnusedCodeFinder{
		// init private storage
		pkgs:          map[string]struct{}{},
		filesByCaller: map[string][]token.Position{},
		funcs:         []UnusedObject{},
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

func (ucf *UnusedCodeFinder) loadProgram() error {
	var conf loader.Config
	_, err := conf.FromArgs(ucf.pkgsAsArray(), ucf.IncludeTests)
	if err != nil {
		return err
	}
	conf.AllowErrors = true
	ucf.Logf("Running loader")
	p, err := conf.Load()
	if err != nil {
		return err
	}
	ucf.program = p
	return nil
}

func (ucf *UnusedCodeFinder) readFuncsAndImportsFromFile(filename string) error {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}

	pkgName, err := getFullPkgName(filename)
	if err != nil {
		return fmt.Errorf("error getting main package path: %v", err)
	}
	ucf.AddPkg(pkgName)

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
	// strip the GOPATH. Error if this doesn't work.
	stripped := trimGopath(abs)
	if stripped != filename {
		return filepath.Dir(stripped), nil
	}
	// a check during initialization ensures that GOPATH != "" so this should be safe
	goPaths := filepath.SplitList(os.Getenv("GOPATH"))
	return "", fmt.Errorf("cd %q and try again", goPaths[len(goPaths)-1])
}

// trimGopath removes the GOPATH from a filepath, for simplicity
func trimGopath(filename string) string {
	goPaths := filepath.SplitList(os.Getenv("GOPATH"))
	for _, p := range goPaths {
		p = filepath.Join(p, "src") + string(filepath.Separator)
		if !strings.HasPrefix(filename, p) {
			continue
		}
		stripped := strings.TrimPrefix(filename, p)
		return stripped
	}
	return filename
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

func (ucf *UnusedCodeFinder) shouldIgnorePath(path string) bool {
	for _, ignoreToken := range ucf.Ignore {
		if strings.Contains(path, ignoreToken) {
			return true
		}
	}
	// skip test pkgs if -tests=false
	if !ucf.IncludeTests && strings.HasSuffix(path, "_test.go") {
		return true
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

func (ucf *UnusedCodeFinder) Run(fileArgs []string) ([]UnusedObject, error) {

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
	// TODO this can probably be replaced with a loader function
	for _, filename := range fileArgs {
		if strings.HasSuffix(filename, "/...") && isDir(filename[:len(filename)-4]) {
			// go tool ./... style
			if err := ucf.readDir(filename[:len(filename)-4]); err != nil {
				ucf.Errorf("Error reading '...': %v", err.Error())
				ucf.Errorf("Continuing...")
			}
		} else if isDir(filename) {
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

	if err := ucf.loadProgram(); err != nil {
		return nil, fmt.Errorf("error loading program data: %v", err)
	}
	if ucf.Idents {
		return ucf.findUnusedIdents()
	}
	return ucf.findUnusedFuncs()
}
