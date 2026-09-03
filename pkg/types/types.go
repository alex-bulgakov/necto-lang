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
	Methods map[string]*FunctionType
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

type ResultType struct {
	OkType  Type
	ErrType Type
}

func (r *ResultType) Name() string {
	return fmt.Sprintf("Result[%s, %s]", r.OkType.Name(), r.ErrType.Name())
}

func (r *ResultType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if et, ok := other.(*EnumType); ok && et.NameStr == "Result" {
		return true
	}
	if rt, ok := other.(*ResultType); ok {
		return r.OkType.Equals(rt.OkType) && r.ErrType.Equals(rt.ErrType)
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

type BoxType struct {
	Inner Type
}

func (b *BoxType) Name() string {
	return fmt.Sprintf("Box[%s]", b.Inner.Name())
}

func (b *BoxType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if bt, ok := other.(*BoxType); ok {
		return b.Inner.Equals(bt.Inner)
	}
	return false
}

type EnumType struct {
	NameStr  string
	Variants map[string][]Type
}

func (e *EnumType) Name() string { return e.NameStr }
func (e *EnumType) Equals(other Type) bool {
	if other == Any {
		return true
	}
	if _, ok := other.(*ResultType); ok && e.NameStr == "Result" {
		return true
	}
	if et, ok := other.(*EnumType); ok {
		return e.NameStr == et.NameStr
	}
	return false
}

func ParseType(name string, structRegistry map[string]*StructType, enumRegistry ...map[string]*EnumType) Type {
	var enums map[string]*EnumType
	if len(enumRegistry) > 0 {
		enums = enumRegistry[0]
	}

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
		// Проверка Box[T]
		if len(name) > 5 && name[:4] == "Box[" && name[len(name)-1] == ']' {
			innerName := name[4 : len(name)-1]
			return &BoxType{Inner: ParseType(innerName, structRegistry, enums)}
		}
		// Проверка Map[K, V]
		if len(name) > 5 && name[:4] == "Map[" && name[len(name)-1] == ']' {
			inner := name[4 : len(name)-1]
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				k := ParseType(strings.TrimSpace(parts[0]), structRegistry, enums)
				v := ParseType(strings.TrimSpace(parts[1]), structRegistry, enums)
				return &MapType{Key: k, Value: v}
			}
		}
		// Проверка Option[T]
		if len(name) > 7 && name[:7] == "Option[" && name[len(name)-1] == ']' {
			innerName := name[7 : len(name)-1]
			return &OptionType{Inner: ParseType(innerName, structRegistry, enums)}
		}
		// Проверка Result[T, E]
		if len(name) > 8 && name[:7] == "Result[" && name[len(name)-1] == ']' {
			inner := name[7 : len(name)-1]
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				okT := ParseType(strings.TrimSpace(parts[0]), structRegistry, enums)
				errT := ParseType(strings.TrimSpace(parts[1]), structRegistry, enums)
				return &ResultType{OkType: okT, ErrType: errT}
			}
		}
		// Проверка [T]
		if len(name) > 2 && name[0] == '[' && name[len(name)-1] == ']' {
			innerName := name[1 : len(name)-1]
			return &ArrayType{Element: ParseType(innerName, structRegistry, enums)}
		}
		// Одиночные заглавные буквы (e.g. T, U, E, K, V) как generic типы
		if len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z' {
			return Any
		}
		if st, ok := structRegistry[name]; ok {
			return st
		}
		if enums != nil {
			if et, ok := enums[name]; ok {
				return et
			}
		}
		return BasicType(name)
	}
}
