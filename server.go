package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type server struct {
	db *database
}

func NewServer(db *database) *server {
	s := server{db: db}
	return &s
}

func (s *server) handle_request(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/":
			fmt.Fprintf(w, "GET: index")
		case "/search":
			query := strings.Trim(r.FormValue("query"), " \r\n\t")
			page, _ := strconv.Atoi(r.FormValue("page"))
			if len(query) == 0 {
				fmt.Fprint(w, "Search query cannot be empty")
				break
			}
			fmt.Fprint(w, s.db.search(query, page))
		default:
			fmt.Fprintf(w, "404 not found")
		}
	default:
		fmt.Fprintf(w, "404 not found")
	}
}

func (s *server) serve(port string) {
	fmt.Printf("INFO: Server running on http://localhost:%s\n", port)
	http.HandleFunc("/", s.handle_request)
	http.ListenAndServe(port, nil)
}
