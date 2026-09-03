package codegen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"necto/pkg/ast"
)

type Compiler struct {
	buf            bytes.Buffer
	indent         int
	structs        map[string]*ast.StructDeclaration
	tempVarCounter int
}

func NewCompiler() *Compiler {
	return &Compiler{
		structs: make(map[string]*ast.StructDeclaration),
	}
}

func (c *Compiler) CompileToC(program *ast.Program) (string, error) {
	c.buf.Reset()

	// Заголовочные файлы
	c.writeLine("#include <stdio.h>")
	c.writeLine("#include <stdlib.h>")
	c.writeLine("#include <stdint.h>")
	c.writeLine("#include <stdbool.h>")
	c.writeLine("#include <string.h>")
	c.writeLine("#include <math.h>")
	c.writeLine("")

	// Вспомогательные функции для Necto Runtime
	c.writeLine("// --- Necto Minimal C Runtime ---")
	c.writeLine("static inline void necto_println_int(int64_t v) { printf(\"%lld\\n\", (long long)v); }")
	c.writeLine("static inline void necto_println_float(double v) { printf(\"%g\\n\", v); }")
	c.writeLine("static inline void necto_println_str(const char* v) { printf(\"%s\\n\", v); }")
	c.writeLine("static inline void necto_println_bool(bool v) { printf(\"%s\\n\", v ? \"true\" : \"false\"); }")
	c.writeLine("")

	// Первый проход: структуры
	for _, stmt := range program.Statements {
		if sd, ok := stmt.(*ast.StructDeclaration); ok {
			c.structs[sd.Name.Value] = sd
			c.writeLine(fmt.Sprintf("typedef struct %s %s;", sd.Name.Value, sd.Name.Value))
		}
	}
	c.writeLine("")

	for _, stmt := range program.Statements {
		if sd, ok := stmt.(*ast.StructDeclaration); ok {
			c.writeLine(fmt.Sprintf("struct %s {", sd.Name.Value))
			c.indent++
			for _, f := range sd.Fields {
				cType := mapTypeToC(f.Type)
				c.writeLine(fmt.Sprintf("%s %s;", cType, f.Name.Value))
			}
			c.indent--
			c.writeLine("};")
			c.writeLine("")
		}
	}

	// Второй проход: предварительное объявление функций
	for _, stmt := range program.Statements {
		if fd, ok := stmt.(*ast.FnDeclaration); ok {
			if fd.Name.Value != "main" {
				sig := c.generateFnSignature(fd)
				c.writeLine(sig + ";")
			}
		}
	}
	c.writeLine("")

	// Третий проход: определение функций
	for _, stmt := range program.Statements {
		if fd, ok := stmt.(*ast.FnDeclaration); ok {
			if err := c.compileFn(fd); err != nil {
				return "", err
			}
		}
	}

	// Проверяем, есть ли операторы верхнего уровня помимо функций и структур.
	// Если явной main() нет, собираем верхнеуровневый код в main().
	hasExplicitMain := false
	var topLevelStmts []ast.Statement

	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.FnDeclaration:
			if s.Name.Value == "main" {
				hasExplicitMain = true
			}
		case *ast.StructDeclaration:
			// уже скомпилировано
		default:
			topLevelStmts = append(topLevelStmts, s)
		}
	}

	if !hasExplicitMain {
		c.writeLine("int main(int argc, char** argv) {")
		c.indent++
		for _, s := range topLevelStmts {
			c.compileStatement(s)
		}
		c.writeLine("return 0;")
		c.indent--
		c.writeLine("}")
	}

	return c.buf.String(), nil
}

func (c *Compiler) generateFnSignature(fd *ast.FnDeclaration) string {
	retType := mapTypeToC(fd.ReturnType)
	var params []string
	for _, p := range fd.Parameters {
		params = append(params, fmt.Sprintf("%s %s", mapTypeToC(p.Type), p.Name.Value))
	}
	if len(params) == 0 {
		params = append(params, "void")
	}
	return fmt.Sprintf("%s %s(%s)", retType, fd.Name.Value, strings.Join(params, ", "))
}

func (c *Compiler) compileFn(fd *ast.FnDeclaration) error {
	if fd.Name.Value == "main" {
		c.writeLine("int main(int argc, char** argv) {")
	} else {
		sig := c.generateFnSignature(fd)
		c.writeLine(sig + " {")
	}
	c.indent++

	for _, stmt := range fd.Body.Statements {
		c.compileStatement(stmt)
	}

	if fd.Name.Value == "main" {
		c.writeLine("return 0;")
	}
	c.indent--
	c.writeLine("}")
	c.writeLine("")
	return nil
}

