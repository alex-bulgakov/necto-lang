package types

import (
	"fmt"

	"necto/pkg/ast"
	"necto/pkg/token"
)

type Symbol struct {
	Name  string
	Type  Type
	IsMut bool
	Pos   token.Pos
}

type Scope struct {
	parent  *Scope
	symbols map[string]Symbol
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]Symbol),
	}
}

func (s *Scope) Insert(sym Symbol) bool {
	if _, ok := s.symbols[sym.Name]; ok {
		return false // уже объявлено в текущем скоупе
	}
	s.symbols[sym.Name] = sym
	return true
}

func (s *Scope) Lookup(name string) (Symbol, bool) {
	if sym, ok := s.symbols[name]; ok {
		return sym, true
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return Symbol{}, false
}

type Checker struct {
	errors         []string
	scope          *Scope
	structRegistry map[string]*StructType
	enumRegistry   map[string]*EnumType
	currentFnRet   Type
}

func NewChecker() *Checker {
	rootScope := NewScope(nil)

	// Регистрируем встроенные функции
	rootScope.Insert(Symbol{
		Name:  "println",
		Type:  &FunctionType{Params: []Type{Any}, ReturnType: Void},
		IsMut: false,
	})
	rootScope.Insert(Symbol{
		Name:  "print",
		Type:  &FunctionType{Params: []Type{Any}, ReturnType: Void},
		IsMut: false,
	})
	rootScope.Insert(Symbol{
		Name:  "assert",
		Type:  &FunctionType{Params: []Type{Bool}, ReturnType: Void},
		IsMut: false,
	})

	enumRegistry := make(map[string]*EnumType)
	enumRegistry["Result"] = &EnumType{
		NameStr: "Result",
		Variants: map[string][]Type{
			"Ok":  []Type{Any},
			"Err": []Type{Any},
		},
	}

	return &Checker{
		errors:         []string{},
		scope:          rootScope,
		structRegistry: make(map[string]*StructType),
		enumRegistry:   enumRegistry,
	}
}

func (c *Checker) Errors() []string {
	return c.errors
}

func (c *Checker) errorf(pos token.Pos, format string, args ...interface{}) {
	msg := fmt.Sprintf("[%d:%d] Type Error: %s", pos.Line, pos.Col, fmt.Sprintf(format, args...))
	c.errors = append(c.errors, msg)
}

func (c *Checker) Check(program *ast.Program) {
	// Первый проход: сбор объявлений структур, enums, impl и сигнатур функций
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.EnumDeclaration:
			variants := make(map[string][]Type)
			for _, v := range s.Variants {
				var pTypes []Type
				for _, tName := range v.Types {
					pTypes = append(pTypes, ParseType(tName, c.structRegistry, c.enumRegistry))
				}
				variants[v.Name.Value] = pTypes
			}
			et := &EnumType{NameStr: s.Name.Value, Variants: variants}
			c.enumRegistry[s.Name.Value] = et

		case *ast.StructDeclaration:
			fields := make(map[string]Type)
			for _, f := range s.Fields {
				fields[f.Name.Value] = ParseType(f.Type, c.structRegistry, c.enumRegistry)
			}
			st := &StructType{NameStr: s.Name.Value, Fields: fields, Methods: make(map[string]*FunctionType)}
			c.structRegistry[s.Name.Value] = st

		case *ast.ImplBlockStatement:
			st, ok := c.structRegistry[s.Target.Value]
			if !ok {
				c.errorf(s.Pos(), "cannot implement methods for undefined struct '%s'", s.Target.Value)
				continue
			}
			if st.Methods == nil {
				st.Methods = make(map[string]*FunctionType)
			}
			for _, m := range s.Methods {
				params := make([]Type, len(m.Parameters))
				for i, p := range m.Parameters {
					if p.Name.Value == "self" {
						params[i] = st
					} else {
						params[i] = ParseType(p.Type, c.structRegistry, c.enumRegistry)
					}
				}
				retType := ParseType(m.ReturnType, c.structRegistry, c.enumRegistry)
				fnType := &FunctionType{Params: params, ReturnType: retType}
				st.Methods[m.Name.Value] = fnType
			}

		case *ast.FnDeclaration:
			params := make([]Type, len(s.Parameters))
			for i, p := range s.Parameters {
				params[i] = ParseType(p.Type, c.structRegistry, c.enumRegistry)
			}
			retType := ParseType(s.ReturnType, c.structRegistry, c.enumRegistry)
			fnType := &FunctionType{Params: params, ReturnType: retType}
			c.scope.Insert(Symbol{
				Name:  s.Name.Value,
				Type:  fnType,
				IsMut: false,
				Pos:   s.Pos(),
			})

		case *ast.ExternBlockStatement:
			for _, fn := range s.Functions {
				params := make([]Type, len(fn.Parameters))
				for i, p := range fn.Parameters {
					params[i] = ParseType(p.Type, c.structRegistry, c.enumRegistry)
				}
				retType := ParseType(fn.ReturnType, c.structRegistry, c.enumRegistry)
				fnType := &FunctionType{Params: params, ReturnType: retType}
				c.scope.Insert(Symbol{
					Name:  fn.Name.Value,
					Type:  fnType,
					IsMut: false,
					Pos:   fn.Pos(),
				})
			}
		}
	}

	// Второй проход: проверка тел операторов и выражений
	for _, stmt := range program.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		c.checkLetStatement(s)
	case *ast.ReturnStatement:
		c.checkReturnStatement(s)
	case *ast.ExpressionStatement:
		c.checkExpression(s.Expression)
	case *ast.BlockStatement:
		c.checkBlockStatement(s)
	case *ast.WhileStatement:
		c.checkWhileStatement(s)
	case *ast.ForInStatement:
		c.checkForInStatement(s)
	case *ast.FnDeclaration:
		c.checkFnBody(s)
	case *ast.TestBlockStatement:
		c.checkBlockStatement(s.Body)
	case *ast.BenchmarkBlockStatement:
		c.checkBlockStatement(s.Body)
	case *ast.AssertStatement:
		condT := c.checkExpression(s.Condition)
		if !condT.Equals(Bool) && !condT.Equals(Any) {
			c.errorf(s.Pos(), "assert condition must be bool, got '%s'", condT.Name())
		}
	case *ast.ImportStatement:
		for _, sym := range s.Symbols {
			c.scope.Insert(Symbol{
				Name:  sym.Value,
				Type:  Any,
				IsMut: false,
				Pos:   sym.Pos(),
			})
		}
	case *ast.ExternBlockStatement:
		// прототипы внешних функций проверены в первом проходе
	case *ast.ImplBlockStatement:
		st := c.structRegistry[s.Target.Value]
		for _, m := range s.Methods {
			c.checkMethodBody(m, st)
		}
	case *ast.EnumDeclaration, *ast.StructDeclaration, *ast.BreakStatement, *ast.ContinueStatement:
		// уже зарегистрировано или тривиально
	default:
		// ignore
	}
}

