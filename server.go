package main

import (
	"fmt"
	"net/http"
	"strconv"
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
			query := r.FormValue("query")
			page, _ := strconv.Atoi(r.FormValue("page"))
			data, err := s.db.search(query, page)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, err)
			} else {
				fmt.Fprint(w, data)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 not found")
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 not found")
	}
}

func (s *server) serve(port string) {
	fmt.Printf("INFO: Server running on http://localhost:%s\n", port)
	http.HandleFunc("/", s.handle_request)
	http.ListenAndServe(port, nil)
}
