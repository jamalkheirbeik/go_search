package lexer

import (
	"errors"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
)

type Lexer struct {
	Content string
}

// trims whitespaces and non printable characters
func (l *Lexer) left_trim() {
	for len(l.Content) > 0 && (l.Content[0] > unicode.MaxASCII || unicode.IsSpace(rune(l.Content[0]))) {
		l.Content = l.Content[1:]
	}
}

func (l *Lexer) chop(n int) string {
	token := l.Content[0:n]
	l.Content = l.Content[n:]
	return token
}

func (l *Lexer) chop_while(predicate func(rune) bool) string {
	n := 0
	for n < len(l.Content) && predicate(rune(l.Content[n])) && l.Content[n] <= unicode.MaxASCII {
		n += 1
	}
	return l.chop(n)
}

func (l *Lexer) Next_token() (string, error) {
	l.left_trim()

	if len(l.Content) == 0 {
		return "", errors.New("empty token")
	}

	if unicode.IsNumber(rune(l.Content[0])) {
		return l.chop_while(unicode.IsNumber), nil
	}

	if unicode.IsLetter(rune(l.Content[0])) {
		word := strings.ToLower(l.chop_while(unicode.IsLetter))
		stemmed, err := snowball.Stem(word, "english", true)
		if err != nil {
			return word, nil
		}
		return stemmed, nil
	}

	return l.chop(1), nil
}
