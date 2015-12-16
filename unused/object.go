package unused

import (
	"fmt"
	"go/token"
	"golang.org/x/tools/go/types"
	"strings"
)

type objType int

const (
	None objType = iota
	FuncDeclaration
	FuncLiteral
	FuncMethod
	Variable
	Parameter
	Field
)

// Object contains all the necessary information to sort, categorize,
// and output unused code.
type Object interface {
	Pos() token.Position
	Type() objType
	Name() string
	Fullname() string
}

// UnusedThing represents a found unused function or identifier
type UnusedObject struct {
	Name     string
	Position token.Position
}

// String prints the position and name of the unused object.
func (uo UnusedObject) String() string {
	return fmt.Sprintf("%v:%v:%v: %v",
		trimGopath(uo.Position.Filename), uo.Position.Line, uo.Position.Column, uo.Name)
}

// ByPosition sorts unused objects by file/location.
// This type is a close copy of a similar sorter from the golint tool.
type ByPosition []UnusedObject

// Len method for sorting
func (p ByPosition) Len() int { return len(p) }

// Swap method for sorting
func (p ByPosition) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Less method for sorting on the Position
func (p ByPosition) Less(i, j int) bool {
	oi, oj := p[i].Position, p[j].Position

	if oi.Filename != oj.Filename {
		return oi.Filename < oj.Filename
	}
	if oi.Line != oj.Line {
		return oi.Line < oj.Line
	}
	if oi.Column != oj.Column {
		return oi.Column < oj.Column
	}

	// it's a bug if this even needs to be used
	return p[i].Name < p[j].Name
}

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
