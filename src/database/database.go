package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jamalkheirbeik/go_search/src/lexer"
	"github.com/jamalkheirbeik/go_search/src/parser"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	conn *sql.DB
}

type Document struct {
	id            int64
	file_path     string
	entries       int64
	last_modified time.Time
}

type Term struct {
	id   int64
	term string
	df   int64
}

type SearchResult struct {
	Pages int64       `json:"pages"`
	Data  []SearchRow `json:"data"`
}

type SearchRow struct {
	File_path string `json:"file_path"`
	ranking   float64
	pages     int64
}

func NewDB() *Database {
	if err := os.MkdirAll("./storage", os.ModePerm); err != nil {
		fmt.Printf("ERROR: Cannot create directory 'storage'. %s\n", err)
		os.Exit(1)
	}
	const db_path = "./storage/go_search.db"
	db := Database{}
	conn, err := sql.Open("sqlite3", db_path)
	if err != nil {
		fmt.Printf("ERROR: Cannot open database '%s'. %s\n", db_path, err)
		os.Exit(1)
	}
	conn.Exec("PRAGMA journal_mode = WAL;")
	db.conn = conn
	return &db
}

func (db *Database) Disconnect() {
	err := db.conn.Close()
	if err != nil {
		fmt.Printf("ERROR: Cannot close database connection. %s\n", err)
	}
}

func (db *Database) Generate_tables() {
	d_query := "CREATE TABLE IF NOT EXISTS documents (id INTEGER PRIMARY KEY, file_path VARCHAR UNIQUE NOT NULL, entries INTEGER DEFAULT 0, last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"
	d_stmt, _ := db.conn.Prepare(d_query)
	defer d_stmt.Close()
	if _, err := d_stmt.Exec(); err != nil {
		fmt.Printf("ERROR: Cannot create table 'documents'. %s\n", err)
		os.Exit(1)
	}

	t_query := "CREATE TABLE IF NOT EXISTS terms (id INTEGER PRIMARY KEY, term VARCHAR UNIQUE NOT NULL, df INTEGER DEFAULT 0)"
	t_stmt, _ := db.conn.Prepare(t_query)
	defer t_stmt.Close()
	if _, err := t_stmt.Exec(); err != nil {
		fmt.Printf("ERROR: Cannot create table 'terms'. %s\n", err)
		os.Exit(1)
	}

	td_query := "CREATE TABLE IF NOT EXISTS term_doc (t_id INTERGER, d_id INTEGER, frequency INTEGER DEFAULT 1, UNIQUE(t_id, d_id), FOREIGN KEY(t_id) REFERENCES terms(id), FOREIGN KEY(d_id) REFERENCES documents(id))"
	td_stmt, _ := db.conn.Prepare(td_query)
	defer td_stmt.Close()
	if _, err := td_stmt.Exec(); err != nil {
		fmt.Printf("ERROR: Cannot create table 'term_doc'. %s\n", err)
		os.Exit(1)
	}

	d_idx_query := "CREATE INDEX IF NOT EXISTS document_id ON term_doc(d_id)"
	if _, err := db.conn.Exec(d_idx_query); err != nil {
		fmt.Printf("ERROR: Cannot create index 'document_id' on table 'term_doc'. %s\n", err)
	}

	t_idx_query := "CREATE INDEX IF NOT EXISTS term_id ON term_doc(t_id)"
	if _, err := db.conn.Exec(t_idx_query); err != nil {
		fmt.Printf("ERROR: Cannot create index 'term_id' on table 'term_doc'. %s\n", err)
	}
}

func (db *Database) Add_dir_files(directory string) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("ERROR: Cannot read files in directory '%s'. %s\n", directory, err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			db.Add_dir_files(directory + "/" + entry.Name())
		} else {
			file_path := directory + "/" + entry.Name()
			f := parser.NewFile(file_path)
			fmt.Printf("INFO: Indexing file '%s'...\n", f.Path)
			document := db.get_document(f)

			if f.Last_modified.After(document.last_modified) {
				db.soft_delete_document_cascade(f)
				db.add_document_and_terms(f)
			}
		}
	}
}

func (db *Database) get_document(f *parser.File) Document {
	query := "SELECT * FROM documents WHERE file_path = ?"
	row := db.conn.QueryRow(query, f.Path)
	var d Document
	row.Scan(&d.id, &d.file_path, &d.entries, &d.last_modified)
	return d
}

func (db *Database) get_term(term string) Term {
	query := "SELECT * FROM terms WHERE term = ?"
	row := db.conn.QueryRow(query, term)
	var t Term
	row.Scan(&t.id, &t.term, &t.df)
	return t
}

func (db *Database) insert_term(term string) int64 {
	t := db.get_term(term)
	if t.id > 0 {
		return t.id
	}
	query := "INSERT INTO terms(term) VALUES(?)"
	stmt, _ := db.conn.Prepare(query)
	defer stmt.Close()
	res, _ := stmt.Exec(term)
	id, _ := res.LastInsertId()
	return id
}

