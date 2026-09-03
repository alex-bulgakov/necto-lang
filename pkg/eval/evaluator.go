package eval

import (
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"necto/pkg/ast"
	"necto/pkg/lexer"
	"necto/pkg/parser"
)

var (
	NULL  = &Null{}
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
	NONE  = &Option{HasValue: false}
)

var modules = map[string]*Module{
	"fs": {
		Name: "fs",
		Methods: map[string]*Builtin{
			"read_file": {
				Fn: func(args ...Object) Object {
					if len(args) != 1 {
						return newError("fs.read_file() takes 1 argument (path)")
					}
					pathStr := args[0].Inspect()
					data, err := os.ReadFile(pathStr)
					if err != nil {
						return NONE
					}
					return &Option{HasValue: true, Value: &String{Value: string(data)}}
				},
			},
			"write_file": {
				Fn: func(args ...Object) Object {
					if len(args) != 2 {
						return newError("fs.write_file() takes 2 arguments (path, content)")
					}
					pathStr := args[0].Inspect()
					contentStr := args[1].Inspect()
					err := os.WriteFile(pathStr, []byte(contentStr), 0644)
					return nativeBoolToBooleanObject(err == nil)
				},
			},
		},
	},
	"os": {
		Name: "os",
		Methods: map[string]*Builtin{
			"args": {
				Fn: func(args ...Object) Object {
					var elems []Object
					for _, a := range os.Args {
						elems = append(elems, &String{Value: a})
					}
					return &Array{Elements: elems}
				},
			},
		},
	},
	"Map": {
		Name: "Map",
		Methods: map[string]*Builtin{
			"new": {
				Fn: func(args ...Object) Object {
					return &MapInstance{Store: make(map[string]Object)}
				},
			},
		},
	},
	"Box": {
		Name: "Box",
		Methods: map[string]*Builtin{
			"new": {
				Fn: func(args ...Object) Object {
					if len(args) != 1 {
						return newError("Box.new() takes exactly 1 argument")
					}
					return &BoxInstance{Value: args[0]}
				},
			},
		},
	},
	"Result": {
		Name: "Result",
		Methods: map[string]*Builtin{
			"Ok": {
				Fn: func(args ...Object) Object {
					var val Object = NULL
					if len(args) > 0 {
						val = args[0]
					}
					return &ResultInstance{IsErr: false, Value: val}
				},
			},
			"Err": {
				Fn: func(args ...Object) Object {
					var val Object = NULL
					if len(args) > 0 {
						val = args[0]
					}
					return &ResultInstance{IsErr: true, Value: val}
				},
			},
		},
	},
}

func init() {
	modules["http"] = &Module{
		Name: "http",
		Methods: map[string]*Builtin{
			"get": {
				Fn: func(args ...Object) Object {
					if len(args) != 1 {
						return newError("http.get() takes 1 argument (url)")
					}
					urlStr := args[0].Inspect()
					client := &http.Client{Timeout: 10 * time.Second}
					resp, err := client.Get(urlStr)
					if err != nil {
						return &ResultInstance{IsErr: true, Value: &String{Value: err.Error()}}
					}
					defer resp.Body.Close()
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						return &ResultInstance{IsErr: true, Value: &String{Value: err.Error()}}
					}
					return &ResultInstance{IsErr: false, Value: &String{Value: string(bodyBytes)}}
				},
			},
			"post": {
				Fn: func(args ...Object) Object {
					if len(args) != 2 {
						return newError("http.post() takes 2 arguments (url, body)")
					}
					urlStr := args[0].Inspect()
					bodyStr := args[1].Inspect()
					client := &http.Client{Timeout: 10 * time.Second}
					resp, err := client.Post(urlStr, "application/json", strings.NewReader(bodyStr))
					if err != nil {
						return &ResultInstance{IsErr: true, Value: &String{Value: err.Error()}}
					}
					defer resp.Body.Close()
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						return &ResultInstance{IsErr: true, Value: &String{Value: err.Error()}}
					}
					return &ResultInstance{IsErr: false, Value: &String{Value: string(bodyBytes)}}
				},
			},
			"listen": {
				Fn: func(args ...Object) Object {
					if len(args) < 2 {
						return newError("http.listen() takes 2 arguments (port, handler)")
					}
					portInt, ok := args[0].(*Integer)
					if !ok {
						return newError("port must be an integer")
					}

					mux := http.NewServeMux()
					mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						reqObj := &String{Value: r.URL.Path}
						res := applyFunction(args[1], []Object{reqObj})
						if res != nil {
							if res.Type() == RETURN_VALUE_OBJ {
								res = res.(*ReturnValue).Value
							}
							if s, ok := res.(*String); ok {
								w.Write([]byte(s.Value))
							} else {
								w.Write([]byte(res.Inspect()))
							}
						}
					})

					server := &http.Server{
						Addr:    fmt.Sprintf(":%d", portInt.Value),
						Handler: mux,
					}
					go server.ListenAndServe()
					time.Sleep(10 * time.Millisecond)
					return TRUE
				},
			},
		},
	}

	modules["net"] = &Module{
		Name: "net",
		Methods: map[string]*Builtin{
			"tcp_connect": {
				Fn: func(args ...Object) Object {
					if len(args) != 1 {
						return newError("net.tcp_connect() takes 1 argument (address)")
					}
					addr := args[0].Inspect()
					conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
					if err != nil {
						return &ResultInstance{IsErr: true, Value: &String{Value: err.Error()}}
					}
					conn.Close()
					return &ResultInstance{IsErr: false, Value: &String{Value: "connected"}}
				},
			},
		},
	}
}

