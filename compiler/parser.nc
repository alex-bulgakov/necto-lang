// compiler/parser.nc
// Модуль синтаксического анализатора (Parser) для самохостингового компилятора Necto

import { Token } from "compiler/token.nc"
import { Expr, Stmt } from "compiler/ast.nc"

struct Parser {
    tokens: [Token]
    cursor: int
}

impl Parser {
    fn new(toks: [Token]) -> Parser {
        return Parser { tokens: toks, cursor: 0 }
    }

    fn has_more(self) -> bool {
        return self.cursor < self.tokens.len()
    }

    fn peek(self) -> Token {
        if self.cursor >= self.tokens.len() {
            return Token.Eof
        }
        return self.tokens[self.cursor]
    }

    fn advance(mut self) -> Token {
        let tok = self.peek()
        self.cursor += 1
        return tok
    }

    fn parse_primary(mut self) -> Result[Expr, str] {
        let tok = self.advance()
        match tok {
            Token.IntLit(n) => {
                return Result.Ok(Expr.Num(n))
            },
            Token.StrLit(s) => {
                return Result.Ok(Expr.StrVal(s))
            },
            Token.BoolLit(b) => {
                return Result.Ok(Expr.BoolVal(b))
            },
            Token.Ident(name) => {
                // Проверяем вызов функции: name(...)
                let next_t = self.peek()
                match next_t {
                    Token.LParen => {
                        self.advance() // пропускаем (
                        let mut args: [Expr] = []
                        let mut closed = false
                        while self.has_more() {
                            let check_close = self.peek()
                            match check_close {
                                Token.RParen => {
                                    self.advance()
                                    closed = true
                                    break
                                },
                                _ => {},
                            }
                            let arg = self.parse_expr()?
                            args.push(arg)
                            let sep = self.peek()
                            match sep {
                                Token.Comma => {
                                    self.advance()
                                },
                                Token.RParen => {
                                    self.advance()
                                    closed = true
                                    break
                                },
                                _ => {
                                    return Result.Err("expected ',' or ')' in argument list")
                                },
                            }
                        }
                        if !closed {
                            return Result.Err("unclosed '(' in function call")
                        }
                        return Result.Ok(Expr.Call(name, args))
                    },
                    _ => {
                        return Result.Ok(Expr.Var(name))
                    },
                }
            },
            Token.LParen => {
                let inner = self.parse_expr()?
                let closing = self.advance()
                match closing {
                    Token.RParen => {
                        return Result.Ok(inner)
                    },
                    _ => {
                        return Result.Err("expected ')' after expression")
                    },
                }
            },
            _ => {
                return Result.Err(f"unexpected token in primary expression: {tok}")
            },
        }
    }

    fn parse_postfix(mut self) -> Result[Expr, str] {
        let mut expr = self.parse_primary()?

        while self.has_more() {
            let tok = self.peek()
            match tok {
                Token.Dot => {
                    self.advance() // .
                    let id_tok = self.advance()
                    let mut member_name = ""
                    match id_tok {
                        Token.Ident(mname) => { member_name = mname },
                        _ => { return Result.Err("expected member name after '.'") },
                    }
                    if self.peek() == Token.LParen {
                        self.advance() // (
                        let mut args: [Expr] = []
                        while self.has_more() {
                            if self.peek() == Token.RParen {
                                self.advance()
                                break
                            }
                            let arg = self.parse_expr()?
                            args.push(arg)
                            if self.peek() == Token.Comma {
                                self.advance()
                            } else if self.peek() == Token.RParen {
                                self.advance()
                                break
                            }
                        }
                        expr = Expr.MethodCall(Box.new(expr), member_name, args)
                    } else {
                        expr = Expr.Dot(Box.new(expr), member_name)
                    }
                },
                _ => {
                    break
                },
            }
        }
        return Result.Ok(expr)
    }

