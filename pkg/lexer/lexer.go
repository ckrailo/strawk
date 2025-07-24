package lexer

import (
	"slices"

	"github.com/ahalbert/strawk/pkg/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	lineNum      int
	linePosition int

	ExpectRegex bool
}

func New(input string) *Lexer {
	l := &Lexer{input: input, lineNum: 1, linePosition: 1, ExpectRegex: false}
	l.readChar()
	return l
}

func (l *Lexer) newToken(tokenType token.TokenType, s string) token.Token {
	return token.Token{Type: tokenType, Literal: s, LineNum: l.lineNum, Position: l.linePosition}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]

		if l.ch == '\n' {
			l.lineNum++
			l.linePosition = 1
		}
	}
	l.linePosition++
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) BacktrackToChar(target byte) {
	if l.readPosition == 0 {
		return
	}

	l.readPosition -= 1
	l.position -= 1
	l.linePosition -= 1
	l.ch = l.input[l.position]

	for l.ch != target {
		if l.ch == '\n' {
			l.lineNum--
			tempPos := l.readPosition - 1
			lineCharCount := 0
			for tempPos > 0 && l.input[tempPos] != '\n' {
				lineCharCount++
				tempPos--
			}
			l.linePosition = lineCharCount
		}
		l.readPosition -= 1
		l.position -= 1
		l.linePosition -= 1
		l.ch = l.input[l.position]
	}
}

func (l *Lexer) peek(lookahead int) string {
	if l.readPosition+lookahead >= len(l.input) {
		return ""
	} else {
		return l.input[l.readPosition : l.readPosition+lookahead]
	}
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func (l *Lexer) readUntilChar(chars ...byte) string {
	position := l.position
	for !slices.Contains(chars, l.ch) && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readWhileChar(chars ...byte) string {
	position := l.position
	for slices.Contains(chars, l.ch) && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || '0' <= ch && ch <= '9' || ch == '_' || ch == '$' || ch == '@'
}

func (l *Lexer) readNumeric() string {
	position := l.position
	for isDigit(l.ch) && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isDigit(ch byte) bool {
	return ch == '.' || '0' <= ch && ch <= '9'
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	if l.ExpectRegex {
		l.readChar()
		return l.newToken(token.REGEX, l.readUntilChar('/'))
	}

	l.skipWhitespace()

	switch l.ch {
	case '/':
		tok = l.newToken(token.SLASH, "/")
	case '\\':
		tok = l.newToken(token.ESCAPED_SLASH, "\\")
	case '"':
		l.readChar()
		tok = l.newToken(token.STRING, l.readUntilChar('"'))
	case '\'':
		l.readChar()
		tok = l.newToken(token.STRING, l.readUntilChar('\''))
	case '`':
		l.readChar()
		tok = l.newToken(token.STRING, l.readUntilChar('`'))
	case '-':
		lookahead := l.peek(1)
		if lookahead == "=" {
			l.readChar()
			tok = l.newToken(token.ASSIGNPLUS, "-=")
		} else if lookahead == "-" {
			l.readChar()
			tok = l.newToken(token.DECREMENT, "--")
		} else {
			tok = l.newToken(token.MINUS, "-")
		}
	case '=':
		if l.peek(1) == "=" {
			l.readChar()
			tok = l.newToken(token.EQ, "==")
		} else {
			tok = l.newToken(token.ASSIGN, "=")
		}
	case '{':
		tok = l.newToken(token.LBRACE, "{")
	case '}':
		tok = l.newToken(token.RBRACE, "}")
	case '(':
		tok = l.newToken(token.LPAREN, "(")
	case ')':
		tok = l.newToken(token.RPAREN, ")")
	case ',':
		tok = l.newToken(token.COMMA, ",")
	case ':':
		tok = l.newToken(token.COLON, ":")
	case ';':
		tok = l.newToken(token.SEMICOLON, ";")
	case '+':
		lookahead := l.peek(1)
		if lookahead == "=" {
			l.readChar()
			tok = l.newToken(token.ASSIGNPLUS, "+=")
		} else if lookahead == "+" {
			l.readChar()
			tok = l.newToken(token.INCREMENT, "++")
		} else {
			tok = l.newToken(token.PLUS, "+")
		}
	case '*':
		tok = l.newToken(token.ASTERISK, "*")
	case 0:
		tok = l.newToken(token.EOF, "")
	default:
		if isDigit(l.ch) {
			tok = l.newToken(token.NUMBER, l.readNumeric())
		} else if isLetter(l.ch) {
			tok = l.newToken(token.IDENT, l.readIdentifier())
			tok.Type = token.LookupIdent(tok.Literal)
		} else {
			tok = l.newToken(token.ILLEGAL, string(l.ch))
		}
		return tok
	}

	l.readChar()
	return tok
}

func (l *Lexer) isInsideRegex() bool {
	var next bool
	var prev bool
	pos := l.readPosition
	for !slices.Contains([]byte{'\n', '/'}, l.input[pos]) && pos < len(l.input) {
		pos += 1
	}
	if l.input[pos] == '/' {
		next = true
	}
	pos = l.readPosition
	for !slices.Contains([]byte{'\n', '/'}, l.input[pos]) && pos < len(l.input) {
		pos -= 1
	}
	if l.input[pos] == '/' {
		prev = true
	}
	return next && prev
}

func (l *Lexer) skipWhitespace() {
	for l.ch == '#' || l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '#' {
			if l.isInsideRegex() {
				return
			}
			l.readUntilChar('\n')
		}
		l.readChar()
	}
}
