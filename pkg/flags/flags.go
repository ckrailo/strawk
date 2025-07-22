package flags

var Flags struct {
	ProgramFile string   `arg:"-f" placeholder:"PROGRAMFILE" help:"Program File to run."`
	Program     string   `arg:"positional" help:"Program to run."`
	InputFiles  []string `arg:"positional" placeholder:"INPUTFILE" help:"File to use as input."`
}
