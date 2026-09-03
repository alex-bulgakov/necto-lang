package eval

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"necto/pkg/ast"
)

type ObjectType string

const (
	INTEGER_OBJ     = "INTEGER"
	FLOAT_OBJ       = "FLOAT"
	BOOLEAN_OBJ     = "BOOLEAN"
	STRING_OBJ      = "STRING"
	NULL_OBJ        = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	BREAK_OBJ       = "BREAK"
	CONTINUE_OBJ    = "CONTINUE"
	ERROR_OBJ       = "ERROR"
	FUNCTION_OBJ    = "FUNCTION"
	BUILTIN_OBJ     = "BUILTIN"
	ARRAY_OBJ       = "ARRAY"
	OPTION_OBJ      = "OPTION"
	STRUCT_OBJ      = "STRUCT"
	MAP_OBJ         = "MAP"
	MODULE_OBJ      = "MODULE"
	BOX_OBJ         = "BOX"
	ENUM_OBJ        = "ENUM"
	RESULT_OBJ      = "RESULT"
	STRUCT_DEF_OBJ  = "STRUCT_DEF"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return strconv.FormatInt(i.Value, 10) }

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return strconv.FormatFloat(f.Value, 'f', -1, 64) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return strconv.FormatBool(b.Value) }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Break struct{}

func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) Inspect() string  { return "break" }

type Continue struct{}

func (c *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue) Inspect() string  { return "continue" }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "Runtime Error: " + e.Message }

type Function struct {
	Parameters []*ast.Parameter
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := make([]string, len(f.Parameters))
	for i, p := range f.Parameters {
		params[i] = p.Name.Value
	}
	out.WriteString("fn(" + strings.Join(params, ", ") + ") { ... }")
	return out.String()
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	elems := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elems[i] = e.Inspect()
	}
	out.WriteString("[" + strings.Join(elems, ", ") + "]")
	return out.String()
}

type Option struct {
	HasValue bool
	Value    Object
}

func (o *Option) Type() ObjectType { return OPTION_OBJ }
func (o *Option) Inspect() string {
	if o.HasValue {
		return fmt.Sprintf("Some(%s)", o.Value.Inspect())
	}
	return "None"
}

type StructInstance struct {
	Name   string
	Fields map[string]Object
}

func (s *StructInstance) Type() ObjectType { return STRUCT_OBJ }
func (s *StructInstance) Inspect() string {
	var out bytes.Buffer
	out.WriteString(s.Name + " { ")
	var parts []string
	for k, v := range s.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v.Inspect()))
	}
	out.WriteString(strings.Join(parts, ", "))
	out.WriteString(" }")
	return out.String()
}

type MapInstance struct {
	Store map[string]Object
}

func (m *MapInstance) Type() ObjectType { return MAP_OBJ }
func (m *MapInstance) Inspect() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	var parts []string
	for k, v := range m.Store {
		parts = append(parts, fmt.Sprintf("%q: %s", k, v.Inspect()))
	}
	out.WriteString(strings.Join(parts, ", "))
	out.WriteString(" }")
	return out.String()
}

type Module struct {
	Name    string
	Methods map[string]*Builtin
}

func (m *Module) Type() ObjectType { return MODULE_OBJ }
func (m *Module) Inspect() string  { return "<module " + m.Name + ">" }

type BoxInstance struct {
	Value Object
}

func (b *BoxInstance) Type() ObjectType { return BOX_OBJ }
func (b *BoxInstance) Inspect() string  { return fmt.Sprintf("Box(%s)", b.Value.Inspect()) }

type EnumInstance struct {
	EnumName string
	Variant  string
	Fields   []Object
}

func (e *EnumInstance) Type() ObjectType { return ENUM_OBJ }
func (e *EnumInstance) Inspect() string {
	var out bytes.Buffer
	out.WriteString(e.EnumName + "." + e.Variant)
	if len(e.Fields) > 0 {
		var parts []string
		for _, f := range e.Fields {
			parts = append(parts, f.Inspect())
		}
		out.WriteString("(" + strings.Join(parts, ", ") + ")")
	}
	return out.String()
}

type ResultInstance struct {
	IsErr bool
	Value Object
}

func (r *ResultInstance) Type() ObjectType { return RESULT_OBJ }
func (r *ResultInstance) Inspect() string {
	if r.IsErr {
		return fmt.Sprintf("Result.Err(%s)", r.Value.Inspect())
	}
	return fmt.Sprintf("Result.Ok(%s)", r.Value.Inspect())
}

type StructDefinition struct {
	Name    string
	Methods map[string]*Function
}

func (s *StructDefinition) Type() ObjectType { return STRUCT_DEF_OBJ }
func (s *StructDefinition) Inspect() string  { return "struct " + s.Name }
