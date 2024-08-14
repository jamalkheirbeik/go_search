package main

import (
	"errors"
	"strings"
	"unicode"
)

type lexer struct {
	content string
}

// trims whitespaces and non printable characters
func left_trim(lexer *lexer) {
	for len(lexer.content) > 0 && (lexer.content[0] > unicode.MaxASCII || unicode.IsSpace(rune(lexer.content[0]))) {
		lexer.content = lexer.content[1:]
	}
}

func chop(lexer *lexer, n int) string {
	token := lexer.content[0:n]
	lexer.content = lexer.content[n:]
	return token
}

func chop_while(lexer *lexer, predicate func(rune) bool) string {
	n := 0
	for n < len(lexer.content) && predicate(rune(lexer.content[n])) && lexer.content[n] <= unicode.MaxASCII {
		n += 1
	}
	return chop(lexer, n)
}

func next_token(lexer *lexer) (string, error) {
	left_trim(lexer)

	if len(lexer.content) == 0 {
		return "", errors.New("empty token")
	}

	if unicode.IsNumber(rune(lexer.content[0])) {
		return chop_while(lexer, unicode.IsNumber), nil
	}

	if unicode.IsLetter(rune(lexer.content[0])) {
		return strings.ToLower(chop_while(lexer, unicode.IsLetter)), nil
	}

	return chop(lexer, 1), nil
}
