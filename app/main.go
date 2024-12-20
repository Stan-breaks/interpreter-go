package main

import (
	"fmt"
	"os"

	"github.com/Stan-breaks/app/parse"
	"github.com/Stan-breaks/app/tokenize"
	"github.com/Stan-breaks/app/utils"
)

func main() {
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: ./your_program.sh tokenize <filename>")
		os.Exit(1)
	}

	command := os.Args[1]

	filename := os.Args[2]
	rawFileContents, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	fileContents := string(rawFileContents)
	fileLenght := len(fileContents) - 1
	tokens := tokenize.Tokenize(fileContents, fileLenght)
	switch command {
	case "tokenize":
		for _, token := range tokens.Success {
			fmt.Println(token)
		}
		fmt.Println("EOF  null")
		if len(tokens.Errors) != 0 {
			for _, err := range tokens.Errors {
				fmt.Fprintf(os.Stderr, "%s\n", err)
			}
			os.Exit(65)
		}
	case "parse":
		value := parse.Parse(tokens)
		result := value.Evaluate()
		switch v := result.(type) {
		case float32:
			fmt.Println(utils.FormatFloat(v))
		default:
			fmt.Println(result)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
