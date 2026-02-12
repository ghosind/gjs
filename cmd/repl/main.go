package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ghosind/gjs/evaluator"
	"github.com/ghosind/gjs/lexer"
	"github.com/ghosind/gjs/parser"
	"github.com/ghosind/gjs/runtime"
)

const PROMPT = "> "

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	env := runtime.New()
	eval := evaluator.New(env)
	for {
		fmt.Print(PROMPT)
		if !scanner.Scan() {
			return
		}
		line := scanner.Bytes()
		l := lexer.New(line)
		p := parser.New(l)
		program, err := p.ParseProgram()
		if err != nil {
			fmt.Fprintf(os.Stderr, "parser error: %s\n", err)
			continue
		}
		evaluated := eval.Eval(program)
		if evaluated != nil {
			fmt.Fprintln(os.Stdout, evaluated.Inspect())
		}
	}
}