var builtins = map[string]*Builtin{
	"Ok": {
		Fn: func(args ...Object) Object {
			var val Object = NULL
			if len(args) > 0 {
				val = args[0]
			}
			return &ResultInstance{IsErr: false, Value: val}
		},
	},
	"Err": {
		Fn: func(args ...Object) Object {
			var val Object = NULL
			if len(args) > 0 {
				val = args[0]
			}
			return &ResultInstance{IsErr: true, Value: val}
		},
	},
	"assert": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("assert() takes at least 1 condition argument")
			}
			if !isTruthy(args[0]) {
				msg := "assertion failed"
				if len(args) > 1 {
					msg = args[1].Inspect()
				}
				return newError("AssertionError: " + msg)
			}
			return NULL
		},
	},
	"println": {
		Fn: func(args ...Object) Object {
			var parts []string
			for _, arg := range args {
				parts = append(parts, arg.Inspect())
			}
			fmt.Println(strings.Join(parts, " "))
			return NULL
		},
	},
	"print": {
		Fn: func(args ...Object) Object {
			var parts []string
			for _, arg := range args {
				parts = append(parts, arg.Inspect())
			}
			fmt.Print(strings.Join(parts, " "))
			return NULL
		},
	},
	"len": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("len() takes exactly 1 argument (%d given)", len(args))
			}
			switch arg := args[0].(type) {
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to len() not supported, got %s", arg.Type())
			}
		},
	},
	"clock": {
		Fn: func(args ...Object) Object {
			return &Float{Value: float64(time.Now().UnixNano()) / 1e9}
		},
	},
}

