package ast

import (
	"bytes"
	"fmt"
	"strings"

	"necto/pkg/token"
)

type Node interface {
	TokenLiteral() string
	String() string
	Pos() token.Pos
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// Program - корень синтаксического дерева
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) Pos() token.Pos {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return token.Pos{Line: 1, Col: 1}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

// --- Statements ---

type LetStatement struct {
	Token          token.Token // токен LET
	Name           *Identifier
	TypeAnnotation string // опциональный тип, например "int", "Option[str]"
	IsMut          bool
	Value          Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) Pos() token.Pos       { return ls.Token.Pos }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString("let ")
	if ls.IsMut {
		out.WriteString("mut ")
	}
	out.WriteString(ls.Name.String())
	if ls.TypeAnnotation != "" {
		out.WriteString(": " + ls.TypeAnnotation)
	}
	if ls.Value != nil {
		out.WriteString(" = ")
		out.WriteString(ls.Value.String())
	}
	return out.String()
}

type ReturnStatement struct {
	Token       token.Token // токен RETURN
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) Pos() token.Pos       { return rs.Token.Pos }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString("return")
	if rs.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(rs.ReturnValue.String())
	}
	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token // первый токен выражения
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) Pos() token.Pos       { return es.Token.Pos }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      token.Token // токен {
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) expressionNode()      {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) Pos() token.Pos       { return bs.Token.Pos }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{\n")
	for _, s := range bs.Statements {
		out.WriteString("  " + s.String() + "\n")
	}
	out.WriteString("}")
	return out.String()
}

type WhileStatement struct {
	Token     token.Token // токен WHILE
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) Pos() token.Pos       { return ws.Token.Pos }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString(" ")
	out.WriteString(ws.Body.String())
	return out.String()
}

type ForInStatement struct {
	Token    token.Token // токен FOR
	Item     *Identifier
	Iterable Expression // например, RangeExpression (0..10) или Array
	Body     *BlockStatement
}

func (fs *ForInStatement) statementNode()       {}
func (fs *ForInStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForInStatement) Pos() token.Pos       { return fs.Token.Pos }
func (fs *ForInStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for " + fs.Item.String() + " in " + fs.Iterable.String() + " ")
	out.WriteString(fs.Body.String())
	return out.String()
}

type Parameter struct {
	Name *Identifier
	Type string
}

type FnDeclaration struct {
	Token      token.Token // токен FN
	Name       *Identifier
	Parameters []*Parameter
	ReturnType string
	Body       *BlockStatement
}

func (fd *FnDeclaration) statementNode()       {}
func (fd *FnDeclaration) TokenLiteral() string { return fd.Token.Literal }
func (fd *FnDeclaration) Pos() token.Pos       { return fd.Token.Pos }
func (fd *FnDeclaration) String() string {
	var out bytes.Buffer
	params := make([]string, len(fd.Parameters))
	for i, p := range fd.Parameters {
		params[i] = p.Name.String() + ": " + p.Type
	}
	out.WriteString("fn " + fd.Name.String() + "(" + strings.Join(params, ", ") + ")")
	if fd.ReturnType != "" {
		out.WriteString(" -> " + fd.ReturnType)
	}
	out.WriteString(" ")
	out.WriteString(fd.Body.String())
	return out.String()
}

type StructFieldDecl struct {
	Name *Identifier
	Type string
}

type StructDeclaration struct {
	Token  token.Token // токен STRUCT
	Name   *Identifier
	Fields []*StructFieldDecl
}

func (sd *StructDeclaration) statementNode()       {}
func (sd *StructDeclaration) TokenLiteral() string { return sd.Token.Literal }
func (sd *StructDeclaration) Pos() token.Pos       { return sd.Token.Pos }
func (sd *StructDeclaration) String() string {
	var out bytes.Buffer
	out.WriteString("struct " + sd.Name.String() + " {\n")
	for _, f := range sd.Fields {
		out.WriteString("  " + f.Name.String() + ": " + f.Type + "\n")
	}
	out.WriteString("}")
	return out.String()
}

type EnumVariantDecl struct {
	Name  *Identifier
	Types []string // типы полезной нагрузки, e.g. Number(int) -> ["int"], Eof -> []
}

type EnumDeclaration struct {
	Token    token.Token // токен ENUM
	Name     *Identifier
	Variants []*EnumVariantDecl
}

