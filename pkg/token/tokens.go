package token

type TokenType string

type Token struct {
	Type     TokenType
	Literal  string
	LineNum  int
	Position int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	//Identfiers
	IDENT   = "IDENT"
	REGEX   = "REGEX"
	STRING  = "STRING"
	NUMBER  = "NUMBER"
	NEWLINE = "\n"
	COMMENT = "COMMENT"

	//symbols

	ESCAPED_SLASH = "\\"
	ASSIGN        = "="
	SEMICOLON     = ";"
	COMMA         = ","
	LBRACE        = "{"
	RBRACE        = "}"
	LBRACKET      = "["
	RBRACKET      = "]"

	BANG = "!"
	AND  = "&&"
	OR   = "||"

	ASSIGNPLUS     = "+="
	PLUS           = "+"
	INCREMENT      = "++"
	MINUS          = "-"
	ASSIGNMINUS    = "-="
	DECREMENT      = "--"
	ASTERISK       = "*"
	ASSIGNMULTIPLY = "*="
	SLASH          = "/"
	ASSIGNDIVIDE   = "/="
	MODULO         = "%"
	ASSIGNMODULO   = "%="
	EXPONENT       = "^"
	ASSIGNEXPONENT = "^="

	REGEXMATCH    = "~"
	NOTREGEXMATCH = "!~"

	TERNARY = "?"
	COLON   = ":"

	LT   = "<"
	GT   = ">"
	LTEQ = "<="
	GTEQ = ">="

	EQ     = "=="
	NOT_EQ = "!="

	LPAREN = "("
	RPAREN = ")"

	//Keywords
	DO       = "DO"
	WHILE    = "WHILE"
	FOR      = "FOR"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	BEGIN    = "BEGIN"
	END      = "END"
	IN       = "IN"
	PRINT    = "PRINT"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	FUNCTION = "FUNCTION"
)

var keywords = map[string]TokenType{
	"do":       DO,
	"while":    WHILE,
	"for":      FOR,
	"break":    BREAK,
	"continue": CONTINUE,
	"in":       IN,
	"print":    PRINT,
	"BEGIN":    BEGIN,
	"END":      END,
	"if":       IF,
	"else":     ELSE,
	"function": FUNCTION,
	"return":   RETURN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