func (db *Database) insert_document(f *parser.File) int64 {
	d := db.get_document(f)
	if d.id > 0 {
		return d.id
	}
	query := "INSERT INTO documents(file_path, last_modified) VALUES(?, ?)"
	stmt, _ := db.conn.Prepare(query)
	defer stmt.Close()
	res, _ := stmt.Exec(f.Path, f.Last_modified)
	id, _ := res.LastInsertId()
	return id
}

func (db *Database) insert_term_doc(term_id int64, doc_id int64) {
	query := `INSERT INTO term_doc(t_id, d_id) VALUES(?, ?)
		ON CONFLICT(t_id, d_id) DO UPDATE SET frequency = frequency + 1`
	stmt, _ := db.conn.Prepare(query)
	defer stmt.Close()
	stmt.Exec(term_id, doc_id)
}

func (db *Database) update_document_entries(id int64, entries int64) {
	query := "UPDATE documents SET entries = ? WHERE id = ?"
	stmt, _ := db.conn.Prepare(query)
	defer stmt.Close()
	stmt.Exec(entries, id)
}

func (db *Database) update_terms_df(id int64) {
	query := "UPDATE terms SET df = df + 1 WHERE id IN (SELECT t_id FROM term_doc WHERE d_id = ? AND frequency > 0)"
	stmt, _ := db.conn.Prepare(query)
	defer stmt.Close()
	stmt.Exec(id)
}

func (db *Database) soft_delete_document_cascade(f *parser.File) {
	d := db.get_document(f)
	if d.id == 0 {
		return
	}
	t_query := "UPDATE terms SET df = df - 1 WHERE id IN (SELECT t_id FROM term_doc WHERE d_id = ?)"
	t_stmt, _ := db.conn.Prepare(t_query)
	defer t_stmt.Close()
	t_stmt.Exec(d.id)

	d_query := "UPDATE documents SET entries = 0 WHERE id = ?"
	d_stmt, _ := db.conn.Prepare(d_query)
	defer d_stmt.Close()
	d_stmt.Exec(d.id)

	td_query := "UPDATE term_doc SET frequency = 0 WHERE d_id = ?"
	td_stmt, _ := db.conn.Prepare(td_query)
	defer td_stmt.Close()
	td_stmt.Exec(d.id)
}

func (db *Database) add_document_and_terms(f *parser.File) {
	l := lexer.Lexer{Content: f.Parse()}
	if len(l.Content) == 0 {
		return
	}
	d_id := db.insert_document(f)
	var entries int64 = 0
	for len(l.Content) > 0 {
		token, err := l.Next_token()
		if err != nil {
			continue
		}
		t_id := db.insert_term(token)
		db.insert_term_doc(t_id, d_id)
		entries++
	}
	db.update_document_entries(d_id, entries)
	db.update_terms_df(d_id)
}

func (db *Database) Search(query string, page_number int) (SearchResult, error) {
	const LIMIT = 10
	result := SearchResult{Pages: 0, Data: make([]SearchRow, 0)}
	offset := (page_number - 1) * LIMIT

	var terms string
	l := lexer.Lexer{Content: query}
	for len(l.Content) > 0 {
		token, err := l.Next_token()
		if err != nil {
			continue
		}
		if len(terms) > 0 {
			terms += ","
		}
		terms += "'" + token + "'"
	}

	if len(terms) == 0 {
		return result, errors.New("search query cannot be empty")
	}

	total_docs_query := "SELECT COUNT(*) FROM documents"
	row := db.conn.QueryRow(total_docs_query)
	var total_docs int64
	row.Scan(&total_docs)

	sql_query := fmt.Sprintf(`
	SELECT	d.file_path,
			SUM(CAST(td.frequency AS FLOAT) / d.entries * LOG(%d / t.df)) ranking,
			CAST(CEIL(COUNT(*) OVER () / CAST(%d AS FLOAT)) AS INTERGER) pages
	FROM terms t
	INNER JOIN term_doc td ON t.id = td.t_id
	INNER JOIN documents d ON td.d_id = d.id
	WHERE t.term in (%s) AND td.frequency > 0
	GROUP BY d.file_path
	ORDER BY ranking DESC
	LIMIT %d OFFSET %d
	`, total_docs, LIMIT, terms, LIMIT, offset)

	rows, _ := db.conn.Query(sql_query)
	defer rows.Close()

	for rows.Next() {
		var row SearchRow
		err := rows.Scan(&row.File_path, &row.ranking, &row.pages)
		if err != nil {
			panic(err)
		}
		fmt.Printf("    %s => %f\n", row.File_path, row.ranking)
		result.Pages = row.pages
		result.Data = append(result.Data, row)
	}
	return result, nil
}
