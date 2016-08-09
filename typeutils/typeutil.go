package typeutils

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/loader"
)

const Anonymous = "<anonymous>"

// Program wraps a loader.Program to extend it with helper utilities.
func Program(p *loader.Program) program {
	return program{p}
}

type program struct{ *loader.Program }

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

// IsInsideTypeDefinition returns true if the object is part of the declaration of a type.
func (p program) IsInsideTypeDefinition(o types.Object) bool {
	path := p.astPath(o.Pos())
	for _, node := range path {
		if _, ok := node.(*ast.TypeSpec); ok {
			return true
		}
	}
	return false
}

func (program) IsMethod(v *types.Func) bool {
	if sig, ok := v.Type().(*types.Signature); ok {
		return sig.Recv() != nil
	}
	return false
}

func (program) RecieverForMethod(v *types.Func) string { //TESTME
	if sig, ok := v.Type().(*types.Signature); ok && sig != nil {
		typeStr := sig.Recv().Type().String()
		if i := strings.LastIndex(typeStr, "."); i-2 > 0 {
			return typeStr[i+1:]
		}
	}
	return ""
}

func (program) IsInterfaceMethod(v *types.Func) bool { //TESTME
	if sig, ok := v.Type().(*types.Signature); ok && sig != nil {
		u := sig.Recv().Type().Underlying()
		_, ok := u.(*types.Interface)
		return ok
	}
	return false
}

func (p program) IsStructField(v *types.Var) bool {
	// embedded fields have a slightly different ast syntax
	if p.IsEmbeddedField(v) {
		return true
	}
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

func (p program) IsEmbeddedField(v *types.Var) bool {
	path := p.astPath(v.Pos())
	if len(path) < 5 {
		return false
	}
	// embedded literal
	_, ok0 := path[0].(*ast.Ident)
	_, ok1 := path[1].(*ast.SelectorExpr)
	_, ok2 := path[2].(*ast.Field)
	_, ok3 := path[3].(*ast.FieldList)
	_, ok4 := path[4].(*ast.StructType)
	// embedded pointer
	_, _ok0 := path[0].(*ast.StarExpr)
	_, _ok1 := path[1].(*ast.Field)
	_, _ok2 := path[2].(*ast.FieldList)
	_, _ok3 := path[3].(*ast.StructType)
	return (ok0 && ok1 && ok2 && ok3 && ok4) || (_ok0 && _ok1 && _ok2 && _ok3)
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
