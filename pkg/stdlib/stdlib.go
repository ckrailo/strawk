package stdlib

import (
	"github.com/ahalbert/strawk/pkg/ast"
)

func Length(args []ast.Expression) ast.Expression {

	if len(args) != 1 {
		panic("Incorrect arguments to function length")
	}

	arg := args[0].String()
	ret := float64(len(arg))
	return &ast.NumericLiteral{Value: ret}
}
