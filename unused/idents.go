package unused

import (
	"fmt"
	"go/token"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"strings"
)

// shorten the method name for nicer printing and say if its a method
func handleMethodName(f *types.Func) string {
	name := f.Name()
	if strings.HasPrefix(f.FullName(), "(") {
		// it's a method! let's shorten the receiver!
		fullName := f.FullName()
		// second to last "."
		sepIdx := strings.LastIndex(fullName[:strings.LastIndex(fullName, ".")], ".")
		if sepIdx <= 0 { // rare special case
			return fullName
		}
		return fmt.Sprintf("(%s", fullName[sepIdx+1:])
	}
	return name
}

type ident struct {
	Name string
	Pos  token.Pos
}

func (ucf *UnusedCodeFinder) findUnusedIdents() ([]UnusedObject, error) {
	var conf loader.Config
	_, err := conf.FromArgs(ucf.pkgsAsArray(), ucf.IncludeTests)
	if err != nil {
		return nil, fmt.Errorf("error loading program data: %v", err)
	}
	conf.AllowErrors = true
	ucf.Logf("Running loader")
	p, err := conf.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading program data: %v", err)
	}

	identToUsage := map[ident]int{}
	defined := map[ident]struct{}{}

	for key, info := range p.Imported {
		if strings.Contains(key, ".") { //TODO do we need this if?

			// find all *used* idents
			for _, kind := range info.Info.Uses {
				if kind.Pkg() != nil {
					name := kind.Name()
					switch asType := kind.(type) {
					case *types.Func:
						//special case for methods
						name = handleMethodName(asType)
					}
					id := ident{Name: name, Pos: kind.Pos()}
					identToUsage[id] = identToUsage[id] + 1
				}
			}

			// find all *declared* idents
			for _, kind := range info.Info.Defs {
				if kind == nil {
					continue
				}
				if kind.Pkg() != nil {
					name := kind.Name()
					if name == "_" ||
						name == "main" ||
						name == "init" ||
						strings.HasPrefix(name, "Test") {
						continue
					}
					switch asType := kind.(type) {
					case *types.Func:
						name = handleMethodName(asType)
					}
					if name == "." {
						continue
					}
					id := ident{Name: name, Pos: kind.Pos()}
					defined[id] = struct{}{}
				}
			}
		}
	}
	unused := []UnusedObject{}
	// see which declared idents are not actually used
	for key, _ := range defined {
		if _, exists := identToUsage[key]; !exists {
			unused = append(unused, UnusedObject{
				Name:     key.Name,
				Position: p.Fset.Position(key.Pos),
			})
		}
	}
	return unused, nil
}
