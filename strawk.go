package main

import (
	"os"

	"github.com/ahalbert/strawk/pkg/flags"
	"github.com/ahalbert/strawk/pkg/interpreter"
	"github.com/ahalbert/strawk/pkg/lexer"
	"github.com/ahalbert/strawk/pkg/parser"
	"github.com/alexflint/go-arg"
)

func main() {

	arg.MustParse(&flags.Flags)

	var program string
	if flags.Flags.ProgramFile != "" {
		buf, err := os.ReadFile(flags.Flags.ProgramFile)
		if err != nil {
			panic("Program File " + flags.Flags.ProgramFile + " not found")
		}
		program = string(buf)
		if flags.Flags.Program != "" {
			flags.Flags.InputFiles = append([]string{flags.Flags.Program}, flags.Flags.InputFiles...)
		}
	} else {
		program = flags.Flags.Program
	}

	if program == "" {
		panic("no program supplied")
	}

	var input []byte
	if len(flags.Flags.InputFiles) > 0 {
		input, _ = os.ReadFile(flags.Flags.InputFiles[0])
	}
	l := lexer.New(string(program))
	p := parser.New(l)
	parsedprogram := p.ParseProgram()
	i := interpreter.NewInterpreter(parsedprogram, os.Stdout)
	i.Run(string(input))
}
