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

func NewJSONModel() *model {
	m := model{TFPD: make(map[string]doc), DF: make(map[string]int)}
	return &m
}

func (m *model) populate() {
	file_path := "./storage/index.json"
	data, err := os.ReadFile(file_path)
	if err != nil {
		fmt.Printf("ERROR: Cannot read '%s' file. %s\n", file_path, err)
	}
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Printf("ERROR: Cannot parse the JSON data in '%s'. %s\n", file_path, err)
	}
}

func (m *model) index_file(f *file) {
	l := lexer{f.parse()}
	for len(l.content) > 0 {
		token, err := l.next_token()
		if err != nil {
			continue
		}

		_, ok := m.TFPD[f.path]
		if !ok {
			m.TFPD[f.path] = doc{TF: make(map[string]int), COUNT: 0, LAST_MODIFIED: f.last_modified}
		}

		if m.TFPD[f.path].TF[token] == 0 {
			m.TFPD[f.path].TF[token] = 1
			if m.DF[token] == 0 {
				m.DF[token] = 1
			} else {
				m.DF[token] += 1
			}
		} else {
			m.TFPD[f.path].TF[token] += 1
		}

		if entry, ok := m.TFPD[f.path]; ok {
			entry.COUNT += 1
			m.TFPD[f.path] = entry
		}
	}
}

func (m *model) remove_file(f *file) {
	for key := range m.TFPD[f.path].TF {
		if m.DF[key] != 0 {
			m.DF[key] -= 1
		}
	}
	delete(m.TFPD, f.path)
}

func (m *model) add_dir_files(directory string) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("ERROR: Cannot read files in directory '%s'. %s\n", directory, err)
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				m.add_dir_files(directory + "/" + entry.Name())
			} else {
				file_path := directory + "/" + entry.Name()
				f := NewFile(file_path)
				fmt.Printf("INFO: Indexing file '%s'...\n", f.path)
				if f.last_modified.After(m.TFPD[f.path].LAST_MODIFIED) {
					m.remove_file(f)
					m.index_file(f)
				}
			}
		}
	}
}

func (m *model) generate_index_file(directory string) {
	m.add_dir_files(directory)
	bytes, _ := json.Marshal(m)
	err := os.WriteFile("./storage/index.json", bytes, 0644)
	if err != nil {
		fmt.Printf("ERROR: Cannot write to './storage/index.json'. %s\n", err)
	}
}

func (m *model) calculate_tf(file_path string, term string) float64 {
	a := m.TFPD[file_path].TF[term]
	b := m.TFPD[file_path].COUNT
	return float64(a) / float64(b)
}

func (m *model) calculate_idf(term string) float64 {
	a := len(m.TFPD)
	b, ok := m.DF[term]
	if !ok {
		b = 1
	}
	return math.Log(float64(a) / float64(b))
}

func (m *model) search(query string, page_number int) string {
	const LIMIT = 10
	if page_number <= 0 {
		page_number = 1
	}
	result := &searchResult{Pages: 0, Data: make([]string, 0)}

	ranks := make(map[string]float64)
	for file_path := range m.TFPD {
		l := lexer{content: query}
		var rank float64 = 0
		for len(l.content) > 0 {
			token, err := l.next_token()
			if err != nil {
				continue
			}
			rank += m.calculate_tf(file_path, token) * m.calculate_idf(token)
		}
		ranks[file_path] = rank
	}

	result.Pages = int(math.Ceil(float64(len(ranks)) / float64(LIMIT)))
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
		start := (page_number - 1) * LIMIT
		end := start + LIMIT
		if end >= len(tmp) {
			end = len(tmp)
		}
		for i := start; i < end; i++ {
			fmt.Printf("    %s => %f\n", tmp[i].Key, tmp[i].Value)
			result.Data = append(result.Data, tmp[i].Key)
		}
	}
	b, _ := json.Marshal(result)
	return string(b)
}
