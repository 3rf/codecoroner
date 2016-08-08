package typeutils

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/loader"
)

const Anonymous = "<anonymous>"

var _ fmt.Stringer //TODO

func Program(p *loader.Program) program {
	return program{p}
}

type program struct {
	*loader.Program
}

// helper for generating an ast path for a single object
func (p program) astPath(pos token.Pos) []ast.Node {
	_, path, _ := p.PathEnclosingInterval(pos, pos)
	return path
}

func (p program) FuncForParameter(v *types.Var) string {
	path := p.astPath(v.Pos())
	if len(path) < 4 {
		return ""
	}
	switch f := path[3].(type) {
	case *ast.FuncDecl:
		return f.Name.Name
	case *ast.FuncType:
		fmt.Printf(":::: %#v\n", path) //FIXME
		if len(path) >= 5 {
			if assign, ok := path[5].(*ast.AssignStmt); ok && len(assign.Lhs) > 0 {
				if lhs, ok := assign.Lhs[0].(*ast.Ident); ok {
					return lhs.Name
				}
			}
		}
		return Anonymous
	default:
		return ""
	}
}

func (program) IsMethod(v *types.Func) bool {
	if sig, ok := v.Type().(*types.Signature); ok {
		return sig.Recv() != nil
	}
	return false
}

func (p program) IsParameter(v *types.Var) bool {
	// parameters are always [Ident] in [Fields] in [FieldLists] in [FuncTypes or FuncDecl]
	path := p.astPath(v.Pos())
	if len(path) < 4 {
		return false
	}
	_, ok0 := path[0].(*ast.Ident)
	_, ok1 := path[1].(*ast.Field)
	_, ok2 := path[2].(*ast.FieldList)
	_, ok3FT := path[3].(*ast.FuncType)
	_, ok3FD := path[3].(*ast.FuncDecl)
	return ok0 && ok1 && ok2 && (ok3FT || ok3FD)
}

func (p program) IsStructField(v *types.Var) bool {
	// struct fields are always [Ident] in [Fields] in [FieldLists] in [StructType]
	path := p.astPath(v.Pos())
	if len(path) < 4 {
		return false
	}
	_, ok0 := path[0].(*ast.Ident)
	_, ok1 := path[1].(*ast.Field)
	_, ok2 := path[2].(*ast.FieldList)
	_, ok3 := path[3].(*ast.StructType)
	return ok0 && ok1 && ok2 && ok3
}

func (p program) StructForField(v *types.Var) string {
	return p.structForField(v.Pos())
}

// recursive helper for grabbing a struct field's struct
func (p program) structForField(pos token.Pos) string {
	path := p.astPath(pos)
	if len(path) < 5 {
		return ""
	}
	switch s := path[4].(type) {
	case *ast.TypeSpec: // struct type declarations
		return s.Name.Name
	case *ast.ValueSpec: // anonymous structs using "var" syntax
		if len(s.Names) > 0 {
			return s.Names[0].Name
		}
		return ""
	case *ast.Field: // struct is a field of another struct
		if len(s.Names) > 0 {
			//fmt.Printf(":::: %#v\n", path) //FIXME
			return p.structForField(path[4].Pos()) + "." + s.Names[0].Name
		}
		return ""
	case *ast.CompositeLit: // local anonymous struct
		if len(path) >= 6 {
			// grab the left-hand side of the assign statment (i.e. "X := ...")
			if assign, ok := path[5].(*ast.AssignStmt); ok && len(assign.Lhs) > 0 {
				if lhs, ok := assign.Lhs[0].(*ast.Ident); ok {
					return lhs.Name
				}
			}
		}
		return Anonymous
	default:
		//fmt.Printf(">>> %#v\n", path[4]) //FIXME
		return ""
	}

}
