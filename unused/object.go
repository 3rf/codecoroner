package unused

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"

	"github.com/3rf/codecoroner/typeutils"
	"golang.org/x/tools/go/loader"
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
	Position() token.Position
	Name() string
	FullName() string
}

func ToObject(prog *loader.Program, o types.Object) Object {
	p := typeutils.Program(prog)
	switch ot := o.(type) {
	case *types.Func:
		if p.IsMethod(ot) {
			return &Func{fn: ot, position: p.Fset.Position(o.Pos())}
		}
		return &Func{fn: ot, position: p.Fset.Position(o.Pos())}
	default:
		return &Misc{o: o, position: p.Fset.Position(o.Pos())}
	}
}

// Func represents an unused function
type Func struct {
	fn       *types.Func
	position token.Position
}

func (f *Func) Position() token.Position { return f.position }
func (f *Func) Name() string             { return f.fn.Name() }
func (f *Func) FullName() string         { return f.Name() }

// Misc represents and un-handled unused identifier
type Misc struct {
	o        types.Object
	position token.Position
}

func (m *Misc) Position() token.Position { return m.position }
func (m *Misc) Name() string             { return m.o.Name() }
func (m *Misc) FullName() string         { return m.Name() }

func ObjectString(o Object) string {
	p := o.Position()
	return fmt.Sprintf("%v:%v:%v: %v", trimGopath(p.Filename), p.Line, p.Column, o.Name())
}

func ObjectFullString(o Object) string {
	p := o.Position()
	return fmt.Sprintf("%v:%v:%v: %v", trimGopath(p.Filename), p.Line, p.Column, o.FullName())
}

// ByPosition sorts unused objects by file/location.
// This type is a close copy of a similar sorter from the golint tool.
type ByPosition []Object

// Len method for sorting
func (p ByPosition) Len() int { return len(p) }

// Swap method for sorting
func (p ByPosition) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Less method for sorting on the Position
func (p ByPosition) Less(i, j int) bool {
	oi, oj := p[i].Position(), p[j].Position()

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
	return p[i].Name() < p[j].Name()
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
