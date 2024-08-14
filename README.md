# About

Go_search is an implementaion for a local search engine allowing searching within a variety of file types.

## Getting started

Clone the repository

```bash
git clone https://github.com/jamalkheirbeik/go_search
```

Make sure to provide a directory containing (.txt | .md | .pdf | .html | .xml | .xhtml) files to be able to test the application.

Note: Check main.go file to see all supported file types.

To run the application you can either directly run the main.go file as below:

```bash
go run .
```

or build the application and run the executable as below:

```bash
go build .
# choose one based on your OS
./go_search.exe   # on windows
./go_search       # on linux or mac
```

Either way you choose you will encounter an error because we did not provide a subcommand yet.

To list all the available commands simply run one the following:

```bash
go run . help
./go_search.exe help   # on windows
./go_search help       # on linux or mac
```
