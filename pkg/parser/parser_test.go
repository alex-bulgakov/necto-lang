package parser

import (
	"testing"

	"necto/pkg/ast"
	"necto/pkg/lexer"
)

func TestLetStatements(t *testing.T) {
	input := `
	let x = 5
	let mut y: int = 10
	let name = "Aura"
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
		isMut              bool
		typeAnnotation     string
	}{
		{"x", false, ""},
		{"y", true, "int"},
		{"name", false, ""},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		letStmt, ok := stmt.(*ast.LetStatement)
		if !ok {
			t.Fatalf("program.Statements[%d] not *ast.LetStatement. got=%T", i, stmt)
		}

		if letStmt.Name.Value != tt.expectedIdentifier {
			t.Fatalf("letStmt.Name.Value not '%s'. got='%s'", tt.expectedIdentifier, letStmt.Name.Value)
		}

		if letStmt.IsMut != tt.isMut {
			t.Fatalf("letStmt.IsMut not %v. got=%v", tt.isMut, letStmt.IsMut)
		}

		if letStmt.TypeAnnotation != tt.typeAnnotation {
			t.Fatalf("letStmt.TypeAnnotation not '%s'. got='%s'", tt.typeAnnotation, letStmt.TypeAnnotation)
		}
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true == false", "(true == false)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.Statements[0].(*ast.ExpressionStatement).Expression.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestFunctionDeclaration(t *testing.T) {
	input := `
	fn add(x: int, y: int) -> int {
		return x + y
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FnDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.FnDeclaration, got=%T", program.Statements[0])
	}

	if fnDecl.Name.Value != "add" {
		t.Fatalf("expected function name 'add', got=%s", fnDecl.Name.Value)
	}

	if len(fnDecl.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got=%d", len(fnDecl.Parameters))
	}

	if fnDecl.ReturnType != "int" {
		t.Fatalf("expected return type 'int', got=%s", fnDecl.ReturnType)
	}
}

func TestStructDeclarationAndLiteral(t *testing.T) {
	input := `
	struct Point {
		x: int
		y: int
	}

	let p = Point { x: 10, y: 20 }
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got=%d", len(program.Statements))
	}

	structDecl, ok := program.Statements[0].(*ast.StructDeclaration)
	if !ok {
		t.Fatalf("first statement is not *ast.StructDeclaration, got=%T", program.Statements[0])
	}

	if structDecl.Name.Value != "Point" {
		t.Fatalf("expected struct name 'Point', got=%s", structDecl.Name.Value)
	}

	if len(structDecl.Fields) != 2 {
		t.Fatalf("expected 2 fields, got=%d", len(structDecl.Fields))
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
