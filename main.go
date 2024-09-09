package main

import (
	"github.com/jamalkheirbeik/go_search/src/database"
	"github.com/jamalkheirbeik/go_search/src/server"
)

func main() {
	db := database.NewDB()
	db.Generate_tables()
	defer db.Disconnect()
	go db.Add_dir_files("documents")
	s := server.NewServer(db, ":8080")
	s.Serve()
}
