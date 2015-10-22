package typeutil

import (
	//"fmt"
	//	"go/token"
	"golang.org/x/tools/go/types"
)

func LookupStructForField(field *types.Var) types.Object {
	if !field.IsField() {
		return nil
	}
	pkg := field.Pkg()
	scope := pkg.Scope()
	if scope == nil {
		return nil
	}

	// first check the pkg Universe for the field
	st := lookForStruct(field, scope)
	if st != nil {
		return st
	}

	// find the innermost scope of the field, to make sure we grab the correct field
	scope = scope.Innermost(field.Pos())
	if scope == nil {
		return nil
	}
	return lookForStruct(field, scope)
}

func lookForStruct(field *types.Var, scope *types.Scope) types.Object {
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if tnObj, ok := obj.(*types.TypeName); ok {
			if stObj, ok := tnObj.Type().Underlying().(*types.Struct); ok {
				if fieldInStruct(field, stObj) {
					return obj
				}
			}
		}
	}
	return nil
}

// fieldInStruct recursivesly checks for the field var inside of a struct.
func fieldInStruct(field *types.Var, s *types.Struct) bool {
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		if inner, ok := f.Type().Underlying().(*types.Struct); ok {
			if fieldInStruct(field, inner) {
				return true
			}
		}
		if f.Pos() == field.Pos() {
			return true
		}
	}
	return false
}
