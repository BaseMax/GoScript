package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <file>\n", os.Args[0])
		return
	}

	fileContent, fileErr := os.ReadFile(os.Args[1])
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	src := string(fileContent)

	scn := CreateScanner(src)
	astParser := CreateParser(scn.lexemes)
	env := CreateEnvironment(nil)
	EvaluateNodes(astParser.astNodes, env)
}