func (c *Checker) checkLetStatement(s *ast.LetStatement) {
	var valType Type = Any
	if s.Value != nil {
		valType = c.checkExpression(s.Value)
	}

	var declaredType Type = Any
	if s.TypeAnnotation != "" {
		declaredType = ParseType(s.TypeAnnotation, c.structRegistry, c.enumRegistry)
		if s.Value != nil && !declaredType.Equals(valType) {
			c.errorf(s.Pos(), "cannot assign value of type '%s' to variable '%s' of type '%s'",
				valType.Name(), s.Name.Value, declaredType.Name())
		}
	} else {
		declaredType = valType
	}

	ok := c.scope.Insert(Symbol{
		Name:  s.Name.Value,
		Type:  declaredType,
		IsMut: s.IsMut,
		Pos:   s.Pos(),
	})
	if !ok {
		c.errorf(s.Pos(), "variable '%s' is already declared in this scope", s.Name.Value)
	}
}

func (c *Checker) checkReturnStatement(s *ast.ReturnStatement) {
	var actualRet Type = Void
	if s.ReturnValue != nil {
		actualRet = c.checkExpression(s.ReturnValue)
	}

	if c.currentFnRet != nil && !c.currentFnRet.Equals(actualRet) {
		c.errorf(s.Pos(), "function expects return type '%s', got '%s'",
			c.currentFnRet.Name(), actualRet.Name())
	}
}

