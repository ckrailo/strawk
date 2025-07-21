package interpreter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ahalbert/strawk/pkg/ast"
	"github.com/ahalbert/strawk/pkg/stdlib"
	"github.com/ahalbert/strawk/pkg/token"
)

type Interpreter struct {
	BeginBlock                   *ast.BeginStatement
	EndBlock                     *ast.EndStatement
	Rules                        []ast.Statement
	Program                      *ast.Program
	Input                        string
	InputPostion                 int
	Stack                        []CallStackEntry
	GlobalVariables              map[string]string
	StdLibFunctions              map[string]func([]ast.Expression) ast.Expression
	mostRecentRegexCaptureGroups map[string]string
}

type CallStackEntry struct {
	isFunction     bool
	LocalVariables map[string]string
}

func NewInterpreter(program *ast.Program, input string) *Interpreter {
	i := &Interpreter{Program: program, Input: input, GlobalVariables: make(map[string]string), StdLibFunctions: make(map[string]func([]ast.Expression) ast.Expression)}
	i.Stack = append(i.Stack, CallStackEntry{})
	i.Stack[0].LocalVariables = make(map[string]string)
	i.Stack[0].LocalVariables["$0"] = ""
	i.InputPostion = 0
	for _, stmt := range program.Statements {
		switch stmt.(type) {
		case *ast.BeginStatement:
			i.BeginBlock = stmt.(*ast.BeginStatement)
		case *ast.EndStatement:
			i.EndBlock = stmt.(*ast.EndStatement)
		default:
			i.Rules = append(i.Rules, stmt)
		}
	}

	i.StdLibFunctions["length"] = stdlib.Length
	return i
}

func (i *Interpreter) Run() {
	i.doBlock(i.BeginBlock)
	i.advanceInput()
	for i.InputPostion < len(i.Input) {
		for _, stmt := range i.Rules {
			i.doStatement(stmt)
		}
		i.advanceInput()
	}
	// i.doBlock(i.EndBlock)
}

func (i *Interpreter) advanceInput() {
	i.InputPostion += 1
	if i.InputPostion >= len(i.Input) {
		return
	}
	i.Stack[0].LocalVariables["$0"] += string(i.Input[i.InputPostion])
}

func (i *Interpreter) backtrackInput() {
	i.InputPostion -= 1
	i.Stack[0].LocalVariables["$0"] = i.Stack[0].LocalVariables["$0"][:len(i.Stack[0].LocalVariables["$0"])-1]
}

func (i *Interpreter) consumeInput() {
	i.Stack[0].LocalVariables["$0"] = ""
}

func (i *Interpreter) lookupVar(varName string) string {
	val, ok := i.Stack[len(i.Stack)-1].LocalVariables[varName]
	if ok {
		return val
	}
	val, ok = i.GlobalVariables[varName]
	if ok {
		return val
	}
	return ""
}

func (i *Interpreter) setVar(varName string, value string) {
	_, ok := i.Stack[len(i.Stack)-1].LocalVariables[varName]
	if ok {
		i.Stack[len(i.Stack)-1].LocalVariables[varName] = value
	} else {
		i.GlobalVariables[varName] = value
	}
}

func (i *Interpreter) createLocalVar(varName string, value string) {
	i.Stack[len(i.Stack)-1].LocalVariables[varName] = value
}

func (i *Interpreter) doStatement(stmt ast.Statement) {
	switch stmt.(type) {
	case *ast.ExpressionStatement:
		i.doExpressionList(stmt.(*ast.ExpressionStatement).Expressions)
	case *ast.PrintStatement:
		i.doPrintStatement(stmt.(*ast.PrintStatement))
	case *ast.ActionBlockStatement:
		i.doBlock(stmt.(*ast.ActionBlockStatement))
	case *ast.AssignStatement:
		i.doAssignStatement(stmt.(*ast.AssignStatement))
	case *ast.AssignAndModifyStatement:
		i.doAssignAndModifyStatement(stmt.(*ast.AssignAndModifyStatement))
	default:
		panic("Unexpected statement type")
	}
}

func (i *Interpreter) doBlock(block ast.Block) {
	i.mostRecentRegexCaptureGroups = make(map[string]string)
	shouldExecuteBlock := false
	switch block.(type) {
	case *ast.BeginStatement:
		shouldExecuteBlock = true
	case *ast.EndStatement:
		shouldExecuteBlock = true
	case *ast.ActionBlockStatement:
		shouldExecuteBlock = i.evaluateActionBlockConditon(block.(*ast.ActionBlockStatement))
	}
	if shouldExecuteBlock {
		i.Stack = append(i.Stack, CallStackEntry{LocalVariables: i.mostRecentRegexCaptureGroups})
		for _, stmt := range block.GetStatements() {
			i.doStatement(stmt)
		}
		i.Stack = i.Stack[:len(i.Stack)-1]
	}
}

func (i *Interpreter) evaluateActionBlockConditon(block *ast.ActionBlockStatement) bool {
	evaluatedExpr := i.doExpression(block.Conditon)
	switch evaluatedExpr.(type) {
	case *ast.Boolean:
		return evaluatedExpr.(*ast.Boolean).Value
	default:
		panic("Expected Bool expression!!!")
	}
}
func (i *Interpreter) doPrintStatement(stmt *ast.PrintStatement) {
	toBePrinted := i.doExpressionList(stmt.Expressions)
	var asStrings []string
	for _, expr := range toBePrinted {
		asStrings = append(asStrings, expr.String())
	}
	fmt.Print(strings.Join(asStrings, " "))
	fmt.Print("\n")
}

