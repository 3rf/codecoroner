package unused

import (
	"fmt"
	"go/token"
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
func (ut UnusedObject) String() string {
	return fmt.Sprintf("%v:%v:%v: %v",
		trimGopath(ut.Position.Filename), ut.Position.Line, ut.Position.Column, ut.Name)
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
