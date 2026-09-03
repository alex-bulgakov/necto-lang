package lexer

import (
	"testing"

	"necto/pkg/token"
)

func TestNextToken(t *testing.T) {
	input := `
	// A sample Aura snippet
	let mut count = 42
	let pi = 3.1415
	let name = "Alice"
	let msg = f"Hello, {name}!"

	fn add(a: int, b: int) -> int {
		return a + b
	}

	if count >= 40 && count != 50 {
		count += 1
	}

	match opt {
		Some(v) => v,
		None => 0,
	}
	`

	expected := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.MUT, "mut"},
		{token.IDENT, "count"},
		{token.ASSIGN, "="},
		{token.INT, "42"},

		{token.LET, "let"},
		{token.IDENT, "pi"},
		{token.ASSIGN, "="},
		{token.FLOAT, "3.1415"},

		{token.LET, "let"},
		{token.IDENT, "name"},
		{token.ASSIGN, "="},
		{token.STRING, "Alice"},

		{token.LET, "let"},
		{token.IDENT, "msg"},
		{token.ASSIGN, "="},
		{token.FSTRING, "Hello, {name}!"},

		{token.FN, "fn"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COLON, ":"},
		{token.IDENT, "int"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.COLON, ":"},
		{token.IDENT, "int"},
		{token.RPAREN, ")"},
		{token.ARROW, "->"},
		{token.IDENT, "int"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.IDENT, "b"},
		{token.RBRACE, "}"},

		{token.IF, "if"},
		{token.IDENT, "count"},
		{token.GTE, ">="},
		{token.INT, "40"},
		{token.AND, "&&"},
		{token.IDENT, "count"},
		{token.NOT_EQ, "!="},
		{token.INT, "50"},
		{token.LBRACE, "{"},
		{token.IDENT, "count"},
		{token.PLUS_ASSIGN, "+="},
		{token.INT, "1"},
		{token.RBRACE, "}"},

		{token.MATCH, "match"},
		{token.IDENT, "opt"},
		{token.LBRACE, "{"},
		{token.SOME, "Some"},
		{token.LPAREN, "("},
		{token.IDENT, "v"},
		{token.RPAREN, ")"},
		{token.FATARROW, "=>"},
		{token.IDENT, "v"},
		{token.COMMA, ","},
		{token.NONE, "None"},
		{token.FATARROW, "=>"},
		{token.INT, "0"},
		{token.COMMA, ","},
		{token.RBRACE, "}"},

		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range expected {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
