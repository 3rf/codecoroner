// The "unused" package wraps the go's static anaylsis packages and provides
// hooks for finding unused functions and identifiers in a codebase
package unused

import (
	"fmt"
	"go/token"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"io"
	"os"
	"strings"
)

type UnusedCodeFinder struct {
	// universal config options
	Ignore    []string
	Verbose   bool
	LogWriter io.Writer

	IncludeTests bool

	filesByCaller map[string][]token.Position
	pkgs          map[string]struct{}
	funcDefs      map[token.Pos]types.Object
	varUses       map[token.Pos]bool
	funcUses      map[token.Pos]bool
	numFilesRead  int
	program       *loader.Program
}

func NewUnusedCodeFinder() *UnusedCodeFinder {
	return &UnusedCodeFinder{
		// init private storage
		pkgs:          map[string]struct{}{},
		filesByCaller: map[string][]token.Position{},
		funcDefs:      map[token.Pos]types.Object{},
		funcUses:      map[token.Pos]bool{},
		varUses:       map[token.Pos]bool{},
		// default to stderr; this can be overwritten before Run() is called
		LogWriter: os.Stderr,
	}
}

// TODO: use log stdlib pkg
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

func (ucf *UnusedCodeFinder) Run(fileArgs []string) ([]Object, error) {
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
	if err := ucf.findFuncUses(); err != nil {
		return nil, fmt.Errorf("error computing function reachability: %v", err)
	}
	ucf.findVarUses()
	return ucf.findUnusedObjects(), nil
}
