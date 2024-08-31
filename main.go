package main

import (
	"fmt"
	"os"
	"strconv"
)

func list_commands() {
	fmt.Println("Usage <program> <subcommand> [args]")
	fmt.Println("    index <directory>           recursively indexes all files in directory and generates an index.json file containing frequency table")
	fmt.Println("    search <query> <page>       Searches within index.json, ranks the result using TF-IDF, divides the result into pages and returns the data related to the provided page")
	fmt.Println("    serve [port]                Serves a local HTTP server on port 8080 by default")
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
			if len(os.Args) > 3 {
				query := os.Args[2]
				if len(query) == 0 {
					fmt.Println("ERROR: search query cannot be empty")
				} else {
					page, _ := strconv.Atoi(os.Args[3])
					search(query, page)
				}
			} else {
				fmt.Println("ERROR: Missing arguements. Use the help command to check out the usage.")
			}
		case "serve":
			port := "8080"
			if len(os.Args) > 2 {
				port = os.Args[2]
			}
			serve(port)
		default:
			fmt.Println("ERROR: Unknown subcommand, to list available commands run 'go run main.go help'")
		}
	} else {
		fmt.Println("ERROR: No subcommand has been provided, to list available commands run 'go run main.go help'")
	}
}
