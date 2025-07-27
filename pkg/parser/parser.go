package parser

import (
	"fmt"
	"strconv"

	"github.com/ahalbert/strawk/pkg/ast"
	"github.com/ahalbert/strawk/pkg/lexer"
	"github.com/ahalbert/strawk/pkg/token"
)

const (
	_ int = iota
	LOWEST
	BOOLEANLOGIC // && or ||
	REGEXMATCH   // ~ or !~
	MEMBERSHIP   // expr in array
	TERNARY      // condition ? a : b
	EQUALITY     // ==
	CONCATENATE  // implied
	SUM          // +
	PRODUCT      // *, /, %
	EXPONENT     // ^
	PREFIX       // -X or !X
	INDEX        // []
	CALL         // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.OR:            BOOLEANLOGIC,
	token.AND:           BOOLEANLOGIC,
	token.NOTREGEXMATCH: REGEXMATCH,
	token.REGEXMATCH:    REGEXMATCH,
	token.IN:            MEMBERSHIP,
	token.TERNARY:       TERNARY,
	token.EQ:            EQUALITY,
	token.NOT_EQ:        EQUALITY,
	token.LT:            EQUALITY,
	token.GT:            EQUALITY,
	token.LTEQ:          EQUALITY,
	token.GTEQ:          EQUALITY,
	token.PLUS:          SUM,
	token.MINUS:         SUM,
	token.ASTERISK:      PRODUCT,
	token.SLASH:         PRODUCT,
	token.MODULO:        PRODUCT,
	token.EXPONENT:      EXPONENT,
	token.LBRACKET:      INDEX,
	token.LPAREN:        CALL,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken       token.Token
	peekToken      token.Token
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifierExpr)
	p.registerPrefix(token.STRING, p.parseStringLiteralExpr)
	p.registerPrefix(token.NUMBER, p.parseNumericLiteralExpr)
	p.registerPrefix(token.SLASH, p.parseRegexExpression)
	p.registerPrefix(token.INCREMENT, p.parsePrefixExpression)
	p.registerPrefix(token.DECREMENT, p.parsePrefixExpression)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	// p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	// p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.EXPONENT, p.parseInfixExpression)

	p.registerInfix(token.REGEXMATCH, p.parseInfixExpression)
	p.registerInfix(token.NOTREGEXMATCH, p.parseInfixExpression)

	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)

	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTEQ, p.parseInfixExpression)
	p.registerInfix(token.GTEQ, p.parseInfixExpression)
	p.registerInfix(token.IN, p.parseArrayMembershipExpression)

	p.registerInfix(token.TERNARY, p.parseTernaryExpression)
	p.registerInfix(token.LBRACKET, p.parseArrayIndexExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) curTokenIs(tokens ...token.TokenType) bool {
	for _, t := range tokens {
		if p.curToken.Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) peekTokenIs(tokens ...token.TokenType) bool {
	for _, t := range tokens {
		if p.peekToken.Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.curToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.BEGIN:
		return p.parseBeginStatement()
	case token.END:
		return p.parseEndStatement()
	// case token.RETURN:
	// 	return p.parseReturnStatement()
	case token.PRINT:
		return p.parsePrintStatement()
	case token.NEWLINE:
		p.nextToken()
		return nil
	default:
		return p.parseExpressionPrefixedStatements()
	}
}

func (p *Parser) parseExpressionPrefixedStatements() ast.Statement {
	exprs := p.parseExpressionList(token.ASSIGN, token.LBRACE)

	switch p.curToken.Type {
	case token.ASSIGN:
		return p.parseAssignStatement(exprs)
	case token.ASSIGNPLUS:
		return p.parseAssignAndModifyStatement(exprs)
	case token.LBRACE:
		return p.parseActionBlockStatement(exprs)
	default:
		return &ast.ExpressionStatement{Token: p.curToken, Expressions: exprs}
	}
}

func (p *Parser) parseBeginStatement() *ast.BeginStatement {
	block := &ast.BeginStatement{Token: p.curToken}

	p.nextToken()
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	p.nextToken()

	return block
}

func (p *Parser) parseEndStatement() *ast.EndStatement {
	block := &ast.EndStatement{Token: p.curToken}

	p.nextToken()
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	return block
}

func (p *Parser) parseAssignStatement(targets []ast.Expression) *ast.AssignStatement {
	stmt := &ast.AssignStatement{Token: p.curToken}
	if !p.curTokenIs(token.ASSIGN) {
		return nil
	}

	//Convert previously parsed expressions to identifiers
	for _, expr := range targets {
		switch expr.(type) {
		case *ast.Identifier:
			stmt.Targets = append(stmt.Targets, expr)
		case *ast.ArrayIndexExpression:
			stmt.Targets = append(stmt.Targets, expr)
		default:
			panic("found non-identifier expression on lhs of assign statement")
		}
	}

	p.nextToken()

	stmt.Values = p.parseExpressionList(token.SEMICOLON)

	return stmt
}

func (p *Parser) parseAssignAndModifyStatement(targets []ast.Expression) *ast.AssignAndModifyStatement {
	if len(targets) != 1 {
		panic(p.curToken.Literal + " should have exactly 1 target")
	}

	var targetIdent *ast.Identifier

	operator := p.curToken
	target := targets[0]
	switch target.(type) {
	case *ast.Identifier:
		targetIdent = target.(*ast.Identifier)
	default:
		panic("found non-identifier expression on lhs of assign statement")
	}

	p.nextToken()

	return &ast.AssignAndModifyStatement{Operator: operator, Target: targetIdent, Value: p.parseExpression(LOWEST)}
}

func (p *Parser) parseActionBlockStatement(conditions []ast.Expression) *ast.ActionBlockStatement {

	if len(conditions) != 1 {
		panic("Action block should have exactly 1 condition")
	}

	stmt := &ast.ActionBlockStatement{Conditon: conditions[0]}

	//If a regex literal by itself, expand to $0 ~ /regex/
	switch stmt.Conditon.(type) {
	case *ast.RegexLiteral:
		stmt.Conditon = &ast.InfixExpression{Left: &ast.Identifier{Value: "$0"}, Operator: "~$0", Right: stmt.Conditon}
	}

	if !p.curTokenIs(token.LBRACE) {
		panic("Expected {")
	}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) {
		s := p.parseStatement()
		if s != nil {
			stmt.Statements = append(stmt.Statements, s)
		}
	}

	p.nextToken()

	return stmt
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken}
	p.nextToken()
	stmt.Expressions = p.parseExpressionList(token.SEMICOLON)
	return stmt
}

