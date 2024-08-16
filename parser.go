package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		fmt.Printf("ERROR: could not read text file '%s'. %s\n", file_path, err)
		return ""
	} else {
		return string(data)
	}
}

func parse_xml_file(file_path string) string {
	data, err := os.ReadFile(file_path)
	if err != nil {
		fmt.Printf("ERROR: could not read xml file '%s'. %s\n", file_path, err)
		return ""
	} else {
		var result string
		buff := bytes.NewBuffer(data)
		dec := xml.NewDecoder(buff)
		var n node
		err := dec.Decode(&n)
		if err != nil {
			fmt.Printf("ERROR: cannot decode xml node. %s", err)
		} else {
			iterate_xml_nodes([]node{n}, func(n node) bool {
				content := string(n.Content)
				content = strings.TrimLeft(content, " \r\n\t")
				if len(content) > 0 && string(content[0]) != "<" {
					result += content + "\n"
				}
				return true
			})
		}
		return result
	}
}

func parse_file_by_extension(file_path string) string {
	var result string
	extension := filepath.Ext(file_path)
	switch extension {
	case ".txt", ".md":
		result = parse_text_file(file_path)
	case ".xml":
		result = parse_xml_file(file_path)
	case ".xhtml":
		fmt.Println("TODO: Parse XHTML documents")
	case ".html":
		fmt.Println("TODO: Parse HTML documents")
	case ".pdf":
		fmt.Println("TODO: Parse PDF documents")
	default:
		fmt.Printf("ERROR: Extension '%s' is not supported\n", extension)
	}
	return result
}
