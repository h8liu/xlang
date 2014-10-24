package parser

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Lexer is a token scanner
type Lexer struct {
	c    *cursor
	last *Tok
	hold *Tok
}

// NewLexer creates a lexer for an io stream.
// It will close the stream automatically when the lexing ends.
func NewLexer(file string, r io.ReadCloser) *Lexer {
	ret := new(Lexer)
	ret.c = newCursor(file, r)

	return ret
}

func (lex *Lexer) emitTok(t *Tok) {
	lex.hold = t
	if lex.hold.Type != TypeComment {
		lex.last = t
	}
}

func (lex *Lexer) emitType(t Type) {
	tok := lex.c.Token(t)
	lex.emitTok(tok)
}

func (lex *Lexer) skipEndl() bool {
	if lex.last == nil {
		return true
	}
	t := lex.last.Type

	switch t {
	case TypeIdent, TypeInt, TypeFloat, TypeString:
		return false
	case TypeOperator:
		lit := lex.last.Lit
		return !(lit == "}" || lit == "]" || lit == ")")
	}

	return true
}

func (lex *Lexer) scanInt() {
	for lex.c.Scan() {
		r := lex.c.Next()
		if !isDigit(r) {
			break
		}
	}

	lex.emitType(TypeInt)
}

func (lex *Lexer) scanIdent() {
	for lex.c.Scan() {
		r := lex.c.Next()
		if !isDigit(r) && !isLetter(r) {
			break
		}
	}

	lex.emitType(TypeIdent)
}

func (lex *Lexer) scanLineComment() {
	for lex.c.Scan() {
		r := lex.c.Next()
		if r == '\n' {
			break
		}
	}

	lex.emitType(TypeComment)
}

func (lex *Lexer) scanBlockComment() {
	var star bool
	var complete bool

	for lex.c.Scan() {
		r := lex.c.Next()
		if star && r == '/' {
			lex.c.Accept()
			complete = true
			break
		}

		star = r == '*'
	}

	lex.emitType(TypeComment)
	if !complete {
		// TODO: report imcomplete block comment parsing error
	}
}

func (lex *Lexer) isWhite(r rune) bool {
	if isWhite(r) {
		return true
	}

	if r == '\n' {
		return lex.skipEndl()
	}

	return false
}

func (lex *Lexer) skipWhite() {
	if lex.c.EOF() || !lex.isWhite(lex.c.Next()) {
		return // no white to skip
	}

	for lex.c.Scan() {
		r := lex.c.Next()
		if !lex.isWhite(r) {
			break
		}
	}

	lex.c.Discard()
}

func (lex *Lexer) scanOperator() {
	r := lex.c.Next()

	if lex.c.Scan() && r == '/' {
		r2 := lex.c.Next()

		if r2 == '/' {
			lex.scanLineComment()
			return
		} else if r2 == '*' {
			lex.scanBlockComment()
			return
		}
	}

	if r == '\n' && lex.skipEndl() {
		panic("bug")
	}

	lex.emitType(TypeOperator)
}

func (lex *Lexer) scanInvalid() {
	lex.c.Accept()
	lex.emitType(TypeInvalid)
}

// Scan returns true where there is a new token
func (lex *Lexer) Scan() bool {
	lex.skipWhite()

	if lex.c.EOF() {
		return false
	}
	r := lex.c.Next()

	if isDigit(r) {
		lex.scanInt()
	} else if isLetter(r) {
		lex.scanIdent()
	} else if isOperator(r) {
		lex.scanOperator()
	} else {
		lex.scanInvalid()
	}

	return true
}

// Token returns the current token.
func (lex *Lexer) Token() *Tok {
	return lex.hold
}

// IOErr returns the IO error on scanning.
func (lex *Lexer) IOErr() error {
	return lex.c.Err()
}

// LexFile creates a lexer over a file.
func LexFile(path string) (*Lexer, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}

	return NewLexer(path, f), nil
}

// LexString creates a lexer over a string.
func LexString(file, s string) *Lexer {
	r := ioutil.NopCloser(strings.NewReader(s))
	return NewLexer(file, r)
}