func (c *Compiler) compileStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		valExpr := "0"
		if s.Value != nil {
			valExpr = c.compileExpression(s.Value)
		}
		cType := "int64_t"
		if s.TypeAnnotation != "" {
			cType = mapTypeToC(s.TypeAnnotation)
		} else if s.Value != nil {
			cType = c.inferCType(s.Value)
		}

		qualifier := "const "
		if s.IsMut {
			qualifier = ""
		}
		c.writeLine(fmt.Sprintf("%s%s %s = %s;", qualifier, cType, s.Name.Value, valExpr))

	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			c.writeLine(fmt.Sprintf("return %s;", c.compileExpression(s.ReturnValue)))
		} else {
			c.writeLine("return;")
		}

	case *ast.ExpressionStatement:
		if ifExpr, ok := s.Expression.(*ast.IfExpression); ok {
			c.writeLine(fmt.Sprintf("if (%s) {", c.compileExpression(ifExpr.Condition)))
			c.indent++
			for _, bs := range ifExpr.Consequence.Statements {
				c.compileStatement(bs)
			}
			c.indent--
			if ifExpr.Alternative != nil {
				c.writeLine("} else {")
				c.indent++
				for _, bs := range ifExpr.Alternative.Statements {
					c.compileStatement(bs)
				}
				c.indent--
			}
			c.writeLine("}")
			return
		}

		if call, ok := s.Expression.(*ast.CallExpression); ok {
			if ident, ok := call.Function.(*ast.Identifier); ok && ident.Value == "println" {
				c.compilePrintlnStatement(call)
				return
			}
		}

		exprStr := c.compileExpression(s.Expression)
		if exprStr != "" {
			c.writeLine(exprStr + ";")
		}

	case *ast.WhileStatement:
		c.writeLine(fmt.Sprintf("while (%s) {", c.compileExpression(s.Condition)))
		c.indent++
		for _, bs := range s.Body.Statements {
			c.compileStatement(bs)
		}
		c.indent--
		c.writeLine("}")

	case *ast.ForInStatement:
		if rng, ok := s.Iterable.(*ast.RangeExpression); ok {
			start := c.compileExpression(rng.Start)
			end := c.compileExpression(rng.End)
			c.writeLine(fmt.Sprintf("for (int64_t %s = %s; %s < %s; %s++) {",
				s.Item.Value, start, s.Item.Value, end, s.Item.Value))
			c.indent++
			for _, bs := range s.Body.Statements {
				c.compileStatement(bs)
			}
			c.indent--
			c.writeLine("}")
		}

	case *ast.BreakStatement:
		c.writeLine("break;")

	case *ast.ContinueStatement:
		c.writeLine("continue;")
	}
}

func (c *Compiler) compilePrintlnStatement(call *ast.CallExpression) {
	if len(call.Arguments) == 0 {
		c.writeLine("printf(\"\\n\");")
		return
	}

	for _, arg := range call.Arguments {
		if fstr, ok := arg.(*ast.FStringLiteral); ok {
			for _, part := range fstr.Parts {
				if sLit, ok := part.(*ast.StringLiteral); ok {
					c.writeLine(fmt.Sprintf("printf(\"%%s\", %q);", sLit.Value))
				} else {
					pExpr := c.compileExpression(part)
					pType := c.inferCType(part)
					switch pType {
					case "double":
						c.writeLine(fmt.Sprintf("printf(\"%%g\", %s);", pExpr))
					case "const char*":
						c.writeLine(fmt.Sprintf("printf(\"%%s\", %s);", pExpr))
					case "bool":
						c.writeLine(fmt.Sprintf("printf(\"%%s\", (%s) ? \"true\" : \"false\");", pExpr))
					default:
						c.writeLine(fmt.Sprintf("printf(\"%%lld\", (long long)(%s));", pExpr))
					}
				}
			}
			c.writeLine("printf(\"\\n\");")
			return
		}

		if sLit, ok := arg.(*ast.StringLiteral); ok {
			c.writeLine(fmt.Sprintf("printf(\"%%s\\n\", %q);", sLit.Value))
			return
		}

		argExpr := c.compileExpression(arg)
		t := c.inferCType(arg)
		switch t {
		case "const char*":
			c.writeLine(fmt.Sprintf("necto_println_str(%s);", argExpr))
		case "double":
			c.writeLine(fmt.Sprintf("necto_println_float(%s);", argExpr))
		case "bool":
			c.writeLine(fmt.Sprintf("necto_println_bool(%s);", argExpr))
		default:
			c.writeLine(fmt.Sprintf("necto_println_int(%s);", argExpr))
		}
	}
}

