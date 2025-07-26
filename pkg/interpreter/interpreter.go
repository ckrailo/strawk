package interpreter

import (
	"errors"
	"fmt"
	"math"
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
	GlobalVariables              map[string]ast.Expression
	StdLibFunctions              map[string]func([]ast.Expression) ast.Expression
	mostRecentRegexCaptureGroups map[string]ast.Expression
}

type CallStackEntry struct {
	isFunction     bool
	LocalVariables map[string]ast.Expression
}

func NewInterpreter(program *ast.Program, input string) *Interpreter {
	i := &Interpreter{Program: program, Input: input, GlobalVariables: make(map[string]ast.Expression), StdLibFunctions: make(map[string]func([]ast.Expression) ast.Expression)}
	i.Stack = append(i.Stack, CallStackEntry{})
	i.Stack[0].LocalVariables = make(map[string]ast.Expression)
	i.Stack[0].LocalVariables["$0"] = &ast.StringLiteral{Value: ""}
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
	i.Stack[0].LocalVariables["$0"] = i.doConcatenate(i.Stack[0].LocalVariables["$0"], &ast.StringLiteral{Value: string(i.Input[i.InputPostion])})
}

func (i *Interpreter) backtrackInput() {
	i.InputPostion -= 1
	length := len(i.Stack[0].LocalVariables["$0"].(*ast.StringLiteral).Value)
	i.Stack[0].LocalVariables["$0"] = &ast.StringLiteral{Value: i.Stack[0].LocalVariables["$0"].(*ast.StringLiteral).Value[:length-1]}
}

func (i *Interpreter) consumeInput() {
	i.Stack[0].LocalVariables["$0"] = &ast.StringLiteral{Value: ""}
}

func (i *Interpreter) lookupVar(varName string) ast.Expression {
	val, ok := i.Stack[len(i.Stack)-1].LocalVariables[varName]
	if ok {
		return val
	}
	val, ok = i.GlobalVariables[varName]
	if ok {
		return val
	}
	return &ast.StringLiteral{Value: ""}
}

func (i *Interpreter) setVar(varName string, value ast.Expression) {
	_, ok := i.Stack[len(i.Stack)-1].LocalVariables[varName]
	if ok {
		i.Stack[len(i.Stack)-1].LocalVariables[varName] = value
	} else {
		i.GlobalVariables[varName] = value
	}
}

func (i *Interpreter) createLocalVar(varName string, value ast.Expression) {
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
	i.mostRecentRegexCaptureGroups = make(map[string]ast.Expression)
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
		i.setVar(target.String(), i.doExpression(stmt.Values[idx]))
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
	i.setVar(stmt.Target.String(), newValue)
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
	case *ast.Identifier:
		return i.lookupVar(expr.(*ast.Identifier).Value)
	}
	return expr
}

func (i *Interpreter) doInfixExpression(expression *ast.InfixExpression) ast.Expression {
	left := i.doExpression(expression.Left)
	right := i.doExpression(expression.Right)
	switch expression.Operator {
	case ".":
		return i.doConcatenate(left, right)
	case "~":
		return i.doRegexMatch(left, right, false)
	case "~$0":
		return i.doRegexMatch(left, right, true)
	case "+":
		return i.doAdd(left, right)
	case "-":
		return i.doMinus(left, right)
	case "*":
		return i.doMultiply(left, right)
	case "/":
		return i.doDivide(left, right)
	case "%":
		return i.doModulus(left, right)
	case "^":
		return i.doExponentiation(left, right)
	case "==":
		return i.doEquality(left, right)
	case "!=":
		return i.doNotEquals(left, right)
	case "<":
		return i.doLessThan(left, right)
	case ">":
		return i.doGreaterThan(left, right)
	case "<=":
		return i.doLessThanEqualTo(left, right)
	case ">=":
		return i.doGreaterThanEqualTo(left, right)
	}
	return nil
}

func (i *Interpreter) doPostfixExpression(expr *ast.PostfixExpression) ast.Expression {
	switch expr.Operator {
	case "++":
		value := &ast.StringLiteral{Value: i.lookupVar(expr.Left.String()).String()}
		i.setVar(expr.Left.String(), i.doExpression(&ast.InfixExpression{Left: expr.Left, Operator: "+", Right: &ast.NumericLiteral{Value: 1}}))
		return value
	case "--":
		value := &ast.StringLiteral{Value: i.lookupVar(expr.Left.String()).String()}
		i.setVar(expr.Left.String(), i.doExpression(&ast.InfixExpression{Left: expr.Left, Operator: "-", Right: &ast.NumericLiteral{Value: 1}}))
		return value
	default:
		panic("Unknown postfix operator!")
	}
}

