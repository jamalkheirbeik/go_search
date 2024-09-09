package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jamalkheirbeik/go_search/src/database"
)

type server struct {
	db   *database.Database
	port string
}

func NewServer(db *database.Database, port string) *server {
	s := server{db: db, port: port}
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
			data, err := s.db.Search(query, page)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, err)
			} else {
				b, err := json.Marshal(data)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, "500 Internal server error")
				} else {
					fmt.Fprint(w, string(b))
				}
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

func (s *server) Serve() {
	fmt.Printf("INFO: Server running on http://localhost:%s\n", s.port)
	http.HandleFunc("/", s.handle_request)
	http.ListenAndServe(s.port, nil)
}