func (c *Compiler) compileExpression(expr ast.Expression) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%dLL", e.Value)
	case *ast.FloatLiteral:
		return fmt.Sprintf("%g", e.Value)
	case *ast.BooleanLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value)
	case *ast.Identifier:
		return e.Value

	case *ast.PrefixExpression:
		return fmt.Sprintf("(%s%s)", e.Operator, c.compileExpression(e.Right))

	case *ast.InfixExpression:
		left := c.compileExpression(e.Left)
		right := c.compileExpression(e.Right)
		return fmt.Sprintf("(%s %s %s)", left, e.Operator, right)

	case *ast.AssignExpression:
		left := c.compileExpression(e.Left)
		right := c.compileExpression(e.Value)
		return fmt.Sprintf("%s %s %s", left, e.Operator, right)

	case *ast.CallExpression:
		fnName := ""
		if ident, ok := e.Function.(*ast.Identifier); ok {
			fnName = ident.Value
		}

		if fnName == "println" && len(e.Arguments) == 1 {
			arg := e.Arguments[0]
			argExpr := c.compileExpression(arg)
			t := c.inferCType(arg)
			switch t {
			case "const char*":
				return fmt.Sprintf("necto_println_str(%s)", argExpr)
			case "double":
				return fmt.Sprintf("necto_println_float(%s)", argExpr)
			case "bool":
				return fmt.Sprintf("necto_println_bool(%s)", argExpr)
			default:
				return fmt.Sprintf("necto_println_int(%s)", argExpr)
			}
		}

		var args []string
		for _, a := range e.Arguments {
			args = append(args, c.compileExpression(a))
		}
		return fmt.Sprintf("%s(%s)", c.compileExpression(e.Function), strings.Join(args, ", "))

	case *ast.DotExpression:
		return fmt.Sprintf("%s.%s", c.compileExpression(e.Left), e.Right.Value)

	case *ast.StructLiteral:
		var fieldInits []string
		for _, f := range e.Fields {
			fieldInits = append(fieldInits, fmt.Sprintf(".%s = %s", f.Name, c.compileExpression(f.Value)))
		}
		return fmt.Sprintf("(%s){ %s }", e.StructName.Value, strings.Join(fieldInits, ", "))

	case *ast.IfExpression:
		// Для оператора if внутри выражений
		return fmt.Sprintf("(%s ? ... : ...)", c.compileExpression(e.Condition))
	}

	return ""
}

func (c *Compiler) writeLine(line string) {
	indentation := strings.Repeat("  ", c.indent)
	c.buf.WriteString(indentation + line + "\n")
}

func mapTypeToC(auraType string) string {
	switch auraType {
	case "int":
		return "int64_t"
	case "float":
		return "double"
	case "bool":
		return "bool"
	case "str", "string":
		return "const char*"
	case "void", "":
		return "void"
	default:
		return auraType
	}
}

func (c *Compiler) inferCType(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return "int64_t"
	case *ast.FloatLiteral:
		return "double"
	case *ast.BooleanLiteral:
		return "bool"
	case *ast.StringLiteral, *ast.FStringLiteral:
		return "const char*"
	case *ast.StructLiteral:
		return e.StructName.Value
	case *ast.DotExpression:
		for _, sd := range c.structs {
			for _, f := range sd.Fields {
				if f.Name.Value == e.Right.Value {
					return mapTypeToC(f.Type)
				}
			}
		}
	}
	return "int64_t"
}

// BuildNativeExecutable компилирует AST в C код, а затем вызывает Clang для получения .exe
func BuildNativeExecutable(program *ast.Program, outputPath string) error {
	cCompiler := NewCompiler()
	cCode, err := cCompiler.CompileToC(program)
	if err != nil {
		return fmt.Errorf("code generation error: %w", err)
	}

	// Создаем временный .c файл
	tmpC := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".necto.tmp.c"
	err = os.WriteFile(tmpC, []byte(cCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write intermediate C file: %w", err)
	}
	defer os.Remove(tmpC)

	// Ищем clang или gcc
	compilerPath, err := exec.LookPath("clang")
	if err != nil {
		compilerPath, err = exec.LookPath("gcc")
		if err != nil {
			return fmt.Errorf("neither clang nor gcc found in PATH")
		}
	}

	cmd := exec.Command(compilerPath, "-O2", tmpC, "-o", outputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("native compilation failed:\n%s\n%w", string(out), err)
	}

	return nil
}
