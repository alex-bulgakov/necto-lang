package token

type TokenType string

type Pos struct {
	Line int
	Col  int
}

type Token struct {
	Type    TokenType
	Literal string
	Pos     Pos
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Идентификаторы и литералы
	IDENT   = "IDENT"
	INT     = "INT"
	FLOAT   = "FLOAT"
	STRING  = "STRING"
	FSTRING = "FSTRING"

	// Операторы
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	PERCENT  = "%"

	EQ     = "=="
	NOT_EQ = "!="
	LT     = "<"
	LTE    = "<="
	GT     = ">"
	GTE    = ">="

	AND = "&&"
	OR  = "||"

	QUESTION = "?"
	ARROW    = "->" // для сигнатур типов
	FATARROW = "=>" // для match/лямбд

	PLUS_ASSIGN  = "+="
	MINUS_ASSIGN = "-="
	MUL_ASSIGN   = "*="
	DIV_ASSIGN   = "/="

	// Разделители
	COMMA     = ","
	COLON     = ":"
	SEMICOLON = ";"
	DOT       = "."
	DOTDOT    = ".."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Ключевые слова
	FN       = "FN"
	LET      = "LET"
	MUT      = "MUT"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	WHILE    = "WHILE"
	FOR      = "FOR"
	IN       = "IN"
	STRUCT   = "STRUCT"
	ENUM     = "ENUM"
	MATCH    = "MATCH"
	TYPE     = "TYPE"
	COMPTIME = "COMPTIME"
	IMPORT   = "IMPORT"
	PACKAGE  = "PACKAGE"
	PUB      = "PUB"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	SOME     = "SOME"
	NONE     = "NONE"
	TEST     = "TEST"
	ASSERT   = "ASSERT"
	FROM     = "FROM"
	IMPL     = "IMPL"
	EXTERN   = "EXTERN"
	BENCH    = "BENCH"
)

var keywords = map[string]TokenType{
	"fn":       FN,
	"let":      LET,
	"mut":      MUT,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"while":    WHILE,
	"for":      FOR,
	"in":       IN,
	"struct":   STRUCT,
	"enum":     ENUM,
	"match":    MATCH,
	"type":     TYPE,
	"comptime": COMPTIME,
	"import":   IMPORT,
	"package":  PACKAGE,
	"pub":      PUB,
	"break":    BREAK,
	"continue": CONTINUE,
	"Some":     SOME,
	"None":     NONE,
	"test":     TEST,
	"bench":    BENCH,
	"assert":   ASSERT,
	"from":     FROM,
	"impl":     IMPL,
	"extern":   EXTERN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
