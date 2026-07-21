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
	env := runtime.New()
	eval := evaluator.New(env)

	if len(os.Args) > 1 {
		filename := os.Args[1]
		runFile(filename, eval)
	} else {
		runREPL(eval)
	}
}

func runFile(filename string, eval *evaluator.Evaluator) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	l := lexer.New(content)
	p := parser.New(l)
	program, err := p.ParseProgram()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	evaluated := eval.Eval(program)
	if evaluated != nil {
		fmt.Fprintln(os.Stdout, evaluated.Inspect())
	}
}

func runREPL(eval *evaluator.Evaluator) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to gjs - a simple JavaScript interpreter")
	fmt.Println("Type JavaScript code or 'exit' to quit.")
	for {
		fmt.Print(PROMPT)
		if !scanner.Scan() {
			return
		}
		line := scanner.Text()
		if line == "exit" || line == "quit" {
			return
		}
		l := lexer.New([]byte(line))
		p := parser.New(l)
		program, err := p.ParseProgram()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		evaluated := eval.Eval(program)
		if evaluated != nil {
			fmt.Fprintln(os.Stdout, evaluated.Inspect())
		}
	}
}
