package unused

import (
	"fmt"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/loader"
	"os"
	"path/filepath"
	"strings"
)

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
