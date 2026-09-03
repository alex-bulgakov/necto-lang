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

    fn emit_formatted_string_printf(self, s: str, newline: bool) -> str {
        let mut fmt_str = ""
        let mut args_list: [str] = []
        let mut i = 0
        let n = s.len()
        let mut has_interp = false

        while i < n {
            let ch = s.char_at(i)
            if ch == 123 {
                let mut j = i + 1
                let mut var_name = ""
                while j < n && s.char_at(j) != 125 {
                    var_name = var_name + s.sub(j, j + 1)
                    j += 1
                }
                if j < n && s.char_at(j) == 125 {
                    has_interp = true
                    fmt_str = fmt_str + "%lld"
                    args_list.push(f"(long long)({var_name})")
                    i = j + 1
                } else {
                    fmt_str = fmt_str + s.sub(i, i + 1)
                    i += 1
                }
            } else {
                fmt_str = fmt_str + s.sub(i, i + 1)
                i += 1
            }
        }

        if !has_interp {
            if newline {
                return f"printf(\"%s\\n\", \"{s}\")"
            }
            return f"printf(\"%s\", \"{s}\")"
        }

        let mut args_joined = ""
        let mut k = 0
        while k < args_list.len() {
            args_joined = args_joined + ", " + args_list[k]
            k += 1
        }

        if newline {
            return f"printf(\"{fmt_str}\\n\"{args_joined})"
        }
        return f"printf(\"{fmt_str}\"{args_joined})"
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
                        let arg = args[0]
                        match arg {
                            Expr.StrVal(s) => {
                                return self.emit_formatted_string_printf(s, true)
                            },
                            _ => {
                                let arg_code = self.emit_expr(arg)
                                return f"printf(\"%lld\\n\", (long long)({arg_code}))"
                            },
                        }
                    }
                }
                if name == "print" {
                    if args.len() == 1 {
                        let arg = args[0]
                        match arg {
                            Expr.StrVal(s) => {
                                return self.emit_formatted_string_printf(s, false)
                            },
                            _ => {
                                let arg_code = self.emit_expr(arg)
                                return f"printf(\"%lld\", (long long)({arg_code}))"
                            },
                        }
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
            Expr.MethodCall(obj, method, args) => {
                let o = self.emit_expr(obj.unwrap())
                if method == "push" && args.len() == 1 {
                    let v = self.emit_expr(args[0])
                    return f"necto_push_array({o}, (long long)({v}))"
                }
                if method == "len" && args.len() == 0 {
                    return f"(long long)({o}->len)"
                }
                let mut arg_strs: [str] = [o]
                for i in 0..args.len() {
                    arg_strs.push(self.emit_expr(args[i]))
                }
                let mut joined = ""
                for j in 0..arg_strs.len() {
                    if j > 0 { joined += ", " }
                    joined += arg_strs[j]
                }
                return f"{method}({joined})"
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
            Stmt.EnumDecl(name, variants) => {
                let mut out = "typedef enum {\n"
                for i in 0..variants.len() {
                    out += "    " + name + "_" + variants[i] + ",\n"
                }
                out += "} " + name + "Tag;\n\n"
                out += "typedef struct {\n    " + name + "Tag tag;\n    union {\n        long long int_val;\n        const char* str_val;\n    } payload;\n} " + name + ";\n\n"
                for j in 0..variants.len() {
                    let v = variants[j]
                    out += "static inline " + name + " " + name + "_" + v + "_val(long long val) {\n"
                    out += "    " + name + " e; e.tag = " + name + "_" + v + "; e.payload.int_val = val; return e;\n"
                    out += "}\n"
                }
                out += "\n"
                return out
            },
            Stmt.ImplBlock(target, methods) => {
                let mut out = ""
                for i in 0..methods.len() {
                    out += self.emit_stmt(methods[i])
                }
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
        code += "typedef struct { long long* data; size_t len; size_t cap; } NectoIntArray;\n"
        code += "static inline NectoIntArray* necto_new_array(size_t cap) {\n"
        code += "    if (cap == 0) cap = 4;\n"
        code += "    NectoIntArray* a = (NectoIntArray*)malloc(sizeof(NectoIntArray));\n"
        code += "    a->data = (long long*)malloc(sizeof(long long) * cap);\n"
        code += "    a->len = 0; a->cap = cap;\n"
        code += "    return a;\n"
        code += "}\n"
        code += "static inline void necto_push_array(NectoIntArray* a, long long v) {\n"
        code += "    if (a->len >= a->cap) { a->cap *= 2; a->data = (long long*)realloc(a->data, sizeof(long long) * a->cap); }\n"
        code += "    a->data[a->len++] = v;\n"
        code += "}\n\n"

        // Сначала генерируем функции
        for i in 0..stmts.len() {
            code += self.emit_stmt(stmts[i])
        }

        return code
    }
}
