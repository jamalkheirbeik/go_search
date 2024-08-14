package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type model struct {
	TFPD doc // term frequency per document
	DF   tf  // term frequency across all documents
}

type doc = map[string]tf

type tf = map[string]int

func read_text_file(file_path string) string {
	data, err := os.ReadFile(file_path)
	if err != nil {
		fmt.Printf("ERROR: could not read text file '%s'. %s\n", file_path, err)
		return ""
	} else {
		str := string(data)
		fmt.Printf("'%s' => %d\n", file_path, len(str))
		return str
	}
}

func index_document(lexer *lexer, model *model, file_path string) {
	for len(lexer.content) > 0 {
		token, err := next_token(lexer)
		if err != nil {
			return
		}

		if len(model.TFPD[file_path]) == 0 {
			model.TFPD[file_path] = make(map[string]int)
		}

		if model.TFPD[file_path][token] == 0 {
			model.TFPD[file_path][token] = 1
		} else {
			model.TFPD[file_path][token] += 1
		}

		if model.DF[token] == 0 {
			model.DF[token] = 1
		} else {
			model.DF[token] += 1
		}
	}
}

func add_dir_files_to_model(directory string, model *model) {
	fmt.Printf("reading files in directory '%s'...\n", directory)
	entries, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("ERROR: cannot read files in directory '%s'. %s\n", directory, err)
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				add_dir_files_to_model(directory+"/"+entry.Name(), model)
			} else {
				extension := filepath.Ext(entry.Name())
				file_path := directory + "/" + entry.Name()
				switch extension {
				case ".txt":
					lexer := lexer{read_text_file(file_path)}
					index_document(&lexer, model, file_path)
				case ".md":
					fmt.Println("TODO: Parse markdown documents")
				case ".xml":
					fmt.Println("TODO: Parse XML documents")
				case ".xhtml":
					fmt.Println("TODO: Parse XHTML documents")
				case ".html":
					fmt.Println("TODO: Parse HTML documents")
				case ".pdf":
					fmt.Println("TODO: Parse PDF documents")
				default:
					fmt.Printf("ERROR: Extension '%s' is not supported\n", extension)
				}
			}
		}
	}
}

func generate_index_file(directory string) {
	model := model{TFPD: make(map[string]map[string]int), DF: make(map[string]int)}
	add_dir_files_to_model(directory, &model)
	bytes, _ := json.Marshal(model)
	err := os.WriteFile("index.json", bytes, 0644)
	if err != nil {
		fmt.Printf("ERROR: cannot write index.json file. %s\n", err)
	}
}

func search(_ string) {
	fmt.Println("TODO: implement searching through indexed model by query")
}
