package parser

import (
	"fmt"
	"strconv"
	"strings"

	"necto/pkg/ast"
	"necto/pkg/lexer"
	"necto/pkg/token"
)

// Приоритеты операторов (Pratt Parsing)
const (
	_ int = iota
	LOWEST
	ASSIGN      // =, +=, -=, *=, /=
	OR          // ||
	AND         // &&
	EQUALS      // ==, !=
	LESSGREATER // >, >=, <, <=
	RANGE       // ..
	SUM         // +, -
	PRODUCT     // *, /, %
	PREFIX      // -X, !X
	CALL        // fn(x)
	INDEX       // arr[i]
	DOT         // obj.prop
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:       ASSIGN,
	token.PLUS_ASSIGN:  ASSIGN,
	token.MINUS_ASSIGN: ASSIGN,
	token.MUL_ASSIGN:   ASSIGN,
	token.DIV_ASSIGN:   ASSIGN,
	token.OR:           OR,
	token.AND:          AND,
	token.EQ:           EQUALS,
	token.NOT_EQ:       EQUALS,
	token.LT:           LESSGREATER,
	token.LTE:          LESSGREATER,
	token.GT:           LESSGREATER,
	token.GTE:          LESSGREATER,
	token.DOTDOT:       RANGE,
	token.PLUS:         SUM,
	token.MINUS:        SUM,
	token.SLASH:        PRODUCT,
	token.ASTERISK:     PRODUCT,
	token.PERCENT:      PRODUCT,
	token.LPAREN:       CALL,
	token.LBRACKET:     INDEX,
	token.DOT:          DOT,
	token.QUESTION:     CALL,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifierOrStructLiteral)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.FSTRING, p.parseFStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.SOME, p.parseSomeExpression)
	p.registerPrefix(token.NONE, p.parseNoneLiteral)
	p.registerPrefix(token.MATCH, p.parseMatchExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.PERCENT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.DOTDOT, p.parseRangeExpression)

	p.registerInfix(token.ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.PLUS_ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.MINUS_ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.MUL_ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.DIV_ASSIGN, p.parseAssignExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseDotExpression)
	p.registerInfix(token.QUESTION, p.parseTryExpression)

	// Читаем два первых токена для инициализации curToken и peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("[%d:%d] expected next token to be %s, got %s instead",
		p.peekToken.Pos.Line, p.peekToken.Pos.Col, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ParseProgram - точка входа парсера
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForInStatement()
	case token.FN:
		return p.parseFnDeclaration()
	case token.STRUCT:
		return p.parseStructDeclaration()
	case token.IMPL:
		return p.parseImplBlockStatement()
	case token.ENUM:
		return p.parseEnumDeclaration()
	case token.IMPORT:
		return p.parseImportStatement()
	case token.TEST:
		return p.parseTestBlockStatement()
	case token.ASSERT:
		return p.parseAssertStatement()
	case token.BREAK:
		return &ast.BreakStatement{Token: p.curToken}
	case token.CONTINUE:
		return &ast.ContinueStatement{Token: p.curToken}
	case token.SEMICOLON:
		return nil
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if p.peekTokenIs(token.MUT) {
		p.nextToken()
		stmt.IsMut = true
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Опциональная аннотация типа: let x: int = ...
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // пропускаем ':'
		stmt.TypeAnnotation = p.parseTypeSignature()
	}

	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken() // пропускаем '='
		p.nextToken() // перемещаемся к выражению
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	if p.peekTokenIs(token.SEMICOLON) || p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) {
		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		return stmt
	}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseForInStatement() *ast.ForInStatement {
	stmt := &ast.ForInStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Item = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IN) {
		return nil
	}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseFnDeclaration() *ast.FnDeclaration {
	stmt := &ast.FnDeclaration{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Опциональные generic параметры: fn foo[T, U](...)
	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken() // [
		for !p.peekTokenIs(token.RBRACKET) && !p.peekTokenIs(token.EOF) {
			p.nextToken()
		}
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFnParameters()

	// Опциональный возвращаемый тип -> RetType
	if p.peekTokenIs(token.ARROW) {
		p.nextToken() // skip ->
		stmt.ReturnType = p.parseTypeSignature()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseFnParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	isMut := false
	if p.curTokenIs(token.MUT) {
		isMut = true
		p.nextToken()
	}

	param := &ast.Parameter{
		Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		IsMut: isMut,
	}
	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		param.Type = p.parseTypeSignature()
	}
	params = append(params, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		pMut := false
		if p.curTokenIs(token.MUT) {
			pMut = true
			p.nextToken()
		}
		pParam := &ast.Parameter{
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
			IsMut: pMut,
		}
		if p.peekTokenIs(token.COLON) {
			p.nextToken()
			pParam.Type = p.parseTypeSignature()
		}
		params = append(params, pParam)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

func (p *Parser) parseTypeSignature() string {
	p.nextToken()

	if p.curTokenIs(token.LBRACKET) {
		inner := p.parseTypeSignature()
		if !p.expectPeek(token.RBRACKET) {
			return "[" + inner + "]"
		}
		return "[" + inner + "]"
	}

	typeStr := p.curToken.Literal
	// Поддержка Option[T], Map[K, V]
	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken() // [
		inner := p.parseTypeSignature()
		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // ,
			inner2 := p.parseTypeSignature()
			if !p.expectPeek(token.RBRACKET) {
				return typeStr
			}
			return typeStr + "[" + inner + ", " + inner2 + "]"
		}
		if !p.expectPeek(token.RBRACKET) {
			return typeStr
		}
		return typeStr + "[" + inner + "]"
	}
	return typeStr
}

func (p *Parser) parseStructDeclaration() *ast.StructDeclaration {
	stmt := &ast.StructDeclaration{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Опциональные generic параметры: struct Stack[T]
	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken() // [
		for !p.peekTokenIs(token.RBRACKET) && !p.peekTokenIs(token.EOF) {
			p.nextToken()
		}
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}
		p.nextToken()
		fieldName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		fieldType := ""
		if p.peekTokenIs(token.COLON) {
			p.nextToken()
			fieldType = p.parseTypeSignature()
		}
		stmt.Fields = append(stmt.Fields, &ast.StructFieldDecl{
			Name: fieldName,
			Type: fieldType,
		})
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return stmt
}

func (p *Parser) parseImplBlockStatement() *ast.ImplBlockStatement {
	stmt := &ast.ImplBlockStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Target = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		if p.peekTokenIs(token.FN) {
			p.nextToken()
			fn := p.parseFnDeclaration()
			if fn != nil {
				stmt.Methods = append(stmt.Methods, fn)
			}
		} else {
			p.nextToken()
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return stmt
}

func (p *Parser) parseEnumDeclaration() *ast.EnumDeclaration {
	stmt := &ast.EnumDeclaration{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}
		p.nextToken()
		varName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		varTypes := []string{}

		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // (
			if !p.peekTokenIs(token.RPAREN) {
				p.nextToken()
				if p.peekTokenIs(token.COLON) {
					p.nextToken() // :
					varTypes = append(varTypes, p.parseTypeSignature())
				} else {
					varTypes = append(varTypes, p.parseTypeFromCurrentToken())
				}

				for p.peekTokenIs(token.COMMA) {
					p.nextToken() // ,
					p.nextToken()
					if p.peekTokenIs(token.COLON) {
						p.nextToken() // :
						varTypes = append(varTypes, p.parseTypeSignature())
					} else {
						varTypes = append(varTypes, p.parseTypeFromCurrentToken())
					}
				}
			}
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}

		stmt.Variants = append(stmt.Variants, &ast.EnumVariantDecl{
			Name:  varName,
			Types: varTypes,
		})
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return stmt
}

func (p *Parser) parseTypeFromCurrentToken() string {
	if p.curTokenIs(token.LBRACKET) {
		inner := p.parseTypeSignature()
		if !p.expectPeek(token.RBRACKET) {
			return "[" + inner + "]"
		}
		return "[" + inner + "]"
	}

	typeStr := p.curToken.Literal
	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken() // [
		inner := p.parseTypeSignature()
		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // ,
			inner2 := p.parseTypeSignature()
			if !p.expectPeek(token.RBRACKET) {
				return typeStr
			}
			return typeStr + "[" + inner + ", " + inner2 + "]"
		}
		if !p.expectPeek(token.RBRACKET) {
			return typeStr
		}
		return typeStr + "[" + inner + "]"
	}
	return typeStr
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			continue
		}
		p.nextToken()
		stmt.Symbols = append(stmt.Symbols, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	if !p.expectPeek(token.FROM) {
		return nil
	}

	if !p.expectPeek(token.STRING) {
		return nil
	}

	stmt.Path = p.curToken.Literal
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseTestBlockStatement() *ast.TestBlockStatement {
	stmt := &ast.TestBlockStatement{Token: p.curToken}

	if !p.expectPeek(token.STRING) {
		return nil
	}
	stmt.Name = p.curToken.Literal

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseAssertStatement() *ast.AssertStatement {
	stmt := &ast.AssertStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("[%d:%d] no prefix parse function for %s (literal %q)",
		p.curToken.Pos.Line, p.curToken.Pos.Col, t, p.curToken.Literal)
	p.errors = append(p.errors, msg)
}

// Prefix parsers

func (p *Parser) parseIdentifierOrStructLiteral() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Проверяем, не является ли это литералом структуры: MyStruct { field: val, ... }
	// Но только если следующая конструкция похожа на инициализатор структуры (например, { field: )
	if p.peekTokenIs(token.LBRACE) {
		// Заглядываем глубже через лексер, чтобы различить блок операторов от struct literal
		// Простой эвристикой: если внутри идентификатор и двоеточие
		return p.tryParseStructLiteral(ident)
	}

	return ident
}

func (p *Parser) tryParseStructLiteral(ident *ast.Identifier) ast.Expression {
	// Сохраняем состояние лексера/парсера нельзя так просто откатить в Go без снимка,
	// поэтому смотрим peekToken. Если сразу за { идет IDENT и затем COLON (или сразу }), это StructLiteral
	// В противном случае возвращаем обычный ident.
	// Для надежности структуры с инициализацией парсим, если это имя с заглавной буквы (конвенция) или если явно { name:
	if isUpper(ident.Value) {
		p.nextToken() // переходим на {
		lit := &ast.StructLiteral{
			Token:      ident.Token,
			StructName: ident,
			Fields:     []ast.StructFieldInit{},
		}

		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return lit
		}

		for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
			p.nextToken()
			fieldName := p.curToken.Literal
			if !p.expectPeek(token.COLON) {
				return nil
			}
			p.nextToken()
			fieldVal := p.parseExpression(LOWEST)
			lit.Fields = append(lit.Fields, ast.StructFieldInit{
				Name:  fieldName,
				Value: fieldVal,
			})
			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
			}
		}

		if !p.expectPeek(token.RBRACE) {
			return nil
		}
		return lit
	}

	return ident
}

