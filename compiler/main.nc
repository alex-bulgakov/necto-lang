// compiler/main.nc
// Точка входа самохостингового компилятора Necto (Stage 1)
// Написано на 100% чистом языке Necto!

import { Token } from "compiler/token.nc"
import { Lexer } from "compiler/lexer.nc"
import { Expr, Stmt } from "compiler/ast.nc"
import { Parser } from "compiler/parser.nc"
import { Codegen } from "compiler/codegen.nc"

fn compile_source_to_c(source: str) -> Result[str, str] {
    let mut lexer = Lexer.new(source)
    let tokens = lexer.tokenize_all()?

    let mut parser = Parser.new(tokens)
    let stmts = parser.parse_program()?

    let mut codegen = Codegen.new()
    let c_code = codegen.emit_program(stmts)

    return Result.Ok(c_code)
}

test "self-hosted lexer tokenize numbers and identifiers" {
    let mut lex = Lexer.new("let mut x = 42 + 10")
    let toks = lex.tokenize_all()
    match toks {
        Result.Ok(list) => {
            assert(list.len() == 8) // let, mut, x, =, 42, +, 10, Eof
        },
        Result.Err(e) => {
            assert(false)
        },
    }
}

test "self-hosted parser and codegen on arithmetic function" {
    let src = "fn add(a: int, b: int) -> int { return a + b; }\nfn main() { println(add(10, 20)); }"
    let res = compile_source_to_c(src)
    match res {
        Result.Ok(c) => {
            assert(c.contains("long long add(long long a, long long b)"))
            assert(c.contains("return (a + b);"))
        },
        Result.Err(e) => {
            println(f"Compilation error: {e}")
            assert(false)
        },
    }
}

fn main() {
    println("==================================================================")
    println("     Necto Self-Hosted Compiler (Stage 1) — Written in Necto      ")
    println("==================================================================")

    let test_program = "fn compute(x: int, y: int) -> int {\n    return (x * 2) + (y * 3);\n}\n\nfn main() {\n    let val = compute(5, 10);\n    println(val);\n}\n"

    println("Compiling test program in pure Necto:")
    println("------------------------------------------------------------------")
    println(test_program)
    println("------------------------------------------------------------------")

    println("Running self-hosted pipeline...")
    let result = compile_source_to_c(test_program)

    match result {
        Result.Ok(c_code) => {
            println("✓ Compilation succeeded! Output C code:\n")
            println(c_code)

            let out_file = "stage1_output.tmp.c"
            fs.write_file(out_file, c_code)
            println(f"✓ Saved native C code to '{out_file}'")
        },
        Result.Err(err) => {
            println(f"✗ Compilation failed: {err}")
        },
    }
}
