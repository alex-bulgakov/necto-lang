// 10_self_compiler_pipeline.nc
// Necto v0.4.0-alpha Milestone:
// Полноценный объектно-ориентированный пайплайн компилятора на чистом Necto
// с использованием методов (impl), Result[T, E] и оператора распространения ошибок '?'!

enum Token {
    Number(int)
    Plus
    Minus
    Star
    Slash
    LParen
    RParen
    Eof
}

enum Expr {
    Num(int)
    Binary(str, Box[Expr], Box[Expr])
}

// 1. Лексер со структурой и методами (impl Lexer)
struct Lexer {
    source: str
    cursor: int
}

impl Lexer {
    fn new(src: str) -> Lexer {
        return Lexer { source: src, cursor: 0 }
    }

    fn has_more(self) -> bool {
        return self.cursor < self.source.len()
    }

    fn peek_char(self) -> int {
        if self.cursor >= self.source.len() {
            return 0
        }
        return self.source.char_at(self.cursor)
    }

    fn next_char(mut self) -> int {
        let ch = self.peek_char()
        self.cursor += 1
        return ch
    }

    fn tokenize(mut self) -> Result[[Token], str] {
        let mut tokens: [Token] = []
        while self.has_more() {
            let ch = self.peek_char()
            // Пропуск пробельных символов
            if ch == 32 || ch == 9 || ch == 10 || ch == 13 {
                self.next_char()
            } else if ch >= 48 && ch <= 57 {
                let start = self.cursor
                while self.has_more() && (self.peek_char() >= 48 && self.peek_char() <= 57) {
                    self.next_char()
                }
                let num_str = self.source.sub(start, self.cursor)
                let mut n = 0
                for k in 0..num_str.len() {
                    n = n * 10 + (num_str.char_at(k) - 48)
                }
                tokens.push(Token.Number(n))
            } else if ch == 43 {
                self.next_char()
                tokens.push(Token.Plus)
            } else if ch == 45 {
                self.next_char()
                tokens.push(Token.Minus)
            } else if ch == 42 {
                self.next_char()
                tokens.push(Token.Star)
            } else if ch == 47 {
                self.next_char()
                tokens.push(Token.Slash)
            } else if ch == 40 {
                self.next_char()
                tokens.push(Token.LParen)
            } else if ch == 41 {
                self.next_char()
                tokens.push(Token.RParen)
            } else {
                return Result.Err(f"unexpected character with code {ch} at position {self.cursor}")
            }
        }
        tokens.push(Token.Eof)
        return Result.Ok(tokens)
    }
}

// 2. Парсер со структурой и методами (impl Parser)
struct Parser {
    tokens: [Token]
    cursor: int
}

impl Parser {
    fn new(toks: [Token]) -> Parser {
        return Parser { tokens: toks, cursor: 0 }
    }

    fn peek_token(self) -> Token {
        if self.cursor >= self.tokens.len() {
            return Token.Eof
        }
        return self.tokens[self.cursor]
    }

    fn next_token(mut self) -> Token {
        let tok = self.peek_token()
        self.cursor += 1
        return tok
    }

    fn parse_primary(mut self) -> Result[Expr, str] {
        let tok = self.next_token()
        match tok {
            Token.Number(n) => {
                return Result.Ok(Expr.Num(n))
            },
            Token.LParen => {
                let inner = self.parse_expr()?
                let closing = self.next_token()
                match closing {
                    Token.RParen => {
                        return Result.Ok(inner)
                    },
                    _ => {
                        return Result.Err("expected closing ')'")
                    },
                }
            },
            _ => {
                return Result.Err("expected number or '('")
            },
        }
    }

    fn parse_factor(mut self) -> Result[Expr, str] {
        let mut left = self.parse_primary()?

        while true {
            let tok = self.peek_token()
            match tok {
                Token.Star => {
                    self.next_token()
                    let right = self.parse_primary()?
                    left = Expr.Binary("*", Box.new(left), Box.new(right))
                },
                Token.Slash => {
                    self.next_token()
                    let right = self.parse_primary()?
                    left = Expr.Binary("/", Box.new(left), Box.new(right))
                },
                _ => {
                    break
                },
            }
        }
        return Result.Ok(left)
    }

