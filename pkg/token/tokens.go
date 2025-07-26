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

	COLON         = ":"
	ESCAPED_SLASH = "\\"
	ASSIGN        = "="
	SEMICOLON     = ";"
	COMMA         = ","
	LBRACE        = "{"
	RBRACE        = "}"

	BANG           = "!"
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
	TILDE          = "~"

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
	DOUNTIL  = "DOUNTIL"
	BEGIN    = "BEGIN"
	END      = "END"
	CAPTURE  = "CAPTURE"
	LABEL    = "LABEL"
	LET      = "LET"
	PRINT    = "PRINT"
	PRINTLN  = "PRINTLN"
	CLEAR    = "CLEAR"
	REWIND   = "REWIND"
	FASTFWD  = "FASTFORWARD"
	PAUSE    = "PAUSE"
	PLAY     = "PLAY"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	FUNCTION = "FUNCTION"
)

var keywords = map[string]TokenType{
	"do":          DO,
	"dountil":     DOUNTIL,
	"capture":     CAPTURE,
	"print":       PRINT,
	"println":     PRINTLN,
	"BEGIN":       BEGIN,
	"END":         END,
	"clear":       CLEAR,
	"let":         LET,
	"rewind":      REWIND,
	"fastforward": FASTFWD,
	"pause":       PAUSE,
	"play":        PLAY,
	"if":          IF,
	"else":        ELSE,
	"function":    FUNCTION,
	"return":      RETURN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
