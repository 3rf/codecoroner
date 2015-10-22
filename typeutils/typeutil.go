package typeutils

import (
	"golang.org/x/tools/go/types"
)

// LookupFuncForParameter returns the func/var object containing
// the given var parameter. Returns nil if the var is not a parameter,
// or if the var is a parameter to an unaddressable func.
func LookupFuncForParameter(param *types.Var) types.Object {
	// parameters will always have a parent scope
	if param.Parent() == nil {
		return nil
	}

	// first check the pkg universe for the func
	scope := param.Pkg().Scope()
	f := lookForFunc(param, scope)
	if f != nil {
		return f
	}

	// now search the local scope
	scope = param.Parent().Parent()
	f = lookForFunc(param, scope)
	if f != nil {
		return f
	}

	return nil
}

// lookForFunc iterates through a scope and checks every func
// for a matching parameter. Returns nil if there is no matching func.
func lookForFunc(param *types.Var, scope *types.Scope) types.Object {
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		// grab the underlying signature from both vars and func declarations
		switch o := obj.(type) {
		case *types.Var, *types.Func:
			if sigObj, ok := o.Type().Underlying().(*types.Signature); ok {
				// iterate through the signature's parameters for a match
				for i := 0; i < sigObj.Params().Len(); i++ {
					if sigObj.Params().At(i).Pos() == param.Pos() {
						return obj
					}
				}
			}
		}
	}
	return nil

}

// LookupStructForField returns the struct definition that contains
// a given field as a *type.Var. Returns nil if the Var is not a field.
func LookupStructForField(field *types.Var) types.Object {
	if !field.IsField() {
		return nil
	}

	// first check the pkg universe for the struct
	scope := field.Pkg().Scope()
	st := lookForStruct(field, scope)
	if st != nil {
		return st
	}

	// find the innermost scope of the field, to make sure we grab the correct struct
	scope = scope.Innermost(field.Pos())
	if scope == nil {
		return nil
	}
	return lookForStruct(field, scope)
}

// lookForStruct crawls a scope for the struct with the given field.
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
		if f.Pos() == field.Pos() {
			return true
		}
		// if the field is a struct, dive into it
		if inner, ok := f.Type().Underlying().(*types.Struct); ok {
			if fieldInStruct(field, inner) {
				return true
			}
		}
	}
	return false
}