func (c *Checker) checkBlockStatement(s *ast.BlockStatement) {
	prevScope := c.scope
	c.scope = NewScope(prevScope)
	defer func() { c.scope = prevScope }()

	for _, stmt := range s.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkWhileStatement(s *ast.WhileStatement) {
	condType := c.checkExpression(s.Condition)
	if !condType.Equals(Bool) && !condType.Equals(Any) {
		c.errorf(s.Condition.Pos(), "while condition must be bool, got '%s'", condType.Name())
	}
	c.checkBlockStatement(s.Body)
}

func (c *Checker) checkForInStatement(s *ast.ForInStatement) {
	iterType := c.checkExpression(s.Iterable)

	prevScope := c.scope
	c.scope = NewScope(prevScope)
	defer func() { c.scope = prevScope }()

	var elemType Type = Int
	if arrType, ok := iterType.(*ArrayType); ok {
		elemType = arrType.Element
	}

	c.scope.Insert(Symbol{
		Name:  s.Item.Value,
		Type:  elemType,
		IsMut: false,
		Pos:   s.Item.Pos(),
	})

	for _, stmt := range s.Body.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkFnBody(s *ast.FnDeclaration) {
	retType := ParseType(s.ReturnType, c.structRegistry, c.enumRegistry)
	prevRet := c.currentFnRet
	c.currentFnRet = retType
	defer func() { c.currentFnRet = prevRet }()

	prevScope := c.scope
	c.scope = NewScope(prevScope)
	defer func() { c.scope = prevScope }()

	for _, param := range s.Parameters {
		pType := ParseType(param.Type, c.structRegistry, c.enumRegistry)
		c.scope.Insert(Symbol{
			Name:  param.Name.Value,
			Type:  pType,
			IsMut: false,
			Pos:   param.Name.Pos(),
		})
	}

	for _, stmt := range s.Body.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkMethodBody(s *ast.FnDeclaration, st *StructType) {
	retType := ParseType(s.ReturnType, c.structRegistry, c.enumRegistry)
	prevRet := c.currentFnRet
	c.currentFnRet = retType
	defer func() { c.currentFnRet = prevRet }()

	prevScope := c.scope
	c.scope = NewScope(prevScope)
	defer func() { c.scope = prevScope }()

	for _, param := range s.Parameters {
		var pType Type
		isMut := false
		if param.Name.Value == "self" {
			if st != nil {
				pType = st
			} else {
				pType = Any
			}
			isMut = true
		} else {
			pType = ParseType(param.Type, c.structRegistry, c.enumRegistry)
		}
		c.scope.Insert(Symbol{
			Name:  param.Name.Value,
			Type:  pType,
			IsMut: isMut,
			Pos:   param.Name.Pos(),
		})
	}

	for _, stmt := range s.Body.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkExpression(expr ast.Expression) Type {
	if expr == nil {
		return Void
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return Int
	case *ast.FloatLiteral:
		return Float
	case *ast.BooleanLiteral:
		return Bool
	case *ast.StringLiteral:
		return Str
	case *ast.FStringLiteral:
		for _, part := range e.Parts {
			c.checkExpression(part)
		}
		return Str
	case *ast.NoneLiteral:
		return &OptionType{Inner: Any}
	case *ast.SomeExpression:
		valT := c.checkExpression(e.Value)
		return &OptionType{Inner: valT}

	case *ast.Identifier:
		if e.Value == "fs" || e.Value == "os" || e.Value == "Map" || e.Value == "Box" || e.Value == "_" {
			return Any
		}
		if _, isEnum := c.enumRegistry[e.Value]; isEnum {
			return Any
		}
		sym, found := c.scope.Lookup(e.Value)
		if !found {
			c.errorf(e.Pos(), "undefined identifier '%s'", e.Value)
			return Any
		}
		return sym.Type

	case *ast.PrefixExpression:
		rightT := c.checkExpression(e.Right)
		if e.Operator == "!" {
			if !rightT.Equals(Bool) && !rightT.Equals(Any) {
				c.errorf(e.Pos(), "operator '!' cannot be applied to type '%s'", rightT.Name())
			}
			return Bool
		} else if e.Operator == "-" {
			if !rightT.Equals(Int) && !rightT.Equals(Float) && !rightT.Equals(Any) {
				c.errorf(e.Pos(), "operator '-' cannot be applied to type '%s'", rightT.Name())
			}
			return rightT
		}
		return rightT

	case *ast.InfixExpression:
		leftT := c.checkExpression(e.Left)
		rightT := c.checkExpression(e.Right)

		switch e.Operator {
		case "+", "-", "*", "/", "%":
			if e.Operator == "+" && (leftT.Equals(Str) || rightT.Equals(Str)) {
				return Str
			}
			if !leftT.Equals(rightT) && !leftT.Equals(Any) && !rightT.Equals(Any) {
				c.errorf(e.Pos(), "mismatched types in operator '%s': '%s' and '%s'",
					e.Operator, leftT.Name(), rightT.Name())
			}
			return leftT
		case "==", "!=":
			return Bool
		case "<", "<=", ">", ">=":
			return Bool
		case "&&", "||":
			if (!leftT.Equals(Bool) && !leftT.Equals(Any)) || (!rightT.Equals(Bool) && !rightT.Equals(Any)) {
				c.errorf(e.Pos(), "logical operator '%s' requires bool operands", e.Operator)
			}
			return Bool
		default:
			return Any
		}

	case *ast.AssignExpression:
		leftT := c.checkExpression(e.Left)
		rightT := c.checkExpression(e.Value)

		// Проверка мутабельности переменной
		if ident, ok := e.Left.(*ast.Identifier); ok {
			sym, found := c.scope.Lookup(ident.Value)
			if found && !sym.IsMut {
				c.errorf(e.Pos(), "cannot assign twice to immutable variable '%s' (declare it with 'let mut')", ident.Value)
			}
		}

		if !leftT.Equals(rightT) && !leftT.Equals(Any) && !rightT.Equals(Any) {
			c.errorf(e.Pos(), "cannot assign '%s' to '%s'", rightT.Name(), leftT.Name())
		}
		return leftT

	case *ast.RangeExpression:
		startT := c.checkExpression(e.Start)
		endT := c.checkExpression(e.End)
		if !startT.Equals(Int) || !endT.Equals(Int) {
			c.errorf(e.Pos(), "range bounds must be integers")
		}
		return &ArrayType{Element: Int}

	case *ast.ArrayLiteral:
		var elemT Type = Any
		if len(e.Elements) > 0 {
			elemT = c.checkExpression(e.Elements[0])
			for i := 1; i < len(e.Elements); i++ {
				t := c.checkExpression(e.Elements[i])
				if !elemT.Equals(t) {
					c.errorf(e.Elements[i].Pos(), "array elements must have same type, expected '%s', got '%s'",
						elemT.Name(), t.Name())
				}
			}
		}
		return &ArrayType{Element: elemT}

	case *ast.IndexExpression:
		leftT := c.checkExpression(e.Left)
		idxT := c.checkExpression(e.Index)
		if !idxT.Equals(Int) && !idxT.Equals(Any) {
			c.errorf(e.Index.Pos(), "index must be int, got '%s'", idxT.Name())
		}
		if arrT, ok := leftT.(*ArrayType); ok {
			return arrT.Element
		}
		return Any

	case *ast.DotExpression:
		// Проверяем вызовы модулей: fs, os, Map
		if ident, ok := e.Left.(*ast.Identifier); ok {
			switch ident.Value {
			case "fs":
				switch e.Right.Value {
				case "read_file":
					return &FunctionType{Params: []Type{Str}, ReturnType: &OptionType{Inner: Str}}
				case "write_file":
					return &FunctionType{Params: []Type{Str, Str}, ReturnType: Bool}
				}
			case "os":
				switch e.Right.Value {
				case "args":
					return &FunctionType{Params: []Type{}, ReturnType: &ArrayType{Element: Str}}
				}
			case "http":
				resType := &EnumType{NameStr: "Result", Variants: map[string][]Type{"Ok": []Type{Str}, "Err": []Type{Str}}}
				switch e.Right.Value {
				case "get":
					return &FunctionType{Params: []Type{Str}, ReturnType: resType}
				case "post":
					return &FunctionType{Params: []Type{Str, Str}, ReturnType: resType}
				case "listen":
					return &FunctionType{Params: []Type{Int, Any}, ReturnType: Bool}
				}
			case "net":
				resType := &EnumType{NameStr: "Result", Variants: map[string][]Type{"Ok": []Type{Str}, "Err": []Type{Str}}}
				switch e.Right.Value {
				case "tcp_connect":
					return &FunctionType{Params: []Type{Str}, ReturnType: resType}
				}
			case "Map":
				switch e.Right.Value {
				case "new":
					return &FunctionType{Params: []Type{}, ReturnType: &MapType{Key: Any, Value: Any}}
				}
			case "Box":
				if e.Right.Value == "new" {
					return &FunctionType{Params: []Type{Any}, ReturnType: &BoxType{Inner: Any}}
				}
			}

			// Статические методы структуры: StructName.new(...)
			if st, exists := c.structRegistry[ident.Value]; exists {
				if mType, ok := st.Methods[e.Right.Value]; ok {
					return mType
				}
			}

			// Проверяем конструкторы вариантов Enum: EnumName.Variant
			if et, exists := c.enumRegistry[ident.Value]; exists {
				if paramTypes, vExists := et.Variants[e.Right.Value]; vExists {
					if len(paramTypes) == 0 {
						return et
					}
					return &FunctionType{Params: paramTypes, ReturnType: et}
				}
				c.errorf(e.Pos(), "variant '%s' does not exist on enum '%s'", e.Right.Value, et.NameStr)
				return et
			}
		}

		leftT := c.checkExpression(e.Left)
		if boxT, ok := leftT.(*BoxType); ok {
			if e.Right.Value == "unwrap" {
				return &FunctionType{Params: []Type{}, ReturnType: boxT.Inner}
			}
		}
		if st, ok := leftT.(*StructType); ok {
			if fType, exists := st.Fields[e.Right.Value]; exists {
				return fType
			}
			if mType, exists := st.Methods[e.Right.Value]; exists {
				if len(mType.Params) > 0 && mType.Params[0].Equals(st) {
					return &FunctionType{Params: mType.Params[1:], ReturnType: mType.ReturnType}
				}
				return mType
			}
			c.errorf(e.Pos(), "field or method '%s' does not exist on struct '%s'", e.Right.Value, st.NameStr)
			return Any
		}

		// Методы массивов
		if arrT, ok := leftT.(*ArrayType); ok {
			switch e.Right.Value {
			case "push":
				return &FunctionType{Params: []Type{arrT.Element}, ReturnType: Void}
			case "pop":
				return &FunctionType{Params: []Type{}, ReturnType: &OptionType{Inner: arrT.Element}}
			case "len":
				return &FunctionType{Params: []Type{}, ReturnType: Int}
			case "clear":
				return &FunctionType{Params: []Type{}, ReturnType: Void}
			}
		}

		// Методы строк
		if leftT.Equals(Str) {
			switch e.Right.Value {
			case "len":
				return &FunctionType{Params: []Type{}, ReturnType: Int}
			case "sub":
				return &FunctionType{Params: []Type{Int, Int}, ReturnType: Str}
			case "char_at":
				return &FunctionType{Params: []Type{Int}, ReturnType: Int}
			case "contains":
				return &FunctionType{Params: []Type{Str}, ReturnType: Bool}
			}
		}

		// Методы словарей
		if mapT, ok := leftT.(*MapType); ok {
			switch e.Right.Value {
			case "set":
				return &FunctionType{Params: []Type{mapT.Key, mapT.Value}, ReturnType: Void}
			case "get":
				return &FunctionType{Params: []Type{mapT.Key}, ReturnType: &OptionType{Inner: mapT.Value}}
			case "has":
				return &FunctionType{Params: []Type{mapT.Key}, ReturnType: Bool}
			case "len":
				return &FunctionType{Params: []Type{}, ReturnType: Int}
			}
		}

		return Any

	case *ast.CallExpression:
		fnT := c.checkExpression(e.Function)
		for _, arg := range e.Arguments {
			c.checkExpression(arg)
		}
		if ft, ok := fnT.(*FunctionType); ok {
			return ft.ReturnType
		}
		return Any

	case *ast.StructLiteral:
		st, ok := c.structRegistry[e.StructName.Value]
		if !ok {
			c.errorf(e.Pos(), "undefined struct '%s'", e.StructName.Value)
			return Any
		}
		for _, field := range e.Fields {
			expectedType, exists := st.Fields[field.Name]
			if !exists {
				c.errorf(e.Pos(), "struct '%s' has no field '%s'", st.NameStr, field.Name)
				continue
			}
			valType := c.checkExpression(field.Value)
			if !expectedType.Equals(valType) && !valType.Equals(Any) {
				c.errorf(field.Value.Pos(), "field '%s' expects type '%s', got '%s'",
					field.Name, expectedType.Name(), valType.Name())
			}
		}
		return st

	case *ast.IfExpression:
		condT := c.checkExpression(e.Condition)
		if !condT.Equals(Bool) && !condT.Equals(Any) {
			c.errorf(e.Condition.Pos(), "if condition must be bool, got '%s'", condT.Name())
		}
		c.checkBlockStatement(e.Consequence)
		if e.Alternative != nil {
			c.checkBlockStatement(e.Alternative)
		}
		return Void

	case *ast.MatchExpression:
		valT := c.checkExpression(e.Value)
		for _, mc := range e.Cases {
			caseScope := NewScope(c.scope)
			prevScope := c.scope
			c.scope = caseScope

			c.bindPatternVariables(mc.Pattern, valT)

			_ = c.checkExpression(mc.Body)
			c.scope = prevScope
		}
		return Any

	case *ast.EnumConstructorExpr:
		if et, exists := c.enumRegistry[e.EnumName]; exists {
			for _, arg := range e.Arguments {
				c.checkExpression(arg)
			}
			return et
		}
		return Any

	case *ast.TryExpression:
		rightT := c.checkExpression(e.Right)
		if rt, ok := rightT.(*ResultType); ok {
			return rt.OkType
		}
		if opt, ok := rightT.(*OptionType); ok {
			return opt.Inner
		}
		return Any
	}

	return Any
}

func (c *Checker) bindPatternVariables(pat ast.Expression, targetType Type) {
	switch p := pat.(type) {
	case *ast.Identifier:
		if p.Value != "_" {
			c.scope.Insert(Symbol{
				Name:  p.Value,
				Type:  targetType,
				IsMut: false,
				Pos:   p.Pos(),
			})
		}
	case *ast.SomeExpression:
		var innerType Type = Any
		if optType, ok := targetType.(*OptionType); ok {
			innerType = optType.Inner
		}
		c.bindPatternVariables(p.Value, innerType)

	case *ast.CallExpression:
		if dot, ok := p.Function.(*ast.DotExpression); ok {
			if ident, ok := dot.Left.(*ast.Identifier); ok {
				if et, ok := c.enumRegistry[ident.Value]; ok {
					if paramTypes, ok := et.Variants[dot.Right.Value]; ok {
						for i, arg := range p.Arguments {
							var pT Type = Any
							if i < len(paramTypes) {
								pT = paramTypes[i]
							}
							c.bindPatternVariables(arg, pT)
						}
					}
				}
			}
		}
	case *ast.DotExpression:
		// Вариант без аргументов (e.g. Token.Eof)
	}
}
