package eval

import (
	"testing"

	"necto/pkg/lexer"
	"necto/pkg/parser"
)

func testEval(input string) Object {
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	env := NewEnvironment()
	return Eval(prog, env)
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalFunctionAndRecursion(t *testing.T) {
	input := `
	fn fib(n: int) -> int {
		if n <= 1 {
			return n
		}
		return fib(n - 1) + fib(n - 2)
	}
	fib(7)
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 13) // 0, 1, 1, 2, 3, 5, 8, 13
}

func TestEvalWhileAndForLoops(t *testing.T) {
	input := `
	let mut sum = 0
	for i in 1..5 {
		sum += i
	}
	sum
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10) // 1 + 2 + 3 + 4 = 10
}

func TestEvalStructAndFields(t *testing.T) {
	input := `
	struct Point {
		x: int
		y: int
	}

	let p = Point { x: 10, y: 25 }
	p.x + p.y
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 35)
}

func TestEvalOptionAndMatch(t *testing.T) {
	input := `
	let opt = Some(42)

	match opt {
		Some(val) => val * 2,
		None => 0,
	}
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 84)
}

func TestEvalFString(t *testing.T) {
	input := `
	let name = "Aura"
	let ver = 1
	f"Language: {name} v{ver}"
	`
	evaluated := testEval(input)
	str, ok := evaluated.(*String)
	if !ok {
		t.Fatalf("object is not String, got=%T (%+v)", evaluated, evaluated)
	}
	if str.Value != "Language: Aura v1" {
		t.Fatalf("expected 'Language: Aura v1', got=%q", str.Value)
	}
}

func TestEvalCollectionsAndIO(t *testing.T) {
	// Тестирование динамических массивов
	arrInput := `
	let mut list = [1, 2]
	list.push(3)
	list.push(4)
	list.len()
	`
	testIntegerObject(t, testEval(arrInput), 4)

	// Тестирование Map
	mapInput := `
	let mut dict = Map.new()
	dict.set("apple", 10)
	dict.set("banana", 20)
	dict["banana"]
	`
	testIntegerObject(t, testEval(mapInput), 20)

	// Тестирование строковых срезов и char_at
	strInput := `
	let s = "Hello, Aura!"
	let sub = s.sub(7, 11)
	sub
	`
	res := testEval(strInput)
	sRes, ok := res.(*String)
	if !ok || sRes.Value != "Aura" {
		t.Fatalf("expected 'Aura', got=%+v", res)
	}
}

func testIntegerObject(t *testing.T, obj Object, expected int64) bool {
	result, ok := obj.(*Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

func TestEvalEnumsAndBox(t *testing.T) {
	input := `
enum Token {
    Number(int)
    Ident(str)
    Eof
}

let t1 = Token.Number(42)
let t2 = Token.Eof

let res1 = match t1 {
    Token.Number(val) => val * 2,
    Token.Ident(name) => 0,
    Token.Eof => -1,
}

let res2 = match t2 {
    Token.Number(val) => 0,
    Token.Eof => 100,
    _ => 0,
}

let b = Box.new(res1 + res2)
let finalVal = b.unwrap()
finalVal
`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 184)
}
