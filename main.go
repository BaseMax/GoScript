package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		RunREPL()
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read file %s: %v", os.Args[1], err)
	}
	lexer := NewLexer(string(data))
	parser := NewParser(lexer.tokens)
	scope := NewScope(nil)
	Evaluate(parser.nodes, scope)
}
