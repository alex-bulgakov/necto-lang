package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"necto/pkg/token"
)

type Lexer struct {
	input        string
	position     int  // текущая позиция в input (указывает на ch)
	readPosition int  // текущая позиция чтения (после ch)
	ch           rune // текущий символ
	line         int  // номер строки (1-indexed)
	col          int  // номер колонки (1-indexed)
}

func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		l.position = l.readPosition
		l.readPosition++
		l.col++
	} else {
		r, width := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.position = l.readPosition
		l.readPosition += width
		if r == '\n' {
			l.line++
			l.col = 0
		} else {
			l.col++
		}
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespaceAndComments()

	curPos := token.Pos{Line: l.line, Col: l.col}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.FATARROW, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.ASSIGN, l.ch, curPos)
		}
	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.PLUS_ASSIGN, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.PLUS, l.ch, curPos)
		}
	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ARROW, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.MINUS_ASSIGN, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.MINUS, l.ch, curPos)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.BANG, l.ch, curPos)
		}
	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.DIV_ASSIGN, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.SLASH, l.ch, curPos)
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.MUL_ASSIGN, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.ASTERISK, l.ch, curPos)
		}
	case '%':
		tok = newToken(token.PERCENT, l.ch, curPos)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.LT, l.ch, curPos)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.GT, l.ch, curPos)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.ILLEGAL, l.ch, curPos)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.ILLEGAL, l.ch, curPos)
		}
	case '?':
		tok = newToken(token.QUESTION, l.ch, curPos)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, curPos)
	case ':':
		tok = newToken(token.COLON, l.ch, curPos)
	case ',':
		tok = newToken(token.COMMA, l.ch, curPos)
	case '.':
		if l.peekChar() == '.' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.DOTDOT, Literal: string(ch) + string(l.ch), Pos: curPos}
		} else {
			tok = newToken(token.DOT, l.ch, curPos)
		}
	case '(':
		tok = newToken(token.LPAREN, l.ch, curPos)
	case ')':
		tok = newToken(token.RPAREN, l.ch, curPos)
	case '{':
		tok = newToken(token.LBRACE, l.ch, curPos)
	case '}':
		tok = newToken(token.RBRACE, l.ch, curPos)
	case '[':
		tok = newToken(token.LBRACKET, l.ch, curPos)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, curPos)
	case '"':
		strVal, err := l.readString()
		if err != nil {
			tok = token.Token{Type: token.ILLEGAL, Literal: err.Error(), Pos: curPos}
		} else {
			tok = token.Token{Type: token.STRING, Literal: strVal, Pos: curPos}
		}
		return tok
	case 'f':
		if l.peekChar() == '"' {
			l.readChar() // пропускаем 'f'
			fstrVal, err := l.readString()
			if err != nil {
				tok = token.Token{Type: token.ILLEGAL, Literal: err.Error(), Pos: curPos}
			} else {
				tok = token.Token{Type: token.FSTRING, Literal: fstrVal, Pos: curPos}
			}
			return tok
		}
		// Обычный идентификатор, начинающийся с 'f'
		ident := l.readIdentifier()
		tokType := token.LookupIdent(ident)
		return token.Token{Type: tokType, Literal: ident, Pos: curPos}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Pos = curPos
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			tokType := token.LookupIdent(ident)
			return token.Token{Type: tokType, Literal: ident, Pos: curPos}
		} else if isDigit(l.ch) {
			lit, isFloat := l.readNumber()
			if isFloat {
				return token.Token{Type: token.FLOAT, Literal: lit, Pos: curPos}
			}
			return token.Token{Type: token.INT, Literal: lit, Pos: curPos}
		} else {
			tok = newToken(token.ILLEGAL, l.ch, curPos)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		// Пропускаем пробелы, табы, переносы
		for unicode.IsSpace(l.ch) {
			l.readChar()
		}

		// Проверяем комментарии
		if l.ch == '/' {
			if l.peekChar() == '/' {
				// Строчный комментарий //
				for l.ch != '\n' && l.ch != 0 {
					l.readChar()
				}
				continue
			} else if l.peekChar() == '*' {
				// Блочный комментарий /* ... */
				l.readChar() // skip /
				l.readChar() // skip *
				for {
					if l.ch == 0 {
						break
					}
					if l.ch == '*' && l.peekChar() == '/' {
						l.readChar() // skip *
						l.readChar() // skip /
						break
					}
					l.readChar()
				}
				continue
			}
		}
		break
	}
}

func (l *Lexer) readIdentifier() string {
	startPos := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPos:l.position]
}

func (l *Lexer) readNumber() (string, bool) {
	startPos := l.position
	isFloat := false

	// Поддержка 0x (hex), 0b (bin)
	if l.ch == '0' {
		next := l.peekChar()
		if next == 'x' || next == 'X' {
			l.readChar() // 0
			l.readChar() // x
			for isHexDigit(l.ch) {
				l.readChar()
			}
			return l.input[startPos:l.position], false
		} else if next == 'b' || next == 'B' {
			l.readChar() // 0
			l.readChar() // b
			for l.ch == '0' || l.ch == '1' {
				l.readChar()
			}
			return l.input[startPos:l.position], false
		}
	}

	for isDigit(l.ch) {
		l.readChar()
	}

	// Проверяем точку для float (но не ..)
	if l.ch == '.' && l.peekChar() != '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // skip .
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[startPos:l.position], isFloat
}

func (l *Lexer) readString() (string, error) {
	var sb strings.Builder
	l.readChar() // пропускаем открывающую кавычку

	for {
		if l.ch == '"' {
			l.readChar() // пропускаем закрывающую кавычку
			break
		}
		if l.ch == 0 {
			return "", fmt.Errorf("unterminated string literal at %d:%d", l.line, l.col)
		}
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(l.ch)
			}
		} else {
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}
	return sb.String(), nil
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') ||
		(ch >= 'a' && ch <= 'f') ||
		(ch >= 'A' && ch <= 'F')
}

func newToken(tokenType token.TokenType, ch rune, pos token.Pos) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Pos: pos}
}