func isUpper(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] >= 'A' && s[0] <= 'Z'
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	var val int64
	var err error
	str := p.curToken.Literal

	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		val, err = strconv.ParseInt(str[2:], 16, 64)
	} else if strings.HasPrefix(str, "0b") || strings.HasPrefix(str, "0B") {
		val, err = strconv.ParseInt(str[2:], 2, 64)
	} else {
		val, err = strconv.ParseInt(str, 10, 64)
	}

	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("[%d:%d] could not parse %q as integer",
			p.curToken.Pos.Line, p.curToken.Pos.Col, p.curToken.Literal))
		return nil
	}

	lit.Value = val
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("[%d:%d] could not parse %q as float",
			p.curToken.Pos.Line, p.curToken.Pos.Col, p.curToken.Literal))
		return nil
	}
	lit.Value = val
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseFStringLiteral() ast.Expression {
	lit := &ast.FStringLiteral{Token: p.curToken}
	content := p.curToken.Literal

	// Разбираем строку с подстановками {expr}
	var parts []ast.Expression
	var curText strings.Builder

	i := 0
	for i < len(content) {
		if content[i] == '{' {
			if curText.Len() > 0 {
				parts = append(parts, &ast.StringLiteral{
					Token: p.curToken,
					Value: curText.String(),
				})
				curText.Reset()
			}
			// Находим закрывающую фигурную скобку
			closeIdx := strings.Index(content[i:], "}")
			if closeIdx == -1 {
				p.errors = append(p.errors, fmt.Sprintf("[%d:%d] unclosed '{' in f-string", p.curToken.Pos.Line, p.curToken.Pos.Col))
				break
			}
			exprStr := content[i+1 : i+closeIdx]
			subLex := lexer.New(exprStr)
			subParser := New(subLex)
			subExpr := subParser.parseExpression(LOWEST)
			if subExpr != nil {
				parts = append(parts, subExpr)
			}
			i += closeIdx + 1
		} else {
			curText.WriteByte(content[i])
			i++
		}
	}

	if curText.Len() > 0 {
		parts = append(parts, &ast.StringLiteral{
			Token: p.curToken,
			Value: curText.String(),
		})
	}

	lit.Parts = parts
	return lit
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if p.peekTokenIs(end) {
			break // Trailing comma
		}
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if p.peekTokenIs(token.IF) {
			p.nextToken()
			// else if: обертка во вложенный блок
			nestedIf := p.parseIfExpression()
			expression.Alternative = &ast.BlockStatement{
				Token:      expression.Token,
				Statements: []ast.Statement{&ast.ExpressionStatement{Expression: nestedIf}},
			}
			return expression
		}

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseSomeExpression() ast.Expression {
	expr := &ast.SomeExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expr.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expr
}

func (p *Parser) parseNoneLiteral() ast.Expression {
	return &ast.NoneLiteral{Token: p.curToken}
}

func (p *Parser) parseMatchExpression() ast.Expression {
	matchExp := &ast.MatchExpression{Token: p.curToken}

	p.nextToken()
	matchExp.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	matchExp.Cases = []*ast.MatchCase{}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		pattern := p.parseExpression(LOWEST)

		if !p.expectPeek(token.FATARROW) {
			return nil
		}

		p.nextToken()
		var body ast.Expression
		if p.curTokenIs(token.LBRACE) {
			body = p.parseBlockStatement()
		} else {
			body = p.parseExpression(LOWEST)
		}

		matchExp.Cases = append(matchExp.Cases, &ast.MatchCase{
			Pattern: pattern,
			Body:    body,
		})

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return matchExp
}

// Infix parsers

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	assign := &ast.AssignExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	p.nextToken()
	assign.Value = p.parseExpression(LOWEST)

	return assign
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	re := &ast.RangeExpression{
		Token: p.curToken,
		Start: left,
	}

	p.nextToken()
	re.End = p.parseExpression(RANGE)

	return re
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	exp := &ast.DotExpression{Token: p.curToken, Left: left}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	exp.Right = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return exp
}

func (p *Parser) parseTryExpression(left ast.Expression) ast.Expression {
	return &ast.TryExpression{Token: p.curToken, Right: left}
}
