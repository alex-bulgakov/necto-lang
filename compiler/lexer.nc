// compiler/lexer.nc
// Модуль лексического анализатора (Lexer) для самохостингового компилятора Necto

import { Token, lookup_keyword } from "compiler/token.nc"

struct Lexer {
    source: str
    cursor: int
    line: int
    col: int
}

impl Lexer {
    fn new(src: str) -> Lexer {
        return Lexer {
            source: src,
            cursor: 0,
            line: 1,
            col: 1,
        }
    }

    fn has_more(self) -> bool {
        return self.cursor < self.source.len()
    }

    fn peek(self) -> int {
        if self.cursor >= self.source.len() {
            return 0
        }
        return self.source.char_at(self.cursor)
    }

    fn peek_next(self) -> int {
        if self.cursor + 1 >= self.source.len() {
            return 0
        }
        return self.source.char_at(self.cursor + 1)
    }

    fn advance(mut self) -> int {
        let ch = self.peek()
        self.cursor += 1
        if ch == 10 {
            self.line += 1
            self.col = 1
        } else {
            self.col += 1
        }
        return ch
    }

    fn skip_whitespace_and_comments(mut self) {
        while self.has_more() {
            let ch = self.peek()
            // Пробелы, табы, переводы строк
            if ch == 32 || ch == 9 || ch == 10 || ch == 13 {
                self.advance()
            } else if ch == 47 && self.peek_next() == 47 {
                // Однострочный комментарий // ...
                while self.has_more() && self.peek() != 10 {
                    self.advance()
                }
            } else {
                break
            }
        }
    }

    fn scan_number(mut self) -> Token {
        let start = self.cursor
        while self.has_more() && (self.peek() >= 48 && self.peek() <= 57) {
            self.advance()
        }
        let num_str = self.source.sub(start, self.cursor)
        let mut val = 0
        for i in 0..num_str.len() {
            val = val * 10 + (num_str.char_at(i) - 48)
        }
        return Token.IntLit(val)
    }

    fn scan_ident_or_keyword(mut self) -> Token {
        let start = self.cursor
        while self.has_more() {
            let ch = self.peek()
            let is_alpha = (ch >= 65 && ch <= 90) || (ch >= 97 && ch <= 122) || ch == 95
            let is_digit = ch >= 48 && ch <= 57
            if is_alpha || is_digit {
                self.advance()
            } else {
                break
            }
        }
        let word = self.source.sub(start, self.cursor)
        let kw = lookup_keyword(word)
        match kw {
            Some(tok) => {
                return tok
            },
            None => {
                return Token.Ident(word)
            },
        }
    }

    fn scan_string(mut self) -> Result[Token, str] {
        self.advance() // пропускаем открывающую кавычку "
        let start = self.cursor
        while self.has_more() && self.peek() != 34 {
            if self.peek() == 10 {
                return Result.Err("unterminated string literal")
            }
            self.advance()
        }
        if !self.has_more() {
            return Result.Err("unterminated string literal at EOF")
        }
        let str_val = self.source.sub(start, self.cursor)
        self.advance() // пропускаем закрывающую кавычку "
        return Result.Ok(Token.StrLit(str_val))
    }

    fn next_token(mut self) -> Result[Token, str] {
        self.skip_whitespace_and_comments()

        if !self.has_more() {
            return Result.Ok(Token.Eof)
        }

        let ch = self.peek()

        // Числа
        if ch >= 48 && ch <= 57 {
            return Result.Ok(self.scan_number())
        }

        // F-строки: f"..."
        if ch == 102 && self.peek_next() == 34 {
            self.advance() // пропускаем 'f'
            return self.scan_string()
        }

        // Идентификаторы и ключевые слова
        let is_alpha = (ch >= 65 && ch <= 90) || (ch >= 97 && ch <= 122) || ch == 95
        if is_alpha {
            return Result.Ok(self.scan_ident_or_keyword())
        }

        // Строки
        if ch == 34 {
            let s_tok = self.scan_string()?
            return Result.Ok(s_tok)
        }

        // Двухсимвольные операторы
        let next_ch = self.peek_next()

        if ch == 61 && next_ch == 61 {
            self.advance()
            self.advance()
            return Result.Ok(Token.Eq)
        }
        if ch == 33 && next_ch == 61 {
            self.advance()
            self.advance()
            return Result.Ok(Token.NotEq)
        }
        if ch == 60 && next_ch == 61 {
            self.advance()
            self.advance()
            return Result.Ok(Token.Lte)
        }
        if ch == 62 && next_ch == 61 {
            self.advance()
            self.advance()
            return Result.Ok(Token.Gte)
        }
        if ch == 43 && next_ch == 61 {
            self.advance()
            self.advance()
            return Result.Ok(Token.PlusAssign)
        }
        if ch == 45 && next_ch == 61 {
            self.advance()
            self.advance()
            return Result.Ok(Token.MinusAssign)
        }
        if ch == 45 && next_ch == 62 {
            self.advance()
            self.advance()
            return Result.Ok(Token.Arrow)
        }
        if ch == 61 && next_ch == 62 {
            self.advance()
            self.advance()
            return Result.Ok(Token.FatArrow)
        }
        if ch == 46 && next_ch == 46 {
            self.advance()
            self.advance()
            return Result.Ok(Token.DotDot)
        }

        // Односимвольные операторы и разделители
        self.advance()

        if ch == 43 { return Result.Ok(Token.Plus) }
        if ch == 45 { return Result.Ok(Token.Minus) }
        if ch == 42 { return Result.Ok(Token.Star) }
        if ch == 47 { return Result.Ok(Token.Slash) }
        if ch == 37 { return Result.Ok(Token.Percent) }
        if ch == 61 { return Result.Ok(Token.Assign) }
        if ch == 60 { return Result.Ok(Token.Lt) }
        if ch == 62 { return Result.Ok(Token.Gt) }
        if ch == 63 { return Result.Ok(Token.Question) }
        if ch == 40 { return Result.Ok(Token.LParen) }
        if ch == 41 { return Result.Ok(Token.RParen) }
        if ch == 123 { return Result.Ok(Token.LBrace) }
        if ch == 125 { return Result.Ok(Token.RBrace) }
        if ch == 91 { return Result.Ok(Token.LBracket) }
        if ch == 93 { return Result.Ok(Token.RBracket) }
        if ch == 44 { return Result.Ok(Token.Comma) }
        if ch == 58 { return Result.Ok(Token.Colon) }
        if ch == 59 { return Result.Ok(Token.Semicolon) }
        if ch == 46 { return Result.Ok(Token.Dot) }

        return Result.Err(f"unexpected character '{ch}' at {self.line}:{self.col}")
    }

    fn tokenize_all(mut self) -> Result[[Token], str] {
        let mut list: [Token] = []
        while true {
            let tok = self.next_token()?
            match tok {
                Token.Eof => {
                    list.push(Token.Eof)
                    break
                },
                _ => {
                    list.push(tok)
                },
            }
        }
        return Result.Ok(list)
    }
}
