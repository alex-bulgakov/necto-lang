// compiler/ast.nc
// Модуль синтаксического дерева (AST) для самохостингового компилятора Necto

enum Expr {
    Num(int)
    StrVal(str)
    BoolVal(bool)
    Var(str)
    Binary(str, Box[Expr], Box[Expr])
    Unary(str, Box[Expr])
    Call(str, [Expr])
    MethodCall(Box[Expr], str, [Expr])
    Try(Box[Expr])
    BoxNew(Box[Expr])
    Dot(Box[Expr], str)
}

enum Stmt {
    Let(str, bool, str, Expr)
    Assign(str, str, Expr)
    Return(Expr)
    ExprStmt(Expr)
    If(Expr, [Stmt], [Stmt])
    While(Expr, [Stmt])
    ForIn(str, int, int, [Stmt])
    FnDecl(str, [str], str, [Stmt])
    StructDecl(str, [str])
    Assert(Expr)
}
