package unused

import (
	"fmt"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"strings"
)

// shorten the method name for nicer printing and say if its a method
func handleMethodName(f *types.Func) (string, bool) {
	name := f.Name()
	if strings.HasPrefix(f.FullName(), "(") {
		// it's a method! let's shorten the receiver!
		fullName := f.FullName()
		// second to last "."
		sepIdx := strings.LastIndex(fullName[:strings.LastIndex(fullName, ".")], ".")
		if sepIdx <= 0 { // rare special case
			return fullName, true
		}
		return fmt.Sprintf("(%s", fullName[sepIdx+1:]), true
	}
	return name, false
}

// add a naming indicator that something is a field and say if its a field
func handleStructField(v *types.Var) (string, bool) {
	name := v.Name()
	if v.IsField() { // No way to get the actual ownder, why???
		name = name + " [struct field]"
		return name, true
	}
	return name, false
}

func (ucf *UnusedCodeFinder) findUnusedIdents() ([]UnusedThing, error) {
	var conf loader.Config
	_, err := conf.FromArgs(ucf.pkgsAsArray(), ucf.IncludeTests)
	if err != nil {
		return nil, fmt.Errorf("error loading program data: %v", err)
	}
	conf.AllowErrors = true //XXX make this configurable?
	ucf.Logf("Running loader...")
	p, err := conf.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading program data: %v", err)
	}

	thingToUsage := map[string]int{}
	defined := map[string]struct{}{}

	for key, info := range p.Imported {
		if strings.Contains(key, ".") {

			// find all *used* idents
			for _, kind := range info.Info.Uses {
				if kind.Pkg() != nil {
					name := kind.Name()
					switch asType := kind.(type) {
					case *types.Func:
						//special case for methods
						name, _ = handleMethodName(asType)
					case *types.Var:
						name, _ = handleStructField(asType)
					}
					id := fmt.Sprintf("%s.%s", kind.Pkg().Path(), name)
					thingToUsage[id] = thingToUsage[id] + 1
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
					if ucf.ExportedOnly && !kind.Exported() {
						// skip unexported things if the user wishes
						continue
					}
					switch asType := kind.(type) {
					case *types.Func:
						var isMethod bool
						name, isMethod = handleMethodName(asType)
						if ucf.SkipMethodsAndFields && isMethod {
							continue
						}
					case *types.Var:
						var isField bool
						name, isField = handleStructField(asType)
						if ucf.SkipMethodsAndFields && isField {
							continue
						}
					}
					if name == "." {
						continue
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