func (ed *EnumDeclaration) statementNode()       {}
func (ed *EnumDeclaration) TokenLiteral() string { return ed.Token.Literal }
func (ed *EnumDeclaration) Pos() token.Pos       { return ed.Token.Pos }
func (ed *EnumDeclaration) String() string {
	var out bytes.Buffer
	out.WriteString("enum " + ed.Name.String() + " {\n")
	for _, v := range ed.Variants {
		out.WriteString("  " + v.Name.String())
		if len(v.Types) > 0 {
			out.WriteString("(" + strings.Join(v.Types, ", ") + ")")
		}
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

type ImportStatement struct {
	Token   token.Token // токен IMPORT
	Symbols []*Identifier
	Path    string // путь к файлу, e.g. "./math.nc"
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) Pos() token.Pos       { return is.Token.Pos }
func (is *ImportStatement) String() string {
	var syms []string
	for _, s := range is.Symbols {
		syms = append(syms, s.String())
	}
	return "import { " + strings.Join(syms, ", ") + " } from \"" + is.Path + "\""
}

type TestBlockStatement struct {
	Token token.Token // токен TEST
	Name  string      // название теста
	Body  *BlockStatement
}

func (tb *TestBlockStatement) statementNode()       {}
func (tb *TestBlockStatement) TokenLiteral() string { return tb.Token.Literal }
func (tb *TestBlockStatement) Pos() token.Pos       { return tb.Token.Pos }
func (tb *TestBlockStatement) String() string {
	return "test \"" + tb.Name + "\" " + tb.Body.String()
}

type AssertStatement struct {
	Token     token.Token // токен ASSERT
	Condition Expression
	Message   string
}

func (as *AssertStatement) statementNode()       {}
func (as *AssertStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssertStatement) Pos() token.Pos       { return as.Token.Pos }
func (as *AssertStatement) String() string {
	return "assert(" + as.Condition.String() + ")"
}

type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) Pos() token.Pos       { return bs.Token.Pos }
func (bs *BreakStatement) String() string       { return "break" }

type ContinueStatement struct {
	Token token.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) Pos() token.Pos       { return cs.Token.Pos }
func (cs *ContinueStatement) String() string       { return "continue" }

// --- Expressions ---

type Identifier struct {
	Token token.Token // токен IDENT
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) Pos() token.Pos       { return i.Token.Pos }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) Pos() token.Pos       { return il.Token.Pos }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) Pos() token.Pos       { return fl.Token.Pos }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) Pos() token.Pos       { return b.Token.Pos }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) Pos() token.Pos       { return sl.Token.Pos }
func (sl *StringLiteral) String() string       { return `"` + sl.Value + `"` }

type FStringLiteral struct {
	Token token.Token
	Parts []Expression // Конкатенируемые выражения (строковые константы и вычисляемые выражения)
}

func (fs *FStringLiteral) expressionNode()      {}
func (fs *FStringLiteral) TokenLiteral() string { return fs.Token.Literal }
func (fs *FStringLiteral) Pos() token.Pos       { return fs.Token.Pos }
func (fs *FStringLiteral) String() string       { return `f"` + fs.Token.Literal + `"` }

type PrefixExpression struct {
	Token    token.Token // например, ! или -
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) Pos() token.Pos       { return pe.Token.Pos }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

type InfixExpression struct {
	Token    token.Token // бинарный оператор (+, -, *, ==, &&, ...)
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) Pos() token.Pos       { return oe.Token.Pos }
func (oe *InfixExpression) String() string {
	return "(" + oe.Left.String() + " " + oe.Operator + " " + oe.Right.String() + ")"
}

type AssignExpression struct {
	Token    token.Token // =, +=, -=, etc.
	Left     Expression
	Operator string
	Value    Expression
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignExpression) Pos() token.Pos       { return ae.Token.Pos }
func (ae *AssignExpression) String() string {
	return ae.Left.String() + " " + ae.Operator + " " + ae.Value.String()
}

type IfExpression struct {
	Token       token.Token // токен IF
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) Pos() token.Pos       { return ie.Token.Pos }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if " + ie.Condition.String() + " " + ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString(" else " + ie.Alternative.String())
	}
	return out.String()
}

