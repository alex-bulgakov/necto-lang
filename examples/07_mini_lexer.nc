// 07_mini_lexer.Necto
// Первый рабочий лексер компилятора, написанный на самом языке Necto!
// Демонстрирует готовность Necto к реализации собственного компилятора (Self-Hosting).

struct Token {
    kind: str
    literal: str
    pos: int
}

fn is_whitespace(code: int) -> bool {
    // 32 = ' ', 9 = '\t', 10 = '\n', 13 = '\r'
    if code == 32 || code == 9 || code == 10 || code == 13 {
        return true
    }
    return false
}

fn is_digit(code: int) -> bool {
    // '0'..'9' -> 48..57
    if code >= 48 && code <= 57 {
        return true
    }
    return false
}

fn is_alpha(code: int) -> bool {
    // 'a'..'z' -> 97..122, 'A'..'Z' -> 65..90, '_' -> 95
    if (code >= 97 && code <= 122) || (code >= 65 && code <= 90) || code == 95 {
        return true
    }
    return false
}

fn main() {
    println("==================================================")
    println("   Necto Mini-Lexer (written in 100% pure Necto)    ")
    println("==================================================")

    let source = "fn add(x: int, y: int) -> int { return x + y }"
    println(f"Source Code to tokenize:\n'{source}'\n")

    let mut keywords = Map.new()
    keywords.set("fn", "KEYWORD_FN")
    keywords.set("let", "KEYWORD_LET")
    keywords.set("return", "KEYWORD_RETURN")
    keywords.set("if", "KEYWORD_IF")

    let mut tokens: [Token] = []
    let length = source.len()
    let mut cursor = 0

    while cursor < length {
        let code = source.char_at(cursor)

        // 1. Пропускаем пробелы
        if is_whitespace(code) {
            cursor += 1
        } else if is_digit(code) {
            // 2. Сканируем числа
            let start = cursor
            while cursor < length && is_digit(source.char_at(cursor)) {
                cursor += 1
            }
            let lit = source.sub(start, cursor)
            tokens.push(Token { kind: "INT", literal: lit, pos: start })
        } else if is_alpha(code) {
            // 3. Сканируем идентификаторы и ключевые слова
            let start = cursor
            while cursor < length && (is_alpha(source.char_at(cursor)) || is_digit(source.char_at(cursor))) {
                cursor += 1
            }
            let ident = source.sub(start, cursor)
            let mut token_kind = "IDENT"
            if keywords.has(ident) {
                match keywords.get(ident) {
                    Some(k) => token_kind = k,
                    None => token_kind = "IDENT",
                }
            }
            tokens.push(Token { kind: token_kind, literal: ident, pos: start })
        } else {
            // 4. Односимвольные и двусимвольные операторы
            let start = cursor
            if code == 45 && cursor + 1 < length && source.char_at(cursor + 1) == 62 {
                // "->"
                tokens.push(Token { kind: "ARROW", literal: "->", pos: start })
                cursor += 2
            } else if code == 43 {
                tokens.push(Token { kind: "PLUS", literal: "+", pos: start })
                cursor += 1
            } else if code == 61 {
                tokens.push(Token { kind: "ASSIGN", literal: "=", pos: start })
                cursor += 1
            } else if code == 40 {
                tokens.push(Token { kind: "LPAREN", literal: "(", pos: start })
                cursor += 1
            } else if code == 41 {
                tokens.push(Token { kind: "RPAREN", literal: ")", pos: start })
                cursor += 1
            } else if code == 123 {
                tokens.push(Token { kind: "LBRACE", literal: "{", pos: start })
                cursor += 1
            } else if code == 125 {
                tokens.push(Token { kind: "RBRACE", literal: "}", pos: start })
                cursor += 1
            } else if code == 58 {
                tokens.push(Token { kind: "COLON", literal: ":", pos: start })
                cursor += 1
            } else if code == 44 {
                tokens.push(Token { kind: "COMMA", literal: ",", pos: start })
                cursor += 1
            } else {
                cursor += 1
            }
        }
    }

    println(f"Successfully tokenized! Total tokens: {tokens.len()}\n")
    println("--- Tokens Stream ---")
    for i in 0..tokens.len() {
        let tok = tokens[i]
        println(f"[{tok.pos}] {tok.kind}: '{tok.literal}'")
    }
}

