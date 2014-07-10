package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
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

func main() {
	callgraph := flag.String("calljson", "", "callgraph json file to compare against")
	flag.Parse()

	allFuncs := []FoundFunc{}

	for _, filename := range flag.Args() {
		if isDir(filename) {
			funcs := readDir(filename)
			allFuncs = append(allFuncs, funcs...)
		} else {
			if strings.Contains(filename, ".go") {
				funcs := getFuncs(filename)
				allFuncs = append(allFuncs, funcs...)
			}
		}
	}

	cg := Callgraph{}
	jsonCG, err := ioutil.ReadFile(*callgraph)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(jsonCG, &cg)
	if err != nil {
		panic(err)
	}
	//fmt.Println(cg.Callgraph)

	filesByCaller := map[string][]string{}
	for _, entry := range cg.Callgraph {
		idx := strings.LastIndex(entry.Name, ".") + 1
		if idx != 0 {
			filesByCaller[entry.Name[idx:]] = append(filesByCaller[entry.Name[idx:]], entry.Pos)
		}
	}

	for _, ff := range allFuncs {
		if !fInCG(ff, filesByCaller) {
			fmt.Printf("%v in %v NOT USED\n", ff.Name, ff.File)
		}
	}
}

func fInCG(ff FoundFunc, cgMap map[string][]string) bool {
	files, ok := cgMap[ff.Name]
	if !ok {
		return false
	}
	for _, path := range files {
		if strings.Contains(path, ff.File) {
			return true
		}
	}
	return false
}

func isDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

func readDir(dirname string) []FoundFunc {
	funcs := []FoundFunc{}
	filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
			funcs = append(funcs, getFuncs(path)...)
		}
		return err
	})
	return funcs
}

func getFuncs(fileName string) []FoundFunc {
	fset := token.NewFileSet() // positions are relative to fset

	found := []FoundFunc{}

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, fileName, nil, 0)
	if err != nil {
		fmt.Println(err)
		return found
	}

	// Print the imports from the file's AST.
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
			default:
				// skip other cases
				////fmt.Printf("FOUND: %s %s\n", fileName, s)
				found = append(found, FoundFunc{s, fileName})
			}
		}
		return true
	})

	return found
}
