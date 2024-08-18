package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

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

func parse_text_file(file_path string) string {
	data, err := os.ReadFile(file_path)
	if err != nil {
		fmt.Printf("ERROR: cannot read text file '%s'. %s\n", file_path, err)
		return ""
	} else {
		return string(data)
	}
}

func array_contains_string(arr []string, target string) bool {
	for _, str := range arr {
		if str == target {
			return true
		}
	}
	return false
}

func parse_xml_file(file_path string, strict_mode bool, excluded_tags ...string) string {
	var result string
	data, err := os.ReadFile(file_path)
	if err != nil {
		fmt.Printf("ERROR: cannot read file '%s'. %s\n", file_path, err)
	} else {
		buff := bytes.NewBuffer(data)
		dec := xml.NewDecoder(buff)
		dec.Strict = strict_mode
		var n node
		err := dec.Decode(&n)
		if err != nil {
			fmt.Printf("ERROR: cannot decode node in file '%s'. %s\n", file_path, err)
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

func parse_pdf_file(file_path string) string {
	pdf.DebugOn = true
	var result string
	f, r, err := pdf.Open(file_path)
	if err != nil {
		fmt.Printf("ERROR: cannot parse file '%s'. %s\n", file_path, err)
	} else {
		var buf bytes.Buffer
		b, err := r.GetPlainText()
		if err != nil {
			fmt.Printf("ERROR: cannot get plain text from file '%s'. %s\n", file_path, err)
		} else {
			buf.ReadFrom(b)
			result = buf.String()
		}
	}
	f.Close()
	return result
}

func parse_file_by_extension(file_path string) string {
	var result string
	extension := filepath.Ext(file_path)
	switch extension {
	case ".txt", ".md":
		result = parse_text_file(file_path)
	case ".xml":
		result = parse_xml_file(file_path, true)
	case ".xhtml", ".html", ".htm":
		result = parse_xml_file(file_path, false, "style", "script")
	case ".pdf":
		result = parse_pdf_file(file_path)
	default:
		fmt.Printf("ERROR: files with extension '%s' are not supported and will be ignored.\n", extension)
	}
	return result
}
