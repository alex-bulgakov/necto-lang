// compiler/codegen.nc
// Модуль генератора нативного C-кода (Codegen) для самохостингового компилятора Necto

import { Expr, Stmt } from "compiler/ast.nc"

struct Codegen {
    indent_level: int
}

impl Codegen {
    fn new() -> Codegen {
        return Codegen { indent_level: 0 }
    }

    fn emit_expr(mut self, e: Expr) -> str {
        match e {
            Expr.Num(n) => {
                return f"{n}"
            },
            Expr.StrVal(s) => {
                return f"\"{s}\""
            },
            Expr.BoolVal(b) => {
                if b { return "1" }
                return "0"
            },
            Expr.Var(v) => {
                return v
            },
            Expr.Binary(op, left, right) => {
                let l = self.emit_expr(left.unwrap())
                let r = self.emit_expr(right.unwrap())
                return f"({l} {op} {r})"
            },
            Expr.Unary(op, right) => {
                let r = self.emit_expr(right.unwrap())
                return f"({op}{r})"
            },
            Expr.Call(name, args) => {
                if name == "println" {
                    if args.len() == 1 {
                        let arg_code = self.emit_expr(args[0])
                        return f"printf(\"%lld\\n\", (long long)({arg_code}))"
                    }
                }
                if name == "print" {
                    if args.len() == 1 {
                        let arg_code = self.emit_expr(args[0])
                        return f"printf(\"%lld\", (long long)({arg_code}))"
                    }
                }
                let mut arg_strs: [str] = []
                for i in 0..args.len() {
                    arg_strs.push(self.emit_expr(args[i]))
                }
                let mut joined = ""
                for j in 0..arg_strs.len() {
                    if j > 0 { joined += ", " }
                    joined += arg_strs[j]
                }
                return f"{name}({joined})"
            },
            Expr.Dot(obj, field) => {
                let o = self.emit_expr(obj.unwrap())
                return f"{o}.{field}"
            },
            _ => {
                return "0"
            },
        }
    }

    fn emit_stmt(mut self, s: Stmt) -> str {
        match s {
            Stmt.Let(name, is_mut, type_ann, val) => {
                let v = self.emit_expr(val)
                return f"    long long {name} = {v};\n"
            },
            Stmt.Assign(target, op, val) => {
                let v = self.emit_expr(val)
                return f"    {target} {op} {v};\n"
            },
            Stmt.Return(val) => {
                let v = self.emit_expr(val)
                return f"    return {v};\n"
            },
            Stmt.ExprStmt(e) => {
                let v = self.emit_expr(e)
                return f"    {v};\n"
            },
            Stmt.If(cond, then_body, else_body) => {
                let c = self.emit_expr(cond)
                let mut out = "    if (" + c + ") {\n"
                for i in 0..then_body.len() {
                    out += self.emit_stmt(then_body[i])
                }
                out += "    }"
                if else_body.len() > 0 {
                    out += " else {\n"
                    for j in 0..else_body.len() {
                        out += self.emit_stmt(else_body[j])
                    }
                    out += "    }\n"
                } else {
                    out += "\n"
                }
                return out
            },
            Stmt.While(cond, body) => {
                let c = self.emit_expr(cond)
                let mut out = "    while (" + c + ") {\n"
                for i in 0..body.len() {
                    out += self.emit_stmt(body[i])
                }
                out += "    }\n"
                return out
            },
            Stmt.ForIn(item, start, end, body) => {
                let mut out = "    for (long long " + item + " = " + f"{start}" + "; " + item + " < " + f"{end}" + "; " + item + "++) {\n"
                for i in 0..body.len() {
                    out += self.emit_stmt(body[i])
                }
                out += "    }\n"
                return out
            },
            Stmt.FnDecl(name, params, ret_type, body) => {
                let mut ret_c = "long long"
                if ret_type == "void" {
                    ret_c = "void"
                }
                if name == "main" {
                    ret_c = "int"
                }

                let mut p_list = ""
                for i in 0..params.len() {
                    if i > 0 { p_list += ", " }
                    p_list += "long long " + params[i]
                }

                let mut out = ret_c + " " + name + "(" + p_list + ") {\n"
                for j in 0..body.len() {
                    out += self.emit_stmt(body[j])
                }
                if name == "main" {
                    out += "    return 0;\n"
                }
                out += "}\n\n"
                return out
            },
            Stmt.StructDecl(name, fields) => {
                let mut out = "typedef struct {\n"
                for i in 0..fields.len() {
                    out += "    long long " + fields[i] + ";\n"
                }
                out += "} " + name + ";\n\n"
                return out
            },
            Stmt.Assert(cond) => {
                let c = self.emit_expr(cond)
                return f"    if (!({c})) {{ fprintf(stderr, \"AssertionError failed\\n\"); return 1; }}\n"
            },
            _ => {
                return ""
            },
        }
    }

    fn emit_program(mut self, stmts: [Stmt]) -> str {
        let mut code = "/* Generated automatically by Necto Self-Hosted Compiler (Stage 1) */\n"
        code += "#include <stdio.h>\n"
        code += "#include <stdlib.h>\n"
        code += "#include <stdbool.h>\n"
        code += "#include <string.h>\n\n"

        // Сначала генерируем функции
        for i in 0..stmts.len() {
            code += self.emit_stmt(stmts[i])
        }

        return code
    }
}
