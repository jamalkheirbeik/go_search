package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

type file struct {
	path          string
	extension     string
	last_modified time.Time
}

func NewFile(file_path string) *file {
	info, err := os.Stat(file_path)
	var last_modified time.Time
	if err != nil {
		last_modified = time.Now()
	} else {
		last_modified = info.ModTime()
	}
	return &file{path: file_path, extension: filepath.Ext(file_path), last_modified: last_modified}
}

type node struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
	Nodes   []node `xml:",any"`
}

func iterate_xml_nodes(nodes []node, f func(node) bool) {
	for _, n := range nodes {
		if f(n) {
			iterate_xml_nodes(n.Nodes, f)
		}
	}
}

func (f *file) parse_text() string {
	var result string
	data, err := os.ReadFile(f.path)
	if err != nil {
		fmt.Printf("ERROR: cannot read text file '%s'. %s\n", f.path, err)
	} else {
		result = string(data)
	}
	return result
}

func array_contains_string(arr []string, target string) bool {
	for _, str := range arr {
		if str == target {
			return true
		}
	}
	return false
}

func (f *file) parse_xml(strict_mode bool, excluded_tags ...string) string {
	var result string
	data, err := os.ReadFile(f.path)
	if err != nil {
		fmt.Printf("ERROR: cannot read file '%s'. %s\n", f.path, err)
	} else {
		buff := bytes.NewBuffer(data)
		dec := xml.NewDecoder(buff)
		dec.Strict = strict_mode
		var n node
		err := dec.Decode(&n)
		if err != nil {
			fmt.Printf("ERROR: cannot decode node in file '%s'. %s\n", f.path, err)
		} else {
			iterate_xml_nodes([]node{n}, func(n node) bool {
				if !array_contains_string(excluded_tags, n.XMLName.Local) {
					content := string(n.Content)
					content = strings.TrimLeft(content, " \r\n\t")
					if len(content) > 0 && string(content[0]) != "<" {
						result += content + "\n"
					}
				}
				return true
			})
		}
	}
	return result
}

func (f *file) parse_pdf() string {
	// pdf.DebugOn = true
	var result string
	file, r, err := pdf.Open(f.path)
	if err != nil {
		fmt.Printf("ERROR: cannot parse file '%s'. %s\n", f.path, err)
	} else {
		var buf bytes.Buffer
		b, err := r.GetPlainText()
		if err != nil {
			fmt.Printf("ERROR: cannot get plain text from file '%s'. %s\n", f.path, err)
		} else {
			buf.ReadFrom(b)
			result = buf.String()
		}
	}
	file.Close()
	return result
}

// automatically parses file based on its extension
func (f *file) parse() string {
	var result string
	switch f.extension {
	case ".txt", ".md":
		result = f.parse_text()
	case ".xml":
		result = f.parse_xml(true)
	case ".xhtml", ".html", ".htm":
		result = f.parse_xml(false, "style", "script")
	case ".pdf":
		result = f.parse_pdf()
	default:
		fmt.Printf("INFO: files with extension '%s' are not supported and will be ignored.\n", f.extension)
	}
	return result
}