func Eval(node ast.Node, env *Environment) Object {
	switch n := node.(type) {
	case *ast.Program:
		return evalProgram(n, env)

	case *ast.BlockStatement:
		return evalBlockStatement(n, env)

	case *ast.ExpressionStatement:
		return Eval(n.Expression, env)

	case *ast.ReturnStatement:
		val := Eval(n.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &ReturnValue{Value: val}

	case *ast.BreakStatement:
		return &Break{}

	case *ast.ContinueStatement:
		return &Continue{}

	case *ast.LetStatement:
		var val Object = NULL
		if n.Value != nil {
			val = Eval(n.Value, env)
			if isError(val) || (val != nil && val.Type() == RETURN_VALUE_OBJ) {
				return val
			}
		}
		env.Set(n.Name.Value, val)
		return NULL

	case *ast.FnDeclaration:
		fn := &Function{
			Parameters: n.Parameters,
			Body:       n.Body,
			Env:        env,
		}
		env.Set(n.Name.Value, fn)
		return NULL

	case *ast.StructDeclaration:
		def := &StructDefinition{Name: n.Name.Value, Methods: make(map[string]*Function)}
		env.Set(n.Name.Value, def)
		return NULL

	case *ast.ImplBlockStatement:
		var def *StructDefinition
		if obj, exists := env.Get(n.Target.Value); exists {
			if d, ok := obj.(*StructDefinition); ok {
				def = d
			}
		}
		if def == nil {
			def = &StructDefinition{Name: n.Target.Value, Methods: make(map[string]*Function)}
			env.Set(n.Target.Value, def)
		}
		for _, m := range n.Methods {
			fn := &Function{
				Parameters: m.Parameters,
				Body:       m.Body,
				Env:        env,
			}
			def.Methods[m.Name.Value] = fn
		}
		return NULL

	case *ast.EnumDeclaration:
		env.Set(n.Name.Value, &EnumInstance{EnumName: n.Name.Value, Variant: ""})
		return NULL

	case *ast.ExternBlockStatement:
		for _, fn := range n.Functions {
			fnName := fn.Name.Value
			builtin := &Builtin{
				Fn: func(args ...Object) Object {
					return callExternC(fnName, args)
				},
			}
			env.Set(fnName, builtin)
		}
		return NULL

	case *ast.AssertStatement:
		cond := Eval(n.Condition, env)
		if isError(cond) {
			return cond
		}
		if !isTruthy(cond) {
			return newError("AssertionError: condition failed: %s", n.Condition.String())
		}
		return NULL

	case *ast.TestBlockStatement:
		return evalBlockStatement(n.Body, NewEnclosedEnvironment(env))

	case *ast.ImportStatement:
		return evalImportStatement(n, env)

	case *ast.EnumConstructorExpr:
		var args []Object
		for _, a := range n.Arguments {
			argVal := Eval(a, env)
			if isError(argVal) {
				return argVal
			}
			args = append(args, argVal)
		}
		return &EnumInstance{
			EnumName: n.EnumName,
			Variant:  n.VariantName,
			Fields:   args,
		}

	case *ast.WhileStatement:
		return evalWhileStatement(n, env)

	case *ast.ForInStatement:
		return evalForInStatement(n, env)

	case *ast.IntegerLiteral:
		return &Integer{Value: n.Value}

	case *ast.FloatLiteral:
		return &Float{Value: n.Value}

	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(n.Value)

	case *ast.StringLiteral:
		return &String{Value: n.Value}

	case *ast.FStringLiteral:
		var sb strings.Builder
		for _, part := range n.Parts {
			evaluated := Eval(part, env)
			if isError(evaluated) {
				return evaluated
			}
			sb.WriteString(evaluated.Inspect())
		}
		return &String{Value: sb.String()}

	case *ast.NoneLiteral:
		return NONE

	case *ast.SomeExpression:
		val := Eval(n.Value, env)
		if isError(val) {
			return val
		}
		return &Option{HasValue: true, Value: val}

	case *ast.Identifier:
		return evalIdentifier(n, env)

	case *ast.PrefixExpression:
		right := Eval(n.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(n.Operator, right)

	case *ast.InfixExpression:
		left := Eval(n.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(n.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(n.Operator, left, right)

	case *ast.AssignExpression:
		return evalAssignExpression(n, env)

	case *ast.IfExpression:
		return evalIfExpression(n, env)

	case *ast.CallExpression:
		function := Eval(n.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(n.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)

	case *ast.IndexExpression:
		left := Eval(n.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(n.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.DotExpression:
		if ident, ok := n.Left.(*ast.Identifier); ok {
			if mod, exists := modules[ident.Value]; exists {
				if method, ok := mod.Methods[n.Right.Value]; ok {
					return method
				}
				return newError("module '%s' has no method '%s'", ident.Value, n.Right.Value)
			}
			// Проверяем статические методы структуры: StructName.new(...)
			if obj, exists := env.Get(ident.Value); exists {
				if def, ok := obj.(*StructDefinition); ok {
					if method, ok := def.Methods[n.Right.Value]; ok {
						return method
					}
				}
				if enumInst, ok := obj.(*EnumInstance); ok {
					return &EnumInstance{
						EnumName: enumInst.EnumName,
						Variant:  n.Right.Value,
						Fields:   nil,
					}
				}
			}
			if _, inEnv := env.Get(ident.Value); !inEnv {
				return &EnumInstance{
					EnumName: ident.Value,
					Variant:  n.Right.Value,
					Fields:   nil,
				}
			}
		}

		left := Eval(n.Left, env)
		if isError(left) {
			return left
		}
		return evalDotExpression(left, n.Right.Value, env)

	case *ast.TryExpression:
		res := Eval(n.Right, env)
		if isError(res) {
			return res
		}
		if r, ok := res.(*ResultInstance); ok {
			if r.IsErr {
				return &ReturnValue{Value: r}
			}
			return r.Value
		}
		if opt, ok := res.(*Option); ok {
			if !opt.HasValue {
				return &ReturnValue{Value: NONE}
			}
			return opt.Value
		}
		return res

	case *ast.StructLiteral:
		fields := make(map[string]Object)
		for _, f := range n.Fields {
			fVal := Eval(f.Value, env)
			if isError(fVal) {
				return fVal
			}
			fields[f.Name] = fVal
		}
		return &StructInstance{
			Name:   n.StructName.Value,
			Fields: fields,
		}

	case *ast.ArrayLiteral:
		elements := evalExpressions(n.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &Array{Elements: elements}

	case *ast.RangeExpression:
		start := Eval(n.Start, env)
		end := Eval(n.End, env)
		if sInt, ok := start.(*Integer); ok {
			if eInt, ok := end.(*Integer); ok {
				var elems []Object
				for i := sInt.Value; i < eInt.Value; i++ {
					elems = append(elems, &Integer{Value: i})
				}
				return &Array{Elements: elems}
			}
		}
		return newError("range expression requires integers")

	case *ast.MatchExpression:
		return evalMatchExpression(n, env)
	}

	return NULL
}

func evalProgram(program *ast.Program, env *Environment) Object {
	var result Object = NULL

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch r := result.(type) {
		case *ReturnValue:
			return r.Value
		case *Error:
			return r
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *Environment) Object {
	var result Object = NULL

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ || rt == BREAK_OBJ || rt == CONTINUE_OBJ {
				return result
			}
		}
	}

	return result
}

func evalWhileStatement(ws *ast.WhileStatement, env *Environment) Object {
	for {
		cond := Eval(ws.Condition, env)
		if isError(cond) {
			return cond
		}
		if !isTruthy(cond) {
			break
		}

		res := evalBlockStatement(ws.Body, env)
		if res != nil {
			if res.Type() == BREAK_OBJ {
				break
			}
			if res.Type() == CONTINUE_OBJ {
				continue
			}
			if res.Type() == RETURN_VALUE_OBJ || res.Type() == ERROR_OBJ {
				return res
			}
		}
	}
	return NULL
}

func evalForInStatement(fs *ast.ForInStatement, env *Environment) Object {
	iterObj := Eval(fs.Iterable, env)
	if isError(iterObj) {
		return iterObj
	}

	arr, ok := iterObj.(*Array)
	if !ok {
		return newError("for-in loop requires an array or range, got %s", iterObj.Type())
	}

	loopEnv := NewEnclosedEnvironment(env)

	for _, elem := range arr.Elements {
		loopEnv.Set(fs.Item.Value, elem)
		res := evalBlockStatement(fs.Body, loopEnv)
		if res != nil {
			if res.Type() == BREAK_OBJ {
				break
			}
			if res.Type() == CONTINUE_OBJ {
				continue
			}
			if res.Type() == RETURN_VALUE_OBJ || res.Type() == ERROR_OBJ {
				return res
			}
		}
	}

	return NULL
}

func evalIdentifier(node *ast.Identifier, env *Environment) Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if mod, ok := modules[node.Value]; ok {
		return mod
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
}

func evalPrefixExpression(operator string, right Object) Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right Object) Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		if opt, ok := right.(*Option); ok && !opt.HasValue {
			return TRUE
		}
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right Object) Object {
	if right.Type() == INTEGER_OBJ {
		value := right.(*Integer).Value
		return &Integer{Value: -value}
	}
	if right.Type() == FLOAT_OBJ {
		value := right.(*Float).Value
		return &Float{Value: -value}
	}
	return newError("unknown operator: -%s", right.Type())
}

func evalInfixExpression(operator string, left, right Object) Object {
	if left.Type() == RETURN_VALUE_OBJ {
		left = left.(*ReturnValue).Value
	}
	if right.Type() == RETURN_VALUE_OBJ {
		right = right.(*ReturnValue).Value
	}

	if left.Type() == INTEGER_OBJ && right.Type() == INTEGER_OBJ {
		return evalIntegerInfixExpression(operator, left, right)
	}
	if (left.Type() == FLOAT_OBJ || left.Type() == INTEGER_OBJ) &&
		(right.Type() == FLOAT_OBJ || right.Type() == INTEGER_OBJ) {
		return evalFloatInfixExpression(operator, left, right)
	}
	if left.Type() == STRING_OBJ || right.Type() == STRING_OBJ {
		if operator == "+" {
			return &String{Value: left.Inspect() + right.Inspect()}
		}
	}
	if left.Type() == STRING_OBJ && right.Type() == STRING_OBJ {
		switch operator {
		case "==":
			return nativeBoolToBooleanObject(left.(*String).Value == right.(*String).Value)
		case "!=":
			return nativeBoolToBooleanObject(left.(*String).Value != right.(*String).Value)
		}
	}

	if leftEnum, ok := left.(*EnumInstance); ok {
		if rightEnum, ok := right.(*EnumInstance); ok {
			same := leftEnum.EnumName == rightEnum.EnumName && leftEnum.Variant == rightEnum.Variant && len(leftEnum.Fields) == len(rightEnum.Fields)
			if same {
				for i := range leftEnum.Fields {
					if leftEnum.Fields[i].Inspect() != rightEnum.Fields[i].Inspect() {
						same = false
						break
					}
				}
			}
			if operator == "==" {
				return nativeBoolToBooleanObject(same)
			}
			if operator == "!=" {
				return nativeBoolToBooleanObject(!same)
			}
		}
	}

	if leftOpt, ok := left.(*Option); ok {
		if rightOpt, ok := right.(*Option); ok {
			same := leftOpt.HasValue == rightOpt.HasValue
			if same && leftOpt.HasValue {
				same = leftOpt.Value.Inspect() == rightOpt.Value.Inspect()
			}
			if operator == "==" {
				return nativeBoolToBooleanObject(same)
			}
			if operator == "!=" {
				return nativeBoolToBooleanObject(!same)
			}
		}
	}

	if leftRes, ok := left.(*ResultInstance); ok {
		if rightRes, ok := right.(*ResultInstance); ok {
			same := leftRes.IsErr == rightRes.IsErr && leftRes.Value.Inspect() == rightRes.Value.Inspect()
			if operator == "==" {
				return nativeBoolToBooleanObject(same)
			}
			if operator == "!=" {
				return nativeBoolToBooleanObject(!same)
			}
		}
	}

	switch operator {
	case "==":
		return nativeBoolToBooleanObject(left == right)
	case "!=":
		return nativeBoolToBooleanObject(left != right)
	case "&&":
		return nativeBoolToBooleanObject(isTruthy(left) && isTruthy(right))
	case "||":
		return nativeBoolToBooleanObject(isTruthy(left) || isTruthy(right))
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right Object) Object {
	leftVal := left.(*Integer).Value
	rightVal := right.(*Integer).Value

	switch operator {
	case "+":
		return &Integer{Value: leftVal + rightVal}
	case "-":
		return &Integer{Value: leftVal - rightVal}
	case "*":
		return &Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("division by zero in modulo")
		}
		return &Integer{Value: leftVal % rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(operator string, left, right Object) Object {
	var leftVal, rightVal float64

	if left.Type() == INTEGER_OBJ {
		leftVal = float64(left.(*Integer).Value)
	} else {
		leftVal = left.(*Float).Value
	}

	if right.Type() == INTEGER_OBJ {
		rightVal = float64(right.(*Integer).Value)
	} else {
		rightVal = right.(*Float).Value
	}

	switch operator {
	case "+":
		return &Float{Value: leftVal + rightVal}
	case "-":
		return &Float{Value: leftVal - rightVal}
	case "*":
		return &Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return &Float{Value: math.Inf(1)}
		}
		return &Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalAssignExpression(n *ast.AssignExpression, env *Environment) Object {
	val := Eval(n.Value, env)
	if isError(val) {
		return val
	}

	switch target := n.Left.(type) {
	case *ast.Identifier:
		if n.Operator != "=" {
			cur, ok := env.Get(target.Value)
			if !ok {
				return newError("identifier not found: %s", target.Value)
			}
			op := strings.TrimSuffix(n.Operator, "=")
			val = evalInfixExpression(op, cur, val)
			if isError(val) {
				return val
			}
		}
		if !env.Update(target.Value, val) {
			env.Set(target.Value, val)
		}
		return val

	case *ast.DotExpression:
		leftObj := Eval(target.Left, env)
		if st, ok := leftObj.(*StructInstance); ok {
			if n.Operator != "=" {
				cur, exists := st.Fields[target.Right.Value]
				if !exists {
					return newError("field '%s' not found on struct '%s'", target.Right.Value, st.Name)
				}
				op := strings.TrimSuffix(n.Operator, "=")
				val = evalInfixExpression(op, cur, val)
				if isError(val) {
					return val
				}
			}
			st.Fields[target.Right.Value] = val
			return val
		}
		return newError("cannot assign property on non-struct %s", leftObj.Type())

	case *ast.IndexExpression:
		leftObj := Eval(target.Left, env)
		idxObj := Eval(target.Index, env)
		if m, ok := leftObj.(*MapInstance); ok {
			key := idxObj.Inspect()
			m.Store[key] = val
			return val
		}
		if arr, ok := leftObj.(*Array); ok {
			if idx, ok := idxObj.(*Integer); ok {
				if idx.Value < 0 || idx.Value >= int64(len(arr.Elements)) {
					return newError("index out of range: %d", idx.Value)
				}
				if n.Operator != "=" {
					cur := arr.Elements[idx.Value]
					op := strings.TrimSuffix(n.Operator, "=")
					val = evalInfixExpression(op, cur, val)
					if isError(val) {
						return val
					}
				}
				arr.Elements[idx.Value] = val
				return val
			}
		}
		return newError("cannot index assignment on %s", leftObj.Type())
	}

	return newError("invalid assignment target")
}

func evalIfExpression(ie *ast.IfExpression, env *Environment) Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return evalBlockStatement(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return evalBlockStatement(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalMatchExpression(me *ast.MatchExpression, env *Environment) Object {
	targetVal := Eval(me.Value, env)
	if isError(targetVal) {
		return targetVal
	}

	for _, c := range me.Cases {
		matchScope := NewEnclosedEnvironment(env)
		if matchesPattern(c.Pattern, targetVal, matchScope) {
			return Eval(c.Body, matchScope)
		}
	}

	return NULL
}

func matchesPattern(pat ast.Expression, val Object, scope *Environment) bool {
	switch p := pat.(type) {
	case *ast.Identifier:
		if p.Value == "_" {
			return true // Wildcard match
		}
		scope.Set(p.Value, val)
		return true

	case *ast.CallExpression:
		target := val
		if boxVal, ok := target.(*BoxInstance); ok {
			target = boxVal.Value
		}
		// Pattern: Result.Ok(val) / Result.Err(err)
		if dot, ok := p.Function.(*ast.DotExpression); ok {
			if resVal, ok := target.(*ResultInstance); ok {
				if dot.Right.Value == "Ok" && !resVal.IsErr && len(p.Arguments) == 1 {
					return matchesPattern(p.Arguments[0], resVal.Value, scope)
				}
				if dot.Right.Value == "Err" && resVal.IsErr && len(p.Arguments) == 1 {
					return matchesPattern(p.Arguments[0], resVal.Value, scope)
				}
			}
			if enumVal, ok := target.(*EnumInstance); ok {
				if enumVal.Variant == dot.Right.Value && len(enumVal.Fields) == len(p.Arguments) {
					for i, argPat := range p.Arguments {
						if !matchesPattern(argPat, enumVal.Fields[i], scope) {
							return false
						}
					}
					return true
				}
			}
		}
		if ident, ok := p.Function.(*ast.Identifier); ok {
			if resVal, ok := target.(*ResultInstance); ok {
				if ident.Value == "Ok" && !resVal.IsErr && len(p.Arguments) == 1 {
					return matchesPattern(p.Arguments[0], resVal.Value, scope)
				}
				if ident.Value == "Err" && resVal.IsErr && len(p.Arguments) == 1 {
					return matchesPattern(p.Arguments[0], resVal.Value, scope)
				}
			}
		}
		return false

	case *ast.DotExpression:
		// Pattern: Enum.Variant without arguments
		if enumVal, ok := val.(*EnumInstance); ok {
			return enumVal.Variant == p.Right.Value && len(enumVal.Fields) == 0
		}
		return false

	case *ast.NoneLiteral:
		if opt, ok := val.(*Option); ok && !opt.HasValue {
			return true
		}
		return false

	case *ast.SomeExpression:
		if opt, ok := val.(*Option); ok && opt.HasValue {
			return matchesPattern(p.Value, opt.Value, scope)
		}
		return false

	case *ast.IntegerLiteral:
		if intVal, ok := val.(*Integer); ok {
			return intVal.Value == p.Value
		}
		return false

	case *ast.StringLiteral:
		if strVal, ok := val.(*String); ok {
			return strVal.Value == p.Value
		}
		return false

	case *ast.BooleanLiteral:
		if boolVal, ok := val.(*Boolean); ok {
			return boolVal.Value == p.Value
		}
		return false
	}

	return false
}

func evalIndexExpression(left, index Object) Object {
	switch {
	case left.Type() == ARRAY_OBJ && index.Type() == INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == STRING_OBJ && index.Type() == INTEGER_OBJ:
		str := left.(*String).Value
		idx := index.(*Integer).Value
		if idx < 0 || idx >= int64(len(str)) {
			return NULL
		}
		return &String{Value: string(str[idx])}
	case left.Type() == MAP_OBJ:
		m := left.(*MapInstance)
		key := index.Inspect()
		if val, exists := m.Store[key]; exists {
			return val
		}
		return NULL
	default:
		return newError("index operator not supported: %s[%s]", left.Type(), index.Type())
	}
}

func evalDotExpression(left Object, prop string, env ...*Environment) Object {
	switch obj := left.(type) {
	case *BoxInstance:
		if prop == "unwrap" {
			return &Builtin{
				Fn: func(args ...Object) Object {
					return obj.Value
				},
			}
		}
		return evalDotExpression(obj.Value, prop, env...)

	case *StructInstance:
		if val, exists := obj.Fields[prop]; exists {
			return val
		}
		if len(env) > 0 && env[0] != nil {
			if defObj, exists := env[0].Get(obj.Name); exists {
				if def, ok := defObj.(*StructDefinition); ok {
					if method, ok := def.Methods[prop]; ok {
						return &Builtin{
							Fn: func(args ...Object) Object {
								callArgs := append([]Object{obj}, args...)
								return applyFunction(method, callArgs)
							},
						}
					}
				}
			}
		}
		if prop == "unwrap" {
			return &Builtin{
				Fn: func(args ...Object) Object {
					return left
				},
			}
		}
		return newError("field or method '%s' not found on struct '%s'", prop, obj.Name)

	case *Module:
		if method, exists := obj.Methods[prop]; exists {
			return method
		}
		return newError("function '%s' not found in module '%s'", prop, obj.Name)

	case *Array:
		switch prop {
		case "push":
			return &Builtin{
				Fn: func(args ...Object) Object {
					if len(args) > 0 {
						obj.Elements = append(obj.Elements, args[0])
					}
					return NULL
				},
			}
		case "pop":
			return &Builtin{
				Fn: func(args ...Object) Object {
					if len(obj.Elements) == 0 {
						return NONE
					}
					last := obj.Elements[len(obj.Elements)-1]
					obj.Elements = obj.Elements[:len(obj.Elements)-1]
					return &Option{HasValue: true, Value: last}
				},
			}
		case "len":
			return &Builtin{
				Fn: func(args ...Object) Object {
					return &Integer{Value: int64(len(obj.Elements))}
				},
			}
		case "clear":
			return &Builtin{
				Fn: func(args ...Object) Object {
					obj.Elements = []Object{}
					return NULL
				},
			}
		}

	case *String:
		switch prop {
		case "len":
			return &Builtin{
				Fn: func(args ...Object) Object {
					return &Integer{Value: int64(len(obj.Value))}
				},
			}
		case "sub":
			return &Builtin{
				Fn: func(args ...Object) Object {
					if len(args) != 2 {
						return newError("str.sub() takes 2 arguments (start, end)")
					}
					sInt, ok1 := args[0].(*Integer)
					eInt, ok2 := args[1].(*Integer)
					if !ok1 || !ok2 {
						return newError("str.sub() arguments must be integers")
					}
					start := sInt.Value
					end := eInt.Value
					sLen := int64(len(obj.Value))
					if start < 0 {
						start = 0
					}
					if end > sLen {
						end = sLen
					}
					if start > end {
						return &String{Value: ""}
					}
					return &String{Value: obj.Value[start:end]}
				},
			}
		case "char_at":
			return &Builtin{
				Fn: func(args ...Object) Object {
					if len(args) != 1 {
						return newError("str.char_at() takes 1 argument (index)")
					}
					idxObj, ok := args[0].(*Integer)
					if !ok {
						return newError("str.char_at() argument must be an integer")
					}
					idx := idxObj.Value
					if idx < 0 || idx >= int64(len(obj.Value)) {
						return &Integer{Value: -1}
					}
					return &Integer{Value: int64(obj.Value[idx])}
				},
			}
		case "contains":
			return &Builtin{
				Fn: func(args ...Object) Object {
					if len(args) != 1 {
						return newError("str.contains() takes 1 argument (substr)")
					}
					return nativeBoolToBooleanObject(strings.Contains(obj.Value, args[0].Inspect()))
				},
			}
		}

	case *MapInstance:
		switch prop {
		case "set":
			return &Builtin{
				Fn: func(args ...Object) Object {
					if len(args) == 2 {
						obj.Store[args[0].Inspect()] = args[1]
					}
					return NULL
				},
			}
		case "get":
			return &Builtin{
				Fn: func(args ...Object) Object {
					if len(args) == 1 {
						if val, exists := obj.Store[args[0].Inspect()]; exists {
							return &Option{HasValue: true, Value: val}
						}
					}
					return NONE
				},
			}
		case "has":
			return &Builtin{
				Fn: func(args ...Object) Object {
					if len(args) == 1 {
						_, exists := obj.Store[args[0].Inspect()]
						return nativeBoolToBooleanObject(exists)
					}
					return FALSE
				},
			}
		case "len":
			return &Builtin{
				Fn: func(args ...Object) Object {
					return &Integer{Value: int64(len(obj.Store))}
				},
			}
		}
	}

	if prop == "unwrap" {
		return &Builtin{
			Fn: func(args ...Object) Object {
				return left
			},
		}
	}

	return newError("property or method '%s' not found on %s", prop, left.Type())
}

func evalArrayIndexExpression(array, index Object) Object {
	arrayObject := array.(*Array)
	idx := index.(*Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalExpressions(exps []ast.Expression, env *Environment) []Object {
	var result []Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn Object, args []Object) Object {
	switch function := fn.(type) {
	case *Function:
		extendedEnv := extendFunctionEnv(function, args)
		evaluated := Eval(function.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *Builtin:
		return function.Fn(args...)

	case *EnumInstance:
		return &EnumInstance{
			EnumName: function.EnumName,
			Variant:  function.Variant,
			Fields:   args,
		}

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *Function, args []Object) *Environment {
	env := NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Name.Value, args[paramIdx])
		}
	}

	return env
}

func unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func isTruthy(obj Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		if opt, ok := obj.(*Option); ok {
			return opt.HasValue
		}
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

func evalImportStatement(stmt *ast.ImportStatement, env *Environment) Object {
	bytes, err := os.ReadFile(stmt.Path)
	if err != nil {
		return newError("cannot import '%s': %s", stmt.Path, err.Error())
	}
	l := lexer.New(string(bytes))
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return newError("syntax error in imported module '%s': %s", stmt.Path, p.Errors()[0])
	}
	moduleEnv := NewEnvironment()
	res := Eval(prog, moduleEnv)
	if isError(res) {
		return res
	}
	for _, sym := range stmt.Symbols {
		if val, ok := moduleEnv.Get(sym.Value); ok {
			env.Set(sym.Value, val)
		} else {
			return newError("symbol '%s' not found in module '%s'", sym.Value, stmt.Path)
		}
	}
	return NULL
}

func callExternC(fnName string, args []Object) Object {
	switch fnName {
	case "sqrt":
		if len(args) == 1 {
			val := getFloatValue(args[0])
			return &Float{Value: math.Sqrt(val)}
		}
	case "abs", "llabs", "fabs":
		if len(args) == 1 {
			if i, ok := args[0].(*Integer); ok {
				if i.Value < 0 {
					return &Integer{Value: -i.Value}
				}
				return i
			}
			val := getFloatValue(args[0])
			return &Float{Value: math.Abs(val)}
		}
	case "puts":
		if len(args) == 1 {
			fmt.Println(args[0].Inspect())
			return &Integer{Value: 0}
		}
	case "sin":
		if len(args) == 1 {
			return &Float{Value: math.Sin(getFloatValue(args[0]))}
		}
	case "cos":
		if len(args) == 1 {
			return &Float{Value: math.Cos(getFloatValue(args[0]))}
		}
	case "tan":
		if len(args) == 1 {
			return &Float{Value: math.Tan(getFloatValue(args[0]))}
		}
	case "pow":
		if len(args) == 2 {
			return &Float{Value: math.Pow(getFloatValue(args[0]), getFloatValue(args[1]))}
		}
	case "floor":
		if len(args) == 1 {
			return &Float{Value: math.Floor(getFloatValue(args[0]))}
		}
	case "ceil":
		if len(args) == 1 {
			return &Float{Value: math.Ceil(getFloatValue(args[0]))}
		}
	case "exit":
		if len(args) == 1 {
			if i, ok := args[0].(*Integer); ok {
				os.Exit(int(i.Value))
			}
			os.Exit(0)
		}
	}
	return NULL
}

func getFloatValue(obj Object) float64 {
	switch v := obj.(type) {
	case *Integer:
		return float64(v.Value)
	case *Float:
		return v.Value
	default:
		return 0
	}
}