    fn parse_expr(mut self) -> Result[Expr, str] {
        let mut left = self.parse_factor()?

        while true {
            let tok = self.peek_token()
            match tok {
                Token.Plus => {
                    self.next_token()
                    let right = self.parse_factor()?
                    left = Expr.Binary("+", Box.new(left), Box.new(right))
                },
                Token.Minus => {
                    self.next_token()
                    let right = self.parse_factor()?
                    left = Expr.Binary("-", Box.new(left), Box.new(right))
                },
                _ => {
                    break
                },
            }
        }
        return Result.Ok(left)
    }
}

// 3. Кодогенератор AST -> C код (impl Codegen)
struct Codegen {
    temp_var_id: int
}

impl Codegen {
    fn new() -> Codegen {
        return Codegen { temp_var_id: 0 }
    }

    fn emit_expr(mut self, e: Expr) -> str {
        match e {
            Expr.Num(val) => {
                return f"{val}"
            },
            Expr.Binary(op, left, right) => {
                let l_code = self.emit_expr(left.unwrap())
                let r_code = self.emit_expr(right.unwrap())
                return f"({l_code} {op} {r_code})"
            },
        }
    }

    fn generate_c_program(mut self, ast_root: Expr) -> str {
        let expr_code = self.emit_expr(ast_root)
        let mut code = "#include <stdio.h>\n\n"
        code += "int main() {\n"
        code += f"    long long result = {expr_code};\n"
        code += "    printf(\"Compiled program evaluated result: %lld\\n\", result);\n"
        code += "    return 0;\n"
        code += "}\n"
        return code
    }
}

// 4. Сквозной пайплайн с оператором '?'
fn run_compiler_pipeline(source: str) -> Result[str, str] {
    let mut lexer = Lexer.new(source)
    let tokens = lexer.tokenize()?

    let mut parser = Parser.new(tokens)
    let ast = parser.parse_expr()?

    let mut codegen = Codegen.new()
    let c_source = codegen.generate_c_program(ast)

    return Result.Ok(c_source)
}

test "compiler pipeline with arithmetic precedence" {
    let src = "2 + 3 * 4"
    let res = run_compiler_pipeline(src)
    match res {
        Result.Ok(code) => {
            assert(code.contains("(2 + (3 * 4))"))
        },
        Result.Err(e) => {
            assert(false)
        },
    }
}

test "compiler pipeline with parenthesis" {
    let src = "(100 - 20) / (2 + 2)"
    let res = run_compiler_pipeline(src)
    match res {
        Result.Ok(code) => {
            assert(code.contains("((100 - 20) / (2 + 2))"))
        },
        Result.Err(e) => {
            assert(false)
        },
    }
}

fn main() {
    println("==================================================================")
    println("   Necto v0.4.0: Self-Hosted Compiler Pipeline (Lexer+Parser+CG)  ")
    println("==================================================================")

    let expr_code = "(50 * 2 + 100) / (5 + 5)"
    println(f"Source Expression: {expr_code}\n")

    println("Running compilation pipeline (Lexer -> Parser -> Codegen)...")
    let pipeline_result = run_compiler_pipeline(expr_code)

    match pipeline_result {
        Result.Ok(generated_c) => {
            println("✓ Compilation succeeded! Generated C intermediate code:\n")
            println("------------------------------------------------------------------")
            println(generated_c)
            println("------------------------------------------------------------------")

            // Записываем сгенерированный код в файл
            let out_file = "generated_pipeline_out.c"
            fs.write_file(out_file, generated_c)
            println(f"✓ Saved generated code to '{out_file}'")
        },
        Result.Err(err_msg) => {
            println(f"✗ Compilation failed with error: {err_msg}")
        },
    }

    // Тест обработки ошибок с оператором '?'
    let invalid_expr = "10 + @42"
    println(f"\nTesting error handling on invalid source: '{invalid_expr}'...")
    let err_result = run_compiler_pipeline(invalid_expr)
    match err_result {
        Result.Ok(c) => println("Error expected but compilation succeeded!"),
        Result.Err(e) => println(f"✓ Successfully intercepted syntax error: '{e}'"),
    }
}