type CallExpression struct {
	Token     token.Token // токен '('
	Function  Expression  // Идентификатор или DotExpression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) Pos() token.Pos       { return ce.Token.Pos }
func (ce *CallExpression) String() string {
	args := make([]string, len(ce.Arguments))
	for i, a := range ce.Arguments {
		args[i] = a.String()
	}
	return ce.Function.String() + "(" + strings.Join(args, ", ") + ")"
}

type IndexExpression struct {
	Token token.Token // токен '['
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) Pos() token.Pos       { return ie.Token.Pos }
func (ie *IndexExpression) String() string {
	return ie.Left.String() + "[" + ie.Index.String() + "]"
}

type DotExpression struct {
	Token token.Token // токен '.'
	Left  Expression
	Right *Identifier
}

func (de *DotExpression) expressionNode()      {}
func (de *DotExpression) TokenLiteral() string { return de.Token.Literal }
func (de *DotExpression) Pos() token.Pos       { return de.Token.Pos }
func (de *DotExpression) String() string {
	return de.Left.String() + "." + de.Right.String()
}

type ArrayLiteral struct {
	Token    token.Token // токен '['
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) Pos() token.Pos       { return al.Token.Pos }
func (al *ArrayLiteral) String() string {
	elements := make([]string, len(al.Elements))
	for i, e := range al.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

type RangeExpression struct {
	Token token.Token // токен '..'
	Start Expression
	End   Expression
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RangeExpression) Pos() token.Pos       { return re.Token.Pos }
func (re *RangeExpression) String() string {
	return re.Start.String() + ".." + re.End.String()
}

type StructFieldInit struct {
	Name  string
	Value Expression
}

type StructLiteral struct {
	Token      token.Token // Имя структуры
	StructName *Identifier
	Fields     []StructFieldInit
}

func (sl *StructLiteral) expressionNode()      {}
func (sl *StructLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StructLiteral) Pos() token.Pos       { return sl.Token.Pos }
func (sl *StructLiteral) String() string {
	var out bytes.Buffer
	out.WriteString(sl.StructName.String() + " { ")
	for i, f := range sl.Fields {
		out.WriteString(f.Name + ": " + f.Value.String())
		if i < len(sl.Fields)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(" }")
	return out.String()
}

type SomeExpression struct {
	Token token.Token // токен SOME
	Value Expression
}

func (se *SomeExpression) expressionNode()      {}
func (se *SomeExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SomeExpression) Pos() token.Pos       { return se.Token.Pos }
func (se *SomeExpression) String() string {
	return "Some(" + se.Value.String() + ")"
}

type NoneLiteral struct {
	Token token.Token // токен NONE
}

func (nl *NoneLiteral) expressionNode()      {}
func (nl *NoneLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NoneLiteral) Pos() token.Pos       { return nl.Token.Pos }
func (nl *NoneLiteral) String() string       { return "None" }

type MatchCase struct {
	Pattern Expression // например, SomeExpression, NoneLiteral, Identifier, Literal
	Body    Expression
}

type MatchExpression struct {
	Token token.Token // токен MATCH
	Value Expression
	Cases []*MatchCase
}

func (me *MatchExpression) expressionNode()      {}
func (me *MatchExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MatchExpression) Pos() token.Pos       { return me.Token.Pos }
func (me *MatchExpression) String() string {
	var out bytes.Buffer
	out.WriteString("match " + me.Value.String() + " {\n")
	for _, c := range me.Cases {
		out.WriteString(fmt.Sprintf("  %s => %s,\n", c.Pattern.String(), c.Body.String()))
	}
	out.WriteString("}")
	return out.String()
}

type EnumConstructorExpr struct {
	Token       token.Token // токен IDENT (имя Enum)
	EnumName    string
	VariantName string
	Arguments   []Expression
}

func (ece *EnumConstructorExpr) expressionNode()      {}
func (ece *EnumConstructorExpr) TokenLiteral() string { return ece.Token.Literal }
func (ece *EnumConstructorExpr) Pos() token.Pos       { return ece.Token.Pos }
func (ece *EnumConstructorExpr) String() string {
	var out bytes.Buffer
	out.WriteString(ece.EnumName + "." + ece.VariantName)
	if len(ece.Arguments) > 0 {
		var args []string
		for _, a := range ece.Arguments {
			args = append(args, a.String())
		}
		out.WriteString("(" + strings.Join(args, ", ") + ")")
	}
	return out.String()
}
