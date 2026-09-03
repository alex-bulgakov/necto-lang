package format

import (
	"strings"
	"testing"
)

func TestFormatBasicProgram(t *testing.T) {
	input := `
fn add(a:int,b:int)->int{
return a+b;
}

fn main(){
let mut x=10;
if x>5{
println(x);
}else{
println(0);
}
}
`

	expected := `fn add(a: int, b: int) -> int {
    return a+b;
}

fn main() {
    let mut x=10;
    if x>5 {
        println(x);
    } else {
        println(0);
    }
}
`

	formatted, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %s", err)
	}

	if strings.TrimSpace(formatted) != strings.TrimSpace(expected) {
		t.Errorf("Formatted output mismatch.\nExpected:\n%s\nGot:\n%s", expected, formatted)
	}
}

func TestFormatCommentsPreservation(t *testing.T) {
	input := `
// Top level comment
struct User {
    // Field comment
    name: str,
    age: int,
}
`
	formatted, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %s", err)
	}

	if !strings.Contains(formatted, "// Top level comment") {
		t.Errorf("Top level comment lost")
	}
	if !strings.Contains(formatted, "    // Field comment") {
		t.Errorf("Field comment indentation incorrect")
	}
}
