package main

func main() {
	db := NewDB()
	db.generate_tables()
	defer db.disconnect()
	go db.add_dir_files("documents")
	s := NewServer(db)
	s.serve(":8080")
}
