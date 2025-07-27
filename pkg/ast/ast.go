package ast

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/ahalbert/strawk/pkg/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

type Block interface {
	GetStatements() []Statement
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type ExpressionStatement struct {
	Token       token.Token // the first token of the expression
	Expressions []Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	var out bytes.Buffer

	for _, exp := range es.Expressions {
		out.WriteString(exp.String())
	}

	return out.String()
}

type ActionBlockStatement struct {
	Token      token.Token // the { token
	Conditon   Expression
	Statements []Statement
}

func (as *ActionBlockStatement) statementNode()             {}
func (as *ActionBlockStatement) TokenLiteral() string       { return as.Token.Literal }
func (as *ActionBlockStatement) GetStatements() []Statement { return as.Statements }
func (as *ActionBlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range as.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type BeginStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BeginStatement) statementNode()             {}
func (bs *BeginStatement) TokenLiteral() string       { return bs.Token.Literal }
func (bs *BeginStatement) GetStatements() []Statement { return bs.Statements }
func (bs *BeginStatement) String() string {
	var out bytes.Buffer

	out.WriteString("BEGIN {\n")
	for _, s := range bs.Statements {
		out.WriteString(s.String() + ";\n")
	}
	out.WriteString("}\n")

	return out.String()
}

type EndStatement struct {
	Token      token.Token
	Statements []Statement
}

func (es *EndStatement) statementNode()             {}
func (es *EndStatement) TokenLiteral() string       { return es.Token.Literal }
func (es *EndStatement) GetStatements() []Statement { return es.Statements }
func (es *EndStatement) String() string {
	var out bytes.Buffer

	for _, s := range es.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type AssignStatement struct {
	Token   token.Token // the { token
	Targets []Expression
	Values  []Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) String() string {
	var out bytes.Buffer

	for idx, s := range as.Targets {
		out.WriteString(s.String() + " = " + as.Values[idx].String())
	}

	return out.String()
}

type AssignAndModifyStatement struct {
	Token    token.Token // the { token
	Operator token.Token
	Target   Expression
	Value    Expression
}

func (ams *AssignAndModifyStatement) statementNode()       {}
func (ams *AssignAndModifyStatement) TokenLiteral() string { return ams.Token.Literal }
func (ams *AssignAndModifyStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ams.Target.String() + " " + ams.Operator.Literal + " " + ams.Value.String())

	return out.String()
}

type PrintStatement struct {
	Token       token.Token // the print token
	Expressions []Expression
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PrintStatement) String() string {
	var out bytes.Buffer
	out.WriteString("print ")

	for idx, s := range ps.Expressions {
		out.WriteString(s.String())
		if idx < len(ps.Expressions) {
			out.WriteString(",")
		}
	}

	return out.String()
}

// Expressions

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type TernaryExpression struct {
	Token     token.Token // The '(' token
	Condition Expression
	IfTrue    Expression
	IfFalse   Expression
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString(te.Condition.String() + " ? " + te.IfTrue.String() + " : " + te.IfFalse.String())

	return out.String()
}

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type PostfixExpression struct {
	Token    token.Token
	Left     *Identifier
	Operator string
}

func (pe *PostfixExpression) expressionNode()      {}
func (pe *PostfixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PostfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Left.String())
	out.WriteString(pe.Operator)
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type NumericLiteral struct {
	Token token.Token
	Value float64
}

func (il *NumericLiteral) expressionNode()      {}
func (il *NumericLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *NumericLiteral) String() string {
	if il.Value == float64(int(il.Value)) {
		return fmt.Sprintf("%d", int(il.Value))
	}
	return fmt.Sprintf("%.5g", il.Value)
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Value }

func NewLiteral(val string) Expression {
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return &StringLiteral{Value: val}
	}
	return &NumericLiteral{Value: parsed}
}

type RegexLiteral struct {
	Token token.Token
	Value string
}

func (rl *RegexLiteral) expressionNode()      {}
func (rl *RegexLiteral) TokenLiteral() string { return rl.Token.Literal }
func (rl *RegexLiteral) String() string       { return rl.Value }

type AssociativeArray struct {
	Token token.Token
	Array map[string]Expression
}

func (aa *AssociativeArray) expressionNode()      {}
func (aa *AssociativeArray) TokenLiteral() string { return aa.Token.Literal }
func (aa *AssociativeArray) String() string {
	var out bytes.Buffer
	entries := []string{}
	for k, v := range aa.Array {
		entries = append(entries, k+" : "+v.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(entries, ", "))
	out.WriteString("}")

	return out.String()
}

type ArrayIndexExpression struct {
	Token     token.Token
	ArrayName string
	IndexList []Expression
}

func (aie *ArrayIndexExpression) expressionNode()      {}
func (aie *ArrayIndexExpression) TokenLiteral() string { return aie.Token.Literal }
func (aie *ArrayIndexExpression) String() string {
	var out bytes.Buffer
	indicies := []string{}

	for _, i := range aie.IndexList {
		indicies = append(indicies, i.String())
	}
	out.WriteString(aie.ArrayName + "[" + strings.Join(indicies, ", ") + "]")

	return out.String()
}
