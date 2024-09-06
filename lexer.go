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
func (l *lexer) left_trim() {
	for len(l.content) > 0 && (l.content[0] > unicode.MaxASCII || unicode.IsSpace(rune(l.content[0]))) {
		l.content = l.content[1:]
	}
}

func (l *lexer) chop(n int) string {
	token := l.content[0:n]
	l.content = l.content[n:]
	return token
}

func (l *lexer) chop_while(predicate func(rune) bool) string {
	n := 0
	for n < len(l.content) && predicate(rune(l.content[n])) && l.content[n] <= unicode.MaxASCII {
		n += 1
	}
	return l.chop(n)
}

func (l *lexer) next_token() (string, error) {
	l.left_trim()

	if len(l.content) == 0 {
		return "", errors.New("empty token")
	}

	if unicode.IsNumber(rune(l.content[0])) {
		return l.chop_while(unicode.IsNumber), nil
	}

	if unicode.IsLetter(rune(l.content[0])) {
		return strings.ToLower(l.chop_while(unicode.IsLetter)), nil
	}

	return l.chop(1), nil
}
