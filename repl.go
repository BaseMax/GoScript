package main

import (
	"bufio"
	"fmt"
	"os"
)

func evalArgs(nodes []Node, s *Scope) []any {
	var result []any
	for _, n := range nodes {
		result = append(result, n.Eval(s))
	}
	return result
}

func applyFunction(fn *FunctionNode, args []any, newScope bool) any {
	scope := fn.Scope
	if newScope {
		scope = NewScope(fn.Scope)
	}
	for i, param := range fn.Params {
		scope.SetVariable(param.Name, args[i])
	}
	return fn.Body.Eval(scope)
}

func Evaluate(nodes chan Node, scope *Scope) any {
	var result any
	for node := range nodes {
		fmt.Printf("Node: %+v\n", node)
		result = node.Eval(scope)
	}
	return result
}

func RunREPL() {
	fmt.Println("GoScript Version 0.1")
	scanner := bufio.NewScanner(os.Stdin)
	scope := NewScope(nil)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		fmt.Printf("Lexer: ")
		lexer := NewLexer(input)
		fmt.Printf("Lex: %+v\n", lexer)
		fmt.Printf("Parser: ")
		parser := NewParser(lexer.tokens)
		fmt.Printf("AST: %+v\n", parser)
		fmt.Printf("Evaluate: ")
		result := Evaluate(parser.nodes, scope)
		fmt.Println(result)
	}
}
