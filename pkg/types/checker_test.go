package types

import (
	"testing"

	"necto/pkg/lexer"
	"necto/pkg/parser"
)

func TestChecker_ValidProgram(t *testing.T) {
	input := `
	struct User {
		id: int
		name: str
	}

	fn greet(u: User) -> str {
		return f"Hello, {u.name}!"
	}

	let mut counter: int = 0
	counter += 1

	let u = User { id: 1, name: "Alice" }
	let msg = greet(u)
	`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	checker := NewChecker()
	checker.Check(prog)

	if len(checker.Errors()) > 0 {
		t.Fatalf("unexpected type check errors: %v", checker.Errors())
	}
}

func TestChecker_ImmutableReassignment(t *testing.T) {
	input := `
	let count = 10
	count = 20
	`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	checker := NewChecker()
	checker.Check(prog)

	if len(checker.Errors()) == 0 {
		t.Fatalf("expected error for reassigning immutable variable, got none")
	}
}

func TestChecker_TypeMismatch(t *testing.T) {
	input := `
	let x: int = "hello"
	`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	checker := NewChecker()
	checker.Check(prog)

	if len(checker.Errors()) == 0 {
		t.Fatalf("expected type mismatch error, got none")
	}
}