func (i *Interpreter) doRegexMatch(left ast.Expression, right ast.Expression, isReadingFromInput bool) ast.Expression {
	var str string
	var regex string
	if isReadingFromInput && len(i.Stack) == 1 {
		isReadingFromInput = true
	} else {
		isReadingFromInput = false
	}

	switch left.(type) {
	case *ast.Identifier:
		str = i.lookupVar(left.(*ast.Identifier).Value).String()
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
			newMatches := re.FindStringSubmatch(i.Stack[0].LocalVariables["$0"].(*ast.StringLiteral).Value)
			newMatch := &newMatches[0]
			for *prevMatch != *newMatch {
				i.advanceInput()
				prevMatches = newMatches
				prevMatch = newMatch
				newMatches := re.FindStringSubmatch(i.Stack[0].LocalVariables["$0"].(*ast.StringLiteral).Value)
				newMatch = &newMatches[0]
			}

			i.backtrackInput()
			i.consumeInput()
			matches = prevMatches
		}
		for idx, match := range matches {
			stridx := "$" + strconv.Itoa(idx)
			i.mostRecentRegexCaptureGroups[stridx] = ast.NewLiteral(match)
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

func convertLiteralForMathOp(expr ast.Expression) float64 {
	switch expr.(type) {
	case *ast.StringLiteral:
		return 0.0
	case *ast.NumericLiteral:
		return (expr.(*ast.NumericLiteral).Value)
	default:
		panic("error in math")
	}
}

func (i *Interpreter) doAdd(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: convertLiteralForMathOp(left) + convertLiteralForMathOp(right)}
}

func (i *Interpreter) doMinus(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: convertLiteralForMathOp(left) - convertLiteralForMathOp(right)}
}

func (i *Interpreter) doMultiply(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: convertLiteralForMathOp(left) * convertLiteralForMathOp(right)}
}

func (i *Interpreter) doDivide(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: convertLiteralForMathOp(left) / convertLiteralForMathOp(right)}
}

func (i *Interpreter) doModulus(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: math.Mod(convertLiteralForMathOp(left), convertLiteralForMathOp(right))}
}

func (i *Interpreter) doExponentiation(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: math.Pow(convertLiteralForMathOp(left), convertLiteralForMathOp(right))}
}

func convertLiteralForStringOp(expr ast.Expression) string {
	switch expr.(type) {
	case *ast.StringLiteral:
		return expr.(*ast.StringLiteral).Value
	case *ast.NumericLiteral:
		return (expr.(*ast.NumericLiteral).String())
	default:
		panic("error in literal to string conversion")
	}
}

func (i *Interpreter) doConcatenate(left ast.Expression, right ast.Expression) ast.Expression {
	lhs := convertLiteralForStringOp(left)
	rhs := convertLiteralForStringOp(right)
	return &ast.StringLiteral{Value: lhs + rhs}
}

func boolToExpression(b bool) ast.Expression {
	if b {
		return &ast.StringLiteral{Value: "1"}
	} else {
		return &ast.StringLiteral{Value: "0"}
	}
}

func (i *Interpreter) doEquality(left ast.Expression, right ast.Expression) ast.Expression {
	lhs := convertLiteralForStringOp(left)
	rhs := convertLiteralForStringOp(right)
	return boolToExpression(lhs == rhs)
}

func (i *Interpreter) doNotEquals(left ast.Expression, right ast.Expression) ast.Expression {
	lhs := convertLiteralForStringOp(left)
	rhs := convertLiteralForStringOp(right)
	return boolToExpression(lhs != rhs)
}

func convertLiteralForComparisonOp(expr ast.Expression) (float64, error) {
	switch expr.(type) {
	case *ast.StringLiteral:
		return 0.0, errors.New("failed to convert to float")
	case *ast.NumericLiteral:
		return (expr.(*ast.NumericLiteral).Value), nil
	default:
		panic("error in math")
	}
}

func (i *Interpreter) doGreaterThan(left ast.Expression, right ast.Expression) ast.Expression {
	lhs_float, lerr := convertLiteralForComparisonOp(left)
	rhs_float, rerr := convertLiteralForComparisonOp(right)
	if lerr == nil && rerr == nil {
		return boolToExpression(lhs_float > rhs_float)
	}

	lhs_str := convertLiteralForStringOp(left)
	rhs_str := convertLiteralForStringOp(right)
	return boolToExpression(lhs_str > rhs_str)
}

func (i *Interpreter) doGreaterThanEqualTo(left ast.Expression, right ast.Expression) ast.Expression {
	lhs_float, lerr := convertLiteralForComparisonOp(left)
	rhs_float, rerr := convertLiteralForComparisonOp(right)
	if lerr == nil && rerr == nil {
		return boolToExpression(lhs_float >= rhs_float)
	}

	lhs_str := convertLiteralForStringOp(left)
	rhs_str := convertLiteralForStringOp(right)
	return boolToExpression(lhs_str >= rhs_str)
}

func (i *Interpreter) doLessThan(left ast.Expression, right ast.Expression) ast.Expression {
	lhs := convertLiteralForStringOp(left)
	rhs := convertLiteralForStringOp(right)
	return boolToExpression(lhs < rhs)
}

func (i *Interpreter) doLessThanEqualTo(left ast.Expression, right ast.Expression) ast.Expression {
	lhs := convertLiteralForStringOp(left)
	rhs := convertLiteralForStringOp(right)
	return boolToExpression(lhs <= rhs)
}
