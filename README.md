# About

Go_search is an implementaion for a local search engine allowing searching within a variety of file types.

## Getting started

Clone the repository

```bash
git clone https://github.com/jamalkheirbeik/go_search
```

Create directory 'documents' within the project and place your (.txt | .md | .pdf | .html | .xml | .xhtml) files within it (you can include folders as file reading is recursive).

You can build the application and run the executable as below:

```bash
go build -tags "sqlite_math_functions"
# choose one based on your OS
./go_search.exe   # on windows
./go_search       # on linux or mac
```

The application will start a http server on localhost:8080 and populate the sqlite database concurrently.

API endpoints:

- GET <http://localhost:8080/> returns the main HTML page (to be implemented)
- GET <http://localhost:8080/search> accepts two params "query" and "page" and returns a list of files ordered by TF-IDF ranking. example usage <http://localhost:8080/search?query=hello&page=1>
