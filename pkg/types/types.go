package types

import (
	"fmt"
	"strings"
)

type Type interface {
	Name() string
	Equals(other Type) bool
}

type MapType struct {
	Key   Type
	Value Type
}

func (m *MapType) Name() string {
	return fmt.Sprintf("Map[%s, %s]", m.Key.Name(), m.Value.Name())
}

func (m *MapType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if mt, ok := other.(*MapType); ok {
		return m.Key.Equals(mt.Key) && m.Value.Equals(mt.Value)
	}
	return false
}

type BasicType string

func (b BasicType) Name() string { return string(b) }
func (b BasicType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if b == Any {
		return true
	}
	return b.Name() == other.Name()
}

var (
	Int  = BasicType("int")
	Float = BasicType("float")
	Bool = BasicType("bool")
	Str  = BasicType("str")
	Void = BasicType("void")
	Any  = BasicType("any")
)

type OptionType struct {
	Inner Type
}

func (o *OptionType) Name() string {
	return fmt.Sprintf("Option[%s]", o.Inner.Name())
}

func (o *OptionType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if opt, ok := other.(*OptionType); ok {
		return o.Inner.Equals(opt.Inner)
	}
	return false
}

type ArrayType struct {
	Element Type
}

func (a *ArrayType) Name() string {
	return fmt.Sprintf("[%s]", a.Element.Name())
}

func (a *ArrayType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if arr, ok := other.(*ArrayType); ok {
		return a.Element.Equals(arr.Element)
	}
	return false
}

type StructField struct {
	Name string
	Type Type
}

type StructType struct {
	NameStr string
	Fields  map[string]Type
}

func (s *StructType) Name() string { return s.NameStr }
func (s *StructType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if st, ok := other.(*StructType); ok {
		return s.NameStr == st.NameStr
	}
	return false
}

type FunctionType struct {
	Params     []Type
	ReturnType Type
}

func (f *FunctionType) Name() string {
	return "fn(...)"
}

func (f *FunctionType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if fn, ok := other.(*FunctionType); ok {
		if len(f.Params) != len(fn.Params) {
			return false
		}
		for i := range f.Params {
			if !f.Params[i].Equals(fn.Params[i]) {
				return false
			}
		}
		return f.ReturnType.Equals(fn.ReturnType)
	}
	return false
}

func ParseType(name string, structRegistry map[string]*StructType) Type {
	switch name {
	case "int":
		return Int
	case "float":
		return Float
	case "bool":
		return Bool
	case "str", "string":
		return Str
	case "void", "":
		return Void
	case "any":
		return Any
	default:
		// Проверка Map[K, V]
		if len(name) > 5 && name[:4] == "Map[" && name[len(name)-1] == ']' {
			inner := name[4 : len(name)-1]
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				k := ParseType(strings.TrimSpace(parts[0]), structRegistry)
				v := ParseType(strings.TrimSpace(parts[1]), structRegistry)
				return &MapType{Key: k, Value: v}
			}
		}
		// Проверка Option[T]
		if len(name) > 7 && name[:7] == "Option[" && name[len(name)-1] == ']' {
			innerName := name[7 : len(name)-1]
			return &OptionType{Inner: ParseType(innerName, structRegistry)}
		}
		// Проверка [T]
		if len(name) > 2 && name[0] == '[' && name[len(name)-1] == ']' {
			innerName := name[1 : len(name)-1]
			return &ArrayType{Element: ParseType(innerName, structRegistry)}
		}
		if st, ok := structRegistry[name]; ok {
			return st
		}
		return BasicType(name)
	}
}
