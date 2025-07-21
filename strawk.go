package main

import (
	"os"

	"github.com/ahalbert/strawk/pkg/interpreter"
	"github.com/ahalbert/strawk/pkg/lexer"
	"github.com/ahalbert/strawk/pkg/parser"
)

func main() {
	program, _ := os.ReadFile("./test.awk")
	input, _ := os.ReadFile("./input.txt")
	l := lexer.New(string(program))
	p := parser.New(l)
	parsedprogram := p.ParseProgram()
	i := interpreter.NewInterpreter(parsedprogram, string(input))
	i.Run()
}