func (i *Interpreter) doAssignStatement(stmt *ast.AssignStatement) {
	for idx, target := range stmt.Targets {
		i.setVar(target.String(), i.doExpression(stmt.Values[idx]).String())
	}
}

func (i *Interpreter) doAssignAndModifyStatement(stmt *ast.AssignAndModifyStatement) {
	var newValue ast.Expression
	switch stmt.Operator.Type {
	case token.ASSIGNPLUS:
		newValue = i.doExpression(&ast.InfixExpression{Left: stmt.Target, Operator: "+", Right: stmt.Value})
	default:
		panic("Unknown Operator.")
	}
	i.Stack[len(i.Stack)-1].LocalVariables[stmt.Target.String()] = newValue.String()
	i.setVar(stmt.Target.String(), newValue.String())
}

func (i *Interpreter) doExpressionList(expressions []ast.Expression) []ast.Expression {
	var results []ast.Expression
	for _, expr := range expressions {
		results = append(results, i.doExpression(expr))
	}
	return results
}

func (i *Interpreter) doExpression(expr ast.Expression) ast.Expression {
	switch expr.(type) {
	case *ast.InfixExpression:
		return i.doInfixExpression(expr.(*ast.InfixExpression))
	case *ast.CallExpression:
		return i.doFunctionCall(expr.(*ast.CallExpression))
	case *ast.PostfixExpression:
		return i.doPostfixExpression(expr.(*ast.PostfixExpression))
	}
	return expr
}

func (i *Interpreter) doInfixExpression(expression *ast.InfixExpression) ast.Expression {
	left := i.doExpression(expression.Left)
	right := i.doExpression(expression.Right)
	switch expression.Operator {
	case "~":
		return i.doRegexMatch(left, right)
	case "+":
		return i.doAdd(left, right)
	}
	return nil
}

func (i *Interpreter) doPostfixExpression(expression *ast.PostfixExpression) ast.Expression {
	switch expression.Operator {
	case "++":
		value := &ast.StringLiteral{Value: i.lookupVar(expression.Left.String())}
		i.setVar(expression.Left.String(), i.doExpression(&ast.InfixExpression{Left: expression.Left, Operator: "+", Right: &ast.IntegerLiteral{Value: 1}}).String())
		return value
	default:
		panic("Unknown postfix operator!")
	}
}

func (i *Interpreter) doRegexMatch(left ast.Expression, right ast.Expression) ast.Expression {
	var str string
	var regex string
	var isReadingFromInput bool
	switch left.(type) {
	case *ast.Identifier:
		if left.(*ast.Identifier).Value == "$0" && len(i.Stack) == 1 {
			isReadingFromInput = true
		}
		str = i.lookupVar(left.(*ast.Identifier).Value)
	case *ast.StringLiteral:
		str = (left.(*ast.StringLiteral).Value)
	default:
		panic("non-string match against regex")
	}

	switch right.(type) {
	case *ast.RegexLiteral:
		regex = right.(*ast.RegexLiteral).Value
	default:
		panic("non-regex match against string")
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		panic("invalid regex")
	}

	matches := re.FindStringSubmatch(str)
	if matches != nil {
		if isReadingFromInput {
			prevMatches := matches
			prevMatch := &matches[0]
			i.advanceInput()
			newMatches := re.FindStringSubmatch(i.Stack[0].LocalVariables["$0"])
			newMatch := &newMatches[0]
			for *prevMatch != *newMatch {
				i.advanceInput()
				prevMatches = newMatches
				prevMatch = newMatch
				newMatches := re.FindStringSubmatch(i.Stack[0].LocalVariables["$0"])
				newMatch = &newMatches[0]
			}

			i.backtrackInput()
			i.consumeInput()
			matches = prevMatches
		}
		for idx, match := range matches {
			stridx := "$" + strconv.Itoa(idx)
			i.mostRecentRegexCaptureGroups[stridx] = match
		}
		return &ast.Boolean{Value: true}
	}
	return &ast.Boolean{Value: false}
}

func (i *Interpreter) doFunctionCall(call *ast.CallExpression) ast.Expression {
	evaluatedArgs := i.doExpressionList(call.Arguments)
	function, ok := i.StdLibFunctions[call.Function.String()]
	if !ok {
		panic("Function not found!")
	}
	return function(evaluatedArgs)
}

func (i *Interpreter) doAdd(left ast.Expression, right ast.Expression) ast.Expression {
	var l string
	switch left.(type) {
	case *ast.Identifier:
		l = i.lookupVar(left.(*ast.Identifier).Value)
	case *ast.StringLiteral:
		l = (left.(*ast.StringLiteral).Value)
	default:
		panic("non-string match against regex")
	}

	var r string
	switch right.(type) {
	case *ast.Identifier:
		r = i.lookupVar(right.(*ast.Identifier).Value)
	case *ast.StringLiteral:
		r = (right.(*ast.StringLiteral).Value)
	case *ast.IntegerLiteral:
		r = (right.(*ast.IntegerLiteral).String())
	default:
		r = i.doExpression(right).String()
	}
	lInt, ok := strconv.Atoi(l)
	if ok != nil {
		panic("lhs not int")
	}
	rInt, ok := strconv.Atoi(r)
	if ok != nil {
		panic("rhs not int")
	}
	return &ast.IntegerLiteral{Value: lInt + rInt}
}
