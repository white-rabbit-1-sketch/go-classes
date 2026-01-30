package main

import (
	"fmt"
	"goo/internal/prsr"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: goo <input.goo> <output.go>")
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Read error: %v\n", err)
		return
	}

	prsr := prsr.NewParser()
	compiledData, err := prsr.Parse(string(data))
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	err = os.WriteFile(outputFile, compiledData, 0644)
	if err != nil {
		fmt.Printf("Write error: %v\n", err)
		return
	}
}
