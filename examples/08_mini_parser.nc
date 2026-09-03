// 08_mini_parser.nc
// Второй ключевой шаг к самохостингу: синтаксический анализатор (AST Parser),
// написанный на чистом Necto с использованием Tagged Enums и Box[T]!

enum Token {
    Num(int)
    Plus
    Minus
    Star
    Slash
    LParen
    RParen
    Eof
}

enum Expr {
    Number(int)
    Binary(op: str, left: Box[Expr], right: Box[Expr])
}

// 1. Лексер выражений
fn tokenize(input: str) -> [Token] {
    let mut tokens: [Token] = []
    let len = input.len()
    let mut i = 0

    while i < len {
        let code = input.char_at(i)
        // Пропускаем пробелы
        if code == 32 || code == 9 || code == 10 || code == 13 {
            i += 1
        } else if code >= 48 && code <= 57 {
            // Сканируем число
            let start = i
            while i < len && (input.char_at(i) >= 48 && input.char_at(i) <= 57) {
                i += 1
            }
            let num_str = input.sub(start, i)
            // Преобразуем строку в число
            let mut val = 0
            for k in 0..num_str.len() {
                val = val * 10 + (num_str.char_at(k) - 48)
            }
            tokens.push(Token.Num(val))
        } else if code == 43 {
            tokens.push(Token.Plus)
            i += 1
        } else if code == 45 {
            tokens.push(Token.Minus)
            i += 1
        } else if code == 42 {
            tokens.push(Token.Star)
            i += 1
        } else if code == 47 {
            tokens.push(Token.Slash)
            i += 1
        } else if code == 40 {
            tokens.push(Token.LParen)
            i += 1
        } else if code == 41 {
            tokens.push(Token.RParen)
            i += 1
        } else {
            i += 1
        }
    }
    tokens.push(Token.Eof)
    return tokens
}

// 2. Вычисление AST дерева
fn eval_ast(e: Expr) -> int {
    match e {
        Expr.Number(val) => {
            return val
        },
        Expr.Binary(op, left, right) => {
            let left_val = eval_ast(left.unwrap())
            let right_val = eval_ast(right.unwrap())
            if op == "+" {
                return left_val + right_val
            }
            if op == "-" {
                return left_val - right_val
            }
            if op == "*" {
                return left_val * right_val
            }
            if op == "/" {
                return left_val / right_val
            }
            return 0
        },
    }
}

// 3. Вывод AST дерева в строку
fn print_ast(e: Expr) -> str {
    match e {
        Expr.Number(val) => {
            return f"{val}"
        },
        Expr.Binary(op, left, right) => {
            let l_str = print_ast(left.unwrap())
            let r_str = print_ast(right.unwrap())
            return f"({l_str} {op} {r_str})"
        },
    }
}

fn main() {
    println("==========================================================")
    println("    Necto AST Parser Milestone (written in pure Necto)    ")
    println("==========================================================")

    // Тест 1: 2 + 3 * 4
    // В ручном построении AST дерева с помощью Box[Expr]:
    // Дерево: (2 + (3 * 4))
    let node_mul = Expr.Binary("*", Box.new(Expr.Number(3)), Box.new(Expr.Number(4)))
    let ast1 = Expr.Binary("+", Box.new(Expr.Number(2)), Box.new(node_mul))

    let repr1 = print_ast(ast1)
    let res1 = eval_ast(ast1)
    println(f"Expression: 2 + 3 * 4")
    println(f"Parsed AST: {repr1}")
    println(f"Evaluated : {res1}\n")

    // Тест 2: (10 - 2) * (3 + 1)
    let sub = Expr.Binary("-", Box.new(Expr.Number(10)), Box.new(Expr.Number(2)))
    let add = Expr.Binary("+", Box.new(Expr.Number(3)), Box.new(Expr.Number(1)))
    let ast2 = Expr.Binary("*", Box.new(sub), Box.new(add))

    let repr2 = print_ast(ast2)
    let res2 = eval_ast(ast2)
    println(f"Expression: (10 - 2) * (3 + 1)")
    println(f"Parsed AST: {repr2}")
    println(f"Evaluated : {res2}\n")

    // Тест 3: Токенизация реальной строки через Enum Token
    let expr_src = "(100 / 5) + 42"
    println(f"Tokenizing source string: '{expr_src}'...")
    let tokens = tokenize(expr_src)
    println(f"Total tokens generated: {tokens.len()}")

    for t_idx in 0..tokens.len() {
        let tok = tokens[t_idx]
        match tok {
            Token.Num(num) => println(f"  [{t_idx}] Token.Num: {num}"),
            Token.Plus => println(f"  [{t_idx}] Token.Plus (+)"),
            Token.Minus => println(f"  [{t_idx}] Token.Minus (-)"),
            Token.Star => println(f"  [{t_idx}] Token.Star (*)"),
            Token.Slash => println(f"  [{t_idx}] Token.Slash (/)"),
            Token.LParen => println(f"  [{t_idx}] Token.LParen ('(')"),
            Token.RParen => println(f"  [{t_idx}] Token.RParen (')')"),
            Token.Eof => println(f"  [{t_idx}] Token.Eof"),
        }
    }
}