func (p *Parser) parseExpressionList(end ...token.TokenType) []ast.Expression {

	list := []ast.Expression{}

	if p.curTokenIs(end...) {
		p.nextToken()
		return nil
	}

	list = append(list, p.parseExpression(LOWEST))

	for p.curTokenIs(token.COMMA) {
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	return list
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	if !p.curTokenIs(token.SEMICOLON, token.COMMA, token.ASSIGN, token.NEWLINE, token.ASSIGNPLUS, token.LBRACE, token.COLON, token.RBRACKET) && precedence < p.curPrecedence() {
		for !p.curTokenIs(token.SEMICOLON, token.COMMA, token.ASSIGN, token.NEWLINE, token.ASSIGNPLUS, token.LBRACE, token.COLON, token.RBRACKET) && precedence < p.curPrecedence() {
			infix := p.infixParseFns[p.curToken.Type]
			if infix == nil {
				return leftExp
			}
			leftExp = infix(leftExp)
		}
		return leftExp
	}

	_, pok := p.prefixParseFns[p.curToken.Type]
	if pok && precedence <= p.curPrecedence() {
		return p.parseConcatenateExpression(leftExp)
	}
	return leftExp
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseIdentifierExpr() ast.Expression {
	defer p.nextToken()
	if p.curToken.Literal == "false" {
		return &ast.Boolean{Value: false}
	}
	if p.curToken.Literal == "true" {
		return &ast.Boolean{Value: true}
	}
	ident := &ast.Identifier{Value: p.curToken.Literal}
	if p.peekTokenIs(token.INCREMENT) {
		expr := &ast.PostfixExpression{Left: ident, Operator: p.peekToken.Literal}
		p.nextToken()
		return expr
	} else if p.peekTokenIs(token.DECREMENT) {
		expr := &ast.PostfixExpression{Left: ident, Operator: p.peekToken.Literal}
		p.nextToken()
		return expr
	}
	return ident
}

func (p *Parser) parseStringLiteralExpr() ast.Expression {
	lit := &ast.StringLiteral{Value: p.curToken.Literal}
	p.nextToken()
	return lit
}

func (p *Parser) parseNumericLiteralExpr() ast.Expression {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		panic("unparsable numeric type")
	}
	lit := &ast.NumericLiteral{Value: val}
	p.nextToken()
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Operator: p.curToken.Literal}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseConcatenateExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: ".",
		Left:     left,
	}

	precedence := CONCATENATE
	expression.Right = p.parseExpression(precedence)

	return expression
}

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

func (p *Parser) parseRegexExpression() ast.Expression {
	if !p.curTokenIs(token.SLASH) {
		return nil
	}

	p.l.ExpectRegex = true
	p.l.BacktrackToChar('/')
	p.nextToken()

	regex := p.peekToken.Literal

	p.l.ExpectRegex = false

	for !p.curTokenIs(token.SLASH) {
		p.nextToken()
	}

	p.nextToken()

	return &ast.RegexLiteral{Value: regex}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	p.nextToken()
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	p.nextToken()
	return exp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exprs := p.parseExpressionList(token.RPAREN)
	var expr ast.Expression
	if len(exprs) > 1 {
		expr = &ast.ArrayIndexExpression{IndexList: exprs}
	} else {
		expr = exprs[0]
	}
	if !p.curTokenIs(token.RPAREN) {
		panic("expected (")
		// p.addError(fmt.Sprintf("expected ), got %s %s", p.curToken.Type, p.curToken.Literal))
	}
	p.nextToken()
	return expr
}

func (p *Parser) parseTernaryExpression(expr ast.Expression) ast.Expression {
	ternaryExpr := &ast.TernaryExpression{Condition: expr}
	p.nextToken()
	ternaryExpr.IfTrue = p.parseExpression(LOWEST)
	if !p.curTokenIs(token.COLON) {
		panic("expected :")
		// p.addError(fmt.Sprintf("expected ), got %s %s", p.curToken.Type, p.curToken.Literal))
	}
	p.nextToken()
	ternaryExpr.IfFalse = p.parseExpression(LOWEST)
	return ternaryExpr
}

func (p *Parser) parseArrayIndexExpression(expr ast.Expression) ast.Expression {
	var id string
	switch expr.(type) {
	case *ast.Identifier:
		id = expr.String()
	default:
		panic("Attempt to address array with non-identifier")
	}
	p.nextToken()
	indicies := p.parseExpressionList()
	p.nextToken()
	return &ast.ArrayIndexExpression{ArrayName: id, IndexList: indicies}
}

func (p *Parser) parseArrayMembershipExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	p.nextToken()
	if !p.curTokenIs(token.IDENT) {
		panic("parse error in key in array expression: expected identifier on the right.")
	}
	right := p.parseExpression(LOWEST)
	expr.Right = right
	return expr
}
