package unused

import (
	"fmt"
	"go/token"
	"go/types"

	"github.com/3rf/codecoroner/typeutils"
	"golang.org/x/tools/go/loader"
)

const (
	Miscs      = "misc"
	Functions  = "funcs"
	Variables  = "vars"
	Parameters = "params"
	Fields     = "fields"
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
	position := p.Fset.Position(o.Pos())
	switch ot := o.(type) {
	case *types.Var:
		if p.IsStructField(ot) {
			return &Field{o: o, position: position, s: p.StructForField(ot)}
		}
		if p.IsParameter(ot) {
			return &Param{o: o, position: position, f: p.FuncForParameter(ot)}
		}
		return &Var{o: o, position: position}
	case *types.Func:
		if p.IsMethod(ot) {
			if p.IsInterfaceMethod(ot) {
				return &Iface{fn: ot, position: position, iface: p.RecieverForMethod(ot)}
			}
			return &Func{fn: ot, position: position, recv: p.RecieverForMethod(ot)}
		}
		return &Func{fn: ot, position: position}
	default:
		return &Misc{o: o, position: position}
	}
}

// Func represents an unused function
type Func struct {
	fn       *types.Func
	position token.Position
	recv     string
}

func (f *Func) Position() token.Position { return f.position }
func (f *Func) Name() string             { return f.fn.Name() }
func (f *Func) FullName() string {
	if f.recv != "" {
		return fmt.Sprintf("(%v).%v", f.recv, f.Name())
	}
	return f.Name()
}

// Iface represents an interface method.
// TODO figure out how to properly handle these, if it is even possible
type Iface struct {
	fn       *types.Func
	position token.Position
	iface    string
}

func (i *Iface) Position() token.Position { return i.position }
func (i *Iface) Name() string             { return i.fn.Name() }
func (i *Iface) FullName() string {
	if i.iface != "" {
		return fmt.Sprintf("(%v).%v", i.iface, i.Name())
	}
	return i.Name()
}

// Var represents unused package variables
type Var struct {
	o        types.Object
	position token.Position
}

func (v *Var) Position() token.Position { return v.position }
func (v *Var) Name() string             { return v.o.Name() }
func (v *Var) FullName() string         { return v.Name() }

// Param represents unused parameters
type Param struct {
	o        types.Object
	position token.Position
	f        string
}

func (p *Param) Position() token.Position { return p.position }
func (p *Param) Name() string             { return p.o.Name() }
func (p *Param) FullName() string {
	return fmt.Sprintf("%v(%v)", p.f, p.Name())
}

// Field represents unused struct field
type Field struct {
	o        types.Object
	position token.Position
	s        string
}

func (f *Field) Position() token.Position { return f.position }
func (f *Field) Name() string             { return f.o.Name() }
func (f *Field) FullName() string {
	return fmt.Sprintf("%v.%v", f.s, f.Name())
}

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

// GroupObjects takes a list of Objects and groups them by their underlying type.
func GroupObjects(objs []Object) map[string][]Object {
	group := map[string][]Object{
		Miscs:      []Object{},
		Functions:  []Object{},
		Variables:  []Object{},
		Parameters: []Object{},
		Fields:     []Object{},
	}
	for _, o := range objs {
		switch o.(type) {
		case *Func:
			group[Functions] = append(group[Functions], o)
		case *Var:
			group[Variables] = append(group[Variables], o)
		case *Param:
			group[Parameters] = append(group[Parameters], o)
		case *Field:
			group[Fields] = append(group[Fields], o)
		default:
			group[Miscs] = append(group[Miscs], o)
		}
	}
	return group
}
