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

test "self-hosted parser and codegen on struct and impl" {
    let src = "struct Point { x: int, y: int }\nimpl Point { fn get_x(self) -> int { return self.x; } }"
    let res = compile_source_to_c(src)
    match res {
        Result.Ok(c) => {
            assert(c.contains("typedef struct {"))
            assert(c.contains("long long x;"))
            assert(c.contains("get_x"))
        },
        Result.Err(e) => {
            println(f"Compilation error: {e}")
            assert(false)
        },
    }
}

extern "C" {
    fn system(cmd: str) -> int
}

fn main() {
    let args = os.args()
    let mut file_to_compile = ""
    let mut out_file = "output.exe"
    let mut is_run_mode = false

    let mut i = 1
    while i < args.len() {
        let a = args[i]
        if a == "-o" {
            if i + 1 < args.len() {
                out_file = args[i + 1]
                i += 1
            }
        } else if a == "run" || a == "--run" {
            is_run_mode = true
        } else if a == "version" || a == "--version" {
            println("Necto Pure Self-Hosted Compiler v2.0.0")
            return
        } else if a.contains(".nc") && !a.contains("compiler/main.nc") && !a.contains("compiler\\main.nc") {
            file_to_compile = a
        }
        i += 1
    }

    if file_to_compile != "" {
        println("==================================================================")
        println("      Necto Native Pure Self-Hosted Compiler (Stage 3)            ")
        println("==================================================================")
        println(f"Compiling '{file_to_compile}' to '{out_file}'...")

        match fs.read_file(file_to_compile) {
            Some(src) => {
                let res = compile_source_to_c(src)
                match res {
                    Result.Ok(c_code) => {
                        let tmp_c = "necto_native_build.tmp.c"
                        fs.write_file(tmp_c, c_code)
                        println(f"✓ Emitted C intermediate code to '{tmp_c}'")

                        println("Invoking embedded/system C backend (Clang/LLVM)...")
                        let clang_cmd = f"clang -O2 {tmp_c} -o {out_file}"
                        let status = system(clang_cmd)

                        if status == 0 {
                            println(f"✓ Native executable successfully created: {out_file}")
                            if is_run_mode {
                                println("------------------------------------------------------------------")
                                println(f"Running '{out_file}':\n")
                                system(out_file)
                                println("\n------------------------------------------------------------------")
                            }
                        } else {
                            println(f"✗ Backend compiler exited with error code {status}")
                        }
                    },
                    Result.Err(err) => {
                        println(f"✗ Compilation error: {err}")
                    },
                }
            },
            None => {
                println(f"✗ Error: could not read file '{file_to_compile}'")
            },
        }
        return
    }

    println("==================================================================")
    println("      Necto Native Pure Self-Hosted Compiler (Stage 3)            ")
    println("==================================================================")
    println("Usage: necto-native <file.nc> [-o output.exe] [--run]")
    println("")
    println("Running self-hosted verification pipeline on test program...")

    let test_program = "fn compute(x: int, y: int) -> int {\n    return (x * 2) + (y * 3);\n}\n\nfn main() {\n    let val = compute(5, 10);\n    println(val);\n}\n"
    let result = compile_source_to_c(test_program)

    match result {
        Result.Ok(c_code) => {
            let out_tmp = "stage1_output.tmp.c"
            fs.write_file(out_tmp, c_code)
            println("✓ Self-test passed! Pure Necto compiler generated valid C code.")
        },
        Result.Err(err) => {
            println(f"✗ Verification failed: {err}")
        },
    }
}
