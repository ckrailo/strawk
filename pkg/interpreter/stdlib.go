package interpreter

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ahalbert/strawk/pkg/ast"
)

func Length(i *Interpreter, args []ast.Expression) ast.Expression {

	if len(args) != 1 {
		panic("Incorrect arguments to function length")
	}

	var ret float64
	switch args[0].(type) {
	case *ast.StringLiteral:
		arg := args[0].(*ast.StringLiteral).Value
		ret = float64(len(arg))
	case *ast.NumericLiteral:
		arg := args[0].(*ast.NumericLiteral).String()
		ret = float64(len(arg))
	case *ast.AssociativeArray:
		ret = float64(len(args[0].(*ast.AssociativeArray).Array))
	default:
		panic("Incorrect argument type to function length")
	}
	return &ast.NumericLiteral{Value: ret}
}

func Sub(i *Interpreter, args []ast.Expression) ast.Expression {
	var in ast.Expression
	if len(args) < 2 || len(args) > 3 {
		panic("Incorrect arguments to function sub")
	}
	if len(args) == 2 {
		in = i.lookupVar(&ast.Identifier{Value: "$0"})
	} else {
		in = args[2]
	}

	switch args[0].(type) {
	case *ast.RegexLiteral:
	default:
		panic("first argument to function sub is not a regex")
	}

	switch args[1].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("second argument to function sub is not a scalar")
	}

	switch in.(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("third argument to function sub is not a scalar")
	}

	re, err := regexp.Compile(args[0].(*ast.RegexLiteral).Value)
	if err != nil {
		panic("First argument to sub not a valid regex")
	}
	found := re.FindString(in.String())
	replaced := in.String()
	if found != "" {
		replaced = strings.Replace(in.String(), found, args[1].String(), 1)
	}
	return &ast.StringLiteral{Value: replaced}
}

func Gsub(i *Interpreter, args []ast.Expression) ast.Expression {
	var in ast.Expression
	if len(args) < 2 || len(args) > 3 {
		panic("Incorrect arguments to function sub")
	}
	if len(args) == 2 {
		in = i.lookupVar(&ast.Identifier{Value: "$0"})
	} else {
		in = args[2]
	}

	switch args[0].(type) {
	case *ast.RegexLiteral:
	default:
		panic("first argument to function sub is not a regex")
	}

	switch args[1].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("second argument to function sub is not a scalar")
	}

	switch in.(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("third argument to function sub is not a scalar")
	}

	re, err := regexp.Compile(args[0].(*ast.RegexLiteral).Value)
	if err != nil {
		panic("First argument to sub not a valid regex")
	}

	replaced := re.ReplaceAllString(in.String(), args[1].String())
	return &ast.StringLiteral{Value: replaced}
}

func Split(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 2 {
		panic("Incorrect arguments to function split")
	}

	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function sub is not a regex")
	}

	switch args[1].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("second argument to function sub is not a scalar")
	}
	splits := strings.Split(args[0].String(), args[1].String())
	ret := make(map[string]ast.Expression)
	for idx, split := range splits {
		ret[strconv.Itoa(idx+1)] = &ast.StringLiteral{Value: split}
	}
	return &ast.AssociativeArray{Array: ret}
}
