package lexer

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type lexer struct {
	file string
	r    io.ReadCloser
}

func open(path string) (*lexer, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}

	return newLexer(path, f), nil
}

func lex(file, s string) *lexer {
	r := ioutil.NopCloser(strings.NewReader(s))
	return newLexer(file, r)
}

func newLexer(file string, r io.ReadCloser) *lexer {
	ret := new(lexer)
	ret.file = file
	ret.r = r

	return ret
}