    fn parse_factor(mut self) -> Result[Expr, str] {
        let mut left = self.parse_postfix()?

        while self.has_more() {
            let tok = self.peek()
            match tok {
                Token.Star => {
                    self.advance()
                    let right = self.parse_postfix()?
                    left = Expr.Binary("*", Box.new(left), Box.new(right))
                },
                Token.Slash => {
                    self.advance()
                    let right = self.parse_postfix()?
                    left = Expr.Binary("/", Box.new(left), Box.new(right))
                },
                Token.Percent => {
                    self.advance()
                    let right = self.parse_postfix()?
                    left = Expr.Binary("%", Box.new(left), Box.new(right))
                },
                _ => {
                    break
                },
            }
        }
        return Result.Ok(left)
    }

    fn parse_term(mut self) -> Result[Expr, str] {
        let mut left = self.parse_factor()?

        while self.has_more() {
            let tok = self.peek()
            match tok {
                Token.Plus => {
                    self.advance()
                    let right = self.parse_factor()?
                    left = Expr.Binary("+", Box.new(left), Box.new(right))
                },
                Token.Minus => {
                    self.advance()
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

    fn parse_comparison(mut self) -> Result[Expr, str] {
        let mut left = self.parse_term()?

        while self.has_more() {
            let tok = self.peek()
            match tok {
                Token.Lt => {
                    self.advance()
                    let right = self.parse_term()?
                    left = Expr.Binary("<", Box.new(left), Box.new(right))
                },
                Token.Lte => {
                    self.advance()
                    let right = self.parse_term()?
                    left = Expr.Binary("<=", Box.new(left), Box.new(right))
                },
                Token.Gt => {
                    self.advance()
                    let right = self.parse_term()?
                    left = Expr.Binary(">", Box.new(left), Box.new(right))
                },
                Token.Gte => {
                    self.advance()
                    let right = self.parse_term()?
                    left = Expr.Binary(">=", Box.new(left), Box.new(right))
                },
                Token.Eq => {
                    self.advance()
                    let right = self.parse_term()?
                    left = Expr.Binary("==", Box.new(left), Box.new(right))
                },
                Token.NotEq => {
                    self.advance()
                    let right = self.parse_term()?
                    left = Expr.Binary("!=", Box.new(left), Box.new(right))
                },
                _ => {
                    break
                },
            }
        }
        return Result.Ok(left)
    }

    fn parse_expr(mut self) -> Result[Expr, str] {
        return self.parse_comparison()
    }

    fn parse_block(mut self) -> Result[[Stmt], str] {
        let open = self.advance()
        match open {
            Token.LBrace => {},
            _ => { return Result.Err("expected '{' at start of block") },
        }

        let mut stmts: [Stmt] = []
        while self.has_more() {
            let check_tok = self.peek()
            match check_tok {
                Token.RBrace => {
                    self.advance()
                    break
                },
                Token.Eof => {
                    return Result.Err("unclosed '{' block before EOF")
                },
                _ => {
                    let s = self.parse_statement()?
                    stmts.push(s)
                },
            }
        }
        return Result.Ok(stmts)
    }

    fn parse_statement(mut self) -> Result[Stmt, str] {
        let tok = self.peek()

        match tok {
            Token.Let => {
                self.advance() // пропускаем let
                let mut is_mut = false
                let check_mut = self.peek()
                match check_mut {
                    Token.Mut => {
                        self.advance()
                        is_mut = true
                    },
                    _ => {},
                }

                let id_tok = self.advance()
                let mut name = ""
                match id_tok {
                    Token.Ident(id_name) => {
                        name = id_name
                    },
                    _ => {
                        return Result.Err("expected variable name after 'let'")
                    },
                }

                let mut type_ann = ""
                let check_colon = self.peek()
                match check_colon {
                    Token.Colon => {
                        self.advance()
                        let type_tok = self.advance()
                        match type_tok {
                            Token.Ident(t_name) => {
                                type_ann = t_name
                            },
                            _ => {
                                return Result.Err("expected type after ':'")
                            },
                        }
                    },
                    _ => {},
                }

                let eq_tok = self.advance()
                match eq_tok {
                    Token.Assign => {},
                    _ => {
                        return Result.Err("expected '=' in let statement")
                    },
                }

                let init_val = self.parse_expr()?
                if self.peek() == Token.Semicolon {
                    self.advance()
                }
                return Result.Ok(Stmt.Let(name, is_mut, type_ann, init_val))
            },

            Token.Return => {
                self.advance() // пропускаем return
                let ret_val = self.parse_expr()?
                if self.peek() == Token.Semicolon {
                    self.advance()
                }
                return Result.Ok(Stmt.Return(ret_val))
            },

            Token.If => {
                self.advance() // пропускаем if
                let cond = self.parse_expr()?
                let then_body = self.parse_block()?
                let mut else_body: [Stmt] = []
                let check_else = self.peek()
                match check_else {
                    Token.Else => {
                        self.advance()
                        else_body = self.parse_block()?
                    },
                    _ => {},
                }
                return Result.Ok(Stmt.If(cond, then_body, else_body))
            },

            Token.While => {
                self.advance() // пропускаем while
                let cond = self.parse_expr()?
                let body = self.parse_block()?
                return Result.Ok(Stmt.While(cond, body))
            },

            Token.For => {
                self.advance() // пропускаем for
                let item_tok = self.advance()
                let mut item_name = ""
                match item_tok {
                    Token.Ident(n) => { item_name = n },
                    _ => { return Result.Err("expected loop variable name in for-in") },
                }
                let in_tok = self.advance()
                match in_tok {
                    Token.In => {},
                    _ => { return Result.Err("expected 'in' after for variable") },
                }
                let start_tok = self.advance()
                let mut start_val = 0
                match start_tok {
                    Token.IntLit(n) => { start_val = n },
                    _ => { return Result.Err("expected integer start range") },
                }
                let dotdot = self.advance()
                match dotdot {
                    Token.DotDot => {},
                    _ => { return Result.Err("expected '..' in range") },
                }
                let end_tok = self.advance()
                let mut end_val = 0
                match end_tok {
                    Token.IntLit(n) => { end_val = n },
                    _ => { return Result.Err("expected integer end range") },
                }
                let body = self.parse_block()?
                return Result.Ok(Stmt.ForIn(item_name, start_val, end_val, body))
            },

            Token.Fn => {
                self.advance() // пропускаем fn
                let fn_name_tok = self.advance()
                let mut fn_name = ""
                match fn_name_tok {
                    Token.Ident(n) => { fn_name = n },
                    _ => { return Result.Err("expected function name after 'fn'") },
                }
                let lparen = self.advance()
                match lparen {
                    Token.LParen => {},
                    _ => { return Result.Err("expected '(' after function name") },
                }
                let mut params: [str] = []
                while self.has_more() {
                    let check_rp = self.peek()
                    match check_rp {
                        Token.RParen => {
                            self.advance()
                            break
                        },
                        _ => {},
                    }
                    let p_tok = self.advance()
                    match p_tok {
                        Token.Ident(pn) => {
                            // Опциональный тип : Type
                            if self.peek() == Token.Colon {
                                self.advance()
                                self.advance() // пропускаем тип
                            }
                            params.push(pn)
                        },
                        _ => { return Result.Err("expected parameter name") },
                    }
                    if self.peek() == Token.Comma {
                        self.advance()
                    }
                }

                let mut ret_type = "void"
                if self.peek() == Token.Arrow {
                    self.advance() // ->
                    let r_tok = self.advance()
                    match r_tok {
                        Token.Ident(rt) => { ret_type = rt },
                        _ => {},
                    }
                }

                let body = self.parse_block()?
                return Result.Ok(Stmt.FnDecl(fn_name, params, ret_type, body))
            },

            Token.Assert => {
                self.advance() // assert
                let lp = self.advance()
                match lp {
                    Token.LParen => {},
                    _ => { return Result.Err("expected '(' after assert") },
                }
                let cond = self.parse_expr()?
                let rp = self.advance()
                match rp {
                    Token.RParen => {},
                    _ => { return Result.Err("expected ')' after assert condition") },
                }
                if self.peek() == Token.Semicolon {
                    self.advance()
                }
                return Result.Ok(Stmt.Assert(cond))
            },

            Token.Struct => {
                self.advance() // struct
                let id_tok = self.advance()
                let mut struct_name = ""
                match id_tok {
                    Token.Ident(sname) => { struct_name = sname },
                    _ => { return Result.Err("expected struct name") },
                }
                match self.advance() {
                    Token.LBrace => {},
                    _ => { return Result.Err("expected '{' after struct name") },
                }
                let mut fields: [str] = []
                while self.has_more() && self.peek() != Token.RBrace {
                    let f_tok = self.advance()
                    match f_tok {
                        Token.Ident(fname) => {
                            if self.peek() == Token.Colon {
                                self.advance()
                                self.advance() // тип
                            }
                            fields.push(fname)
                        },
                        _ => {},
                    }
                    if self.peek() == Token.Comma {
                        self.advance()
                    }
                }
                match self.advance() {
                    Token.RBrace => {},
                    _ => { return Result.Err("expected '}' at end of struct declaration") },
                }
                return Result.Ok(Stmt.StructDecl(struct_name, fields))
            },

            Token.Impl => {
                self.advance() // impl
                let id_tok = self.advance()
                let mut target_name = ""
                match id_tok {
                    Token.Ident(tname) => { target_name = tname },
                    _ => { return Result.Err("expected target struct name after 'impl'") },
                }
                match self.advance() {
                    Token.LBrace => {},
                    _ => { return Result.Err("expected '{' after impl target") },
                }
                let mut methods: [Stmt] = []
                while self.has_more() && self.peek() != Token.RBrace {
                    let m_stmt = self.parse_statement()?
                    methods.push(m_stmt)
                }
                match self.advance() {
                    Token.RBrace => {},
                    _ => { return Result.Err("expected '}' at end of impl block") },
                }
                return Result.Ok(Stmt.ImplBlock(target_name, methods))
            },

            _ => {
                // Присваивание или ExpressionStatement
                let expr = self.parse_expr()?
                let mut is_assign = false
                let mut op = "="
                let next_t = self.peek()
                match next_t {
                    Token.Assign => {
                        is_assign = true
                        op = "="
                        self.advance()
                    },
                    Token.PlusAssign => {
                        is_assign = true
                        op = "+="
                        self.advance()
                    },
                    Token.MinusAssign => {
                        is_assign = true
                        op = "-="
                        self.advance()
                    },
                    Token.MulAssign => {
                        is_assign = true
                        op = "*="
                        self.advance()
                    },
                    Token.DivAssign => {
                        is_assign = true
                        op = "/="
                        self.advance()
                    },
                    _ => {},
                }
                if is_assign {
                    let new_val = self.parse_expr()?
                    match expr {
                        Expr.Var(v_name) => {
                            if self.peek() == Token.Semicolon {
                                self.advance()
                            }
                            return Result.Ok(Stmt.Assign(v_name, op, new_val))
                        },
                        _ => {
                            return Result.Err("invalid assignment target")
                        },
                    }
                }
                if self.peek() == Token.Semicolon {
                    self.advance()
                }
                return Result.Ok(Stmt.ExprStmt(expr))
            },
        }
    }

    fn parse_program(mut self) -> Result[[Stmt], str] {
        let mut stmts: [Stmt] = []
        while self.has_more() {
            let tok = self.peek()
            match tok {
                Token.Eof => {
                    break
                },
                _ => {
                    let s = self.parse_statement()?
                    stmts.push(s)
                },
            }
        }
        return Result.Ok(stmts)
    }
}
