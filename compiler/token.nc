// compiler/token.nc
// Модуль определения токенов языка Necto для самохостингового компилятора

enum Token {
    // Литералы и идентификаторы
    Ident(str)
    IntLit(int)
    StrLit(str)
    BoolLit(bool)

    // Ключевые слова
    Fn
    Let
    Mut
    If
    Else
    While
    For
    In
    Return
    Struct
    Impl
    Enum
    Match
    Import
    From
    Test
    Assert
    Break
    Continue

    // Операторы
    Plus
    Minus
    Star
    Slash
    Percent
    Eq
    NotEq
    Lt
    Lte
    Gt
    Gte
    Assign
    PlusAssign
    MinusAssign
    Question
    Arrow
    FatArrow

    // Разделители
    LParen
    RParen
    LBrace
    RBrace
    LBracket
    RBracket
    Comma
    Colon
    Semicolon
    Dot
    DotDot

    // Служебные
    Eof
    Illegal(str)
}

fn lookup_keyword(name: str) -> Option[Token] {
    if name == "fn" { return Some(Token.Fn) }
    if name == "let" { return Some(Token.Let) }
    if name == "mut" { return Some(Token.Mut) }
    if name == "if" { return Some(Token.If) }
    if name == "else" { return Some(Token.Else) }
    if name == "while" { return Some(Token.While) }
    if name == "for" { return Some(Token.For) }
    if name == "in" { return Some(Token.In) }
    if name == "return" { return Some(Token.Return) }
    if name == "struct" { return Some(Token.Struct) }
    if name == "impl" { return Some(Token.Impl) }
    if name == "enum" { return Some(Token.Enum) }
    if name == "match" { return Some(Token.Match) }
    if name == "import" { return Some(Token.Import) }
    if name == "from" { return Some(Token.From) }
    if name == "test" { return Some(Token.Test) }
    if name == "assert" { return Some(Token.Assert) }
    if name == "break" { return Some(Token.Break) }
    if name == "continue" { return Some(Token.Continue) }
    if name == "true" { return Some(Token.BoolLit(true)) }
    if name == "false" { return Some(Token.BoolLit(false)) }
    return None
}
