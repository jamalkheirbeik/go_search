package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

type model struct {
	TFPD docs // term frequency per document
	DF   tf   // document frequency for term
}

type doc struct {
	TF            tf
	COUNT         int
	LAST_MODIFIED time.Time
}

type docs = map[string]doc

type tf = map[string]int

type searchResult struct {
	Pages int
	Data  []string
}

func index_file(lexer *lexer, m *model, file_path string) {
	for len(lexer.content) > 0 {
		token, err := next_token(lexer)
		if err != nil {
			return
		}

		_, ok := m.TFPD[file_path]
		if !ok {
			info, err := os.Stat(file_path)
			var last_modified time.Time
			if err != nil {
				fmt.Printf("ERROR: cannot access file info. %s\n", err)
				last_modified = time.Now()
			} else {
				last_modified = info.ModTime()
			}
			m.TFPD[file_path] = doc{TF: make(map[string]int), COUNT: 0, LAST_MODIFIED: last_modified}
		}

		if m.TFPD[file_path].TF[token] == 0 {
			m.TFPD[file_path].TF[token] = 1
			if m.DF[token] == 0 {
				m.DF[token] = 1
			} else {
				m.DF[token] += 1
			}
		} else {
			m.TFPD[file_path].TF[token] += 1
		}

		if entry, ok := m.TFPD[file_path]; ok {
			entry.COUNT += 1
			m.TFPD[file_path] = entry
		}
	}
}

func remove_file_from_model(m *model, file_path string) {
	for key := range m.TFPD[file_path].TF {
		if m.DF[key] != 0 {
			m.DF[key] -= 1
		}
	}
	delete(m.TFPD, file_path)
}

func add_dir_files_to_model(directory string, m *model) {
	fmt.Printf("reading files in directory '%s'...\n", directory)
	entries, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("ERROR: cannot read files in directory '%s'. %s\n", directory, err)
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				add_dir_files_to_model(directory+"/"+entry.Name(), m)
			} else {
				file_path := directory + "/" + entry.Name()
				stats, err := os.Stat(file_path)
				if err != nil || stats.ModTime().After(m.TFPD[file_path].LAST_MODIFIED) {
					remove_file_from_model(m, file_path)
					lexer := lexer{parse_file_by_extension(file_path)}
					index_file(&lexer, m, file_path)
				}
			}
		}
	}
}

func generate_index_file(directory string) {
	document := model{TFPD: make(map[string]doc), DF: make(map[string]int)}
	data, _ := os.ReadFile("index.json")
	json.Unmarshal(data, &document)
	add_dir_files_to_model(directory, &document)
	bytes, _ := json.Marshal(document)
	err := os.WriteFile("index.json", bytes, 0644)
	if err != nil {
		fmt.Printf("ERROR: cannot write index.json file. %s\n", err)
	}
}

func calculate_tf(term string, document doc) float64 {
	a := document.TF[term]
	b := document.COUNT
	return float64(a) / float64(b)
}

func calculate_idf(term string, m model) float64 {
	a := len(m.TFPD)
	b, ok := m.DF[term]
	if !ok {
		b = 1
	}
	return math.Log(float64(a) / float64(b))
}

func search(query string, page_number int) string {
	const MAX_ITEMS_PER_PAGE = 10
	if page_number <= 0 {
		page_number = 1
	}
	result := &searchResult{Pages: 0, Data: make([]string, 0)}
	data, err := os.ReadFile("index.json")
	if err != nil {
		fmt.Printf("ERROR: cannot read 'index.json' file. %s\n", err)
	} else {
		var m model
		json.Unmarshal(data, &m)

		ranks := make(map[string]float64)
		for file_path := range m.TFPD {
			lexer := lexer{content: query}
			var rank float64 = 0
			for len(lexer.content) > 0 {
				token, err := next_token(&lexer)
				if err != nil {
					continue
				}
				rank += calculate_tf(token, m.TFPD[file_path]) * calculate_idf(token, m)
			}
			ranks[file_path] = rank
		}

		result.Pages = int(math.Ceil(float64(len(ranks)) / float64(MAX_ITEMS_PER_PAGE)))
		if page_number <= result.Pages {
			// sorting ranks
			type kv struct {
				Key   string
				Value float64
			}
			var tmp []kv
			for k, v := range ranks {
				tmp = append(tmp, kv{k, v})
			}

			sort.Slice(tmp, func(i, j int) bool {
				return tmp[i].Value > tmp[j].Value
			})
			// fetching data within page scope
			start := (page_number - 1) * MAX_ITEMS_PER_PAGE
			end := start + MAX_ITEMS_PER_PAGE
			if end >= len(tmp) {
				end = len(tmp)
			}
			for i := start; i < end; i++ {
				fmt.Printf("    %s => %f\n", tmp[i].Key, tmp[i].Value)
				result.Data = append(result.Data, tmp[i].Key)
			}
		}
	}
	b, _ := json.Marshal(result)
	return string(b)
}
