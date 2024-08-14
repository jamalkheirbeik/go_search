package main

import (
	"fmt"
	"os"
)

func list_commands() {
	fmt.Println("Usage <subcommand> [args]")
	fmt.Println("    index <directory>           recursively indexes all files in directory and generates an index.json file containing frequency table")
	fmt.Println("    search <query>              TF-IDF search within index.json and returns documents matching that query sorted from most relative to least")
	fmt.Println("    serve                       Serves a local HTTP server on port 8080")
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			list_commands()
		case "index":
			if len(os.Args) > 2 {
				generate_index_file(os.Args[2])
			} else {
				fmt.Println("ERROR: No directory has been provided")
			}
		case "search":
			if len(os.Args) > 2 {
				search(os.Args[2])
			} else {
				fmt.Println("ERROR: search query is not provided")
			}
		default:
			fmt.Println("ERROR: Unknown subcommand, to list available commands run 'go run main.go help'")
		}
	} else {
		fmt.Println("ERROR: No subcommand has been provided, to list available commands run 'go run main.go help'")
	}
}
