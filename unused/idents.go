package unused

import (
	"fmt"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"strings"
)

const Nothing = 6

var IgnoreMe = 77

// shorten the method name for nicer printing
//TODO unit test me XXX
func getCleanMethodName(f *types.Func) string {
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

func (uff *UnusedFuncFinder) findUnusedIdents() ([]UnusedThing, error) {
	var conf loader.Config
	_, err := conf.FromArgs(uff.pkgsAsArray(), true)
	if err != nil {
		return nil, fmt.Errorf("error loading program data: %v", err)
	}
	p, err := conf.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading program data: %v", err)
	}

	thingToUsage := map[string]int{}
	defined := map[string]struct{}{}

	for key, info := range p.Imported {
		if strings.Contains(key, ".") {
			// find all used idents
			for _, kind := range info.Info.Uses {
				if kind.Pkg() != nil {
					name := kind.Name()
					if asFunc, ok := kind.(*types.Func); ok {
						//special case for methods
						name = getCleanMethodName(asFunc)
					}
					id := fmt.Sprintf("%s.%s", kind.Pkg().Path(), name)
					thingToUsage[id] = thingToUsage[id] + 1
				}
			}
			// find all declared idents
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
					//fmt.Printf("%#v - %s\n", kind, name)
					if uff.ExportedOnly && !kind.Exported() {
						// skip unexported things if the user wishes
						continue
					}
					if asFunc, ok := kind.(*types.Func); ok {
						name = getCleanMethodName(asFunc)
					}
					id := fmt.Sprintf("%s.%s", kind.Pkg().Path(), name)
					defined[id] = struct{}{}
				}
			}
		}
	}
	unused := []UnusedThing{}
	// see which declared idents are not actually used
	for key, _ := range defined {
		if _, exists := thingToUsage[key]; !exists {
			unused = append(unused, UnusedThing{Name: key})
		}
	}
	return unused, nil

}
