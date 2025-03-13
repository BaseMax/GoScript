package main

import (
	"bufio"
	"io"
	"log"
	"strings"
	"unicode"
)

type TokenType string

type Token struct {
	Type  TokenType
	Value string
}

const (
	EOF       TokenType = "EOF"
	LPARENT   TokenType = "("
	RPARENT   TokenType = ")"
	LCURLY    TokenType = "{"
	RCURLY    TokenType = "}"
	LBRACKET  TokenType = "["
	RBRACKET  TokenType = "]"
	COMMA     TokenType = ","
	DOT       TokenType = "."
	PLUS      TokenType = "+"
	MINUS     TokenType = "-"
	STAR      TokenType = "*"
	SLASH     TokenType = "/"
	COLON     TokenType = ":"
	SEMICOLON TokenType = ";"
	NOT       TokenType = "!"
	QUESTION  TokenType = "?"
	EQ        TokenType = "="
	EQEQ      TokenType = "=="
	NEQ       TokenType = "!="
	GREATER   TokenType = ">"
	LESSER    TokenType = "<"
	GEQ       TokenType = ">="
	LEQ       TokenType = "<="
	STRING    TokenType = "STRING"
	INT       TokenType = "INT"
	FLOAT     TokenType = "FLOAT"
	IDENT     TokenType = "IDENT"
	TRUE      TokenType = "TRUE"
	FALSE     TokenType = "FALSE"
	IF        TokenType = "IF"
	ELSE      TokenType = "ELSE"
	FN        TokenType = "FN"
	PRINT     TokenType = "PRINT"
	PRINTLN   TokenType = "PRINTLN"
	RETURN    TokenType = "RETURN"
	FOR       TokenType = "FOR"
	DOTDOT    TokenType = ".."
	SWAP      TokenType = "SWAP"
	INPUT     TokenType = "INPUT"
	LEN       TokenType = "LEN"
	IMPORT    TokenType = "IMPORT"
	OR        TokenType = "OR"
	AND       TokenType = "AND"
)

var keywords = map[string]TokenType{
	"fn":      FN,
	"print":   PRINT,
	"println": PRINTLN,
	"return":  RETURN,
	"true":    TRUE,
	"false":   FALSE,
	"if":      IF,
	"else":    ELSE,
	"for":     FOR,
	"swap":    SWAP,
	"input":   INPUT,
	"len":     LEN,
	"import":  IMPORT,
	"or":      OR,
	"and":     AND,
}

var symbols = map[string]TokenType{
	"(":  LPARENT,
	")":  RPARENT,
	"{":  LCURLY,
	"}":  RCURLY,
	"[":  LBRACKET,
	"]":  RBRACKET,
	",":  COMMA,
	"..": DOTDOT,
	".":  DOT,
	"+":  PLUS,
	"-":  MINUS,
	"*":  STAR,
	"/":  SLASH,
	":":  COLON,
	"!":  NOT,
	"?":  QUESTION,
	"=":  EQ,
	"==": EQEQ,
	"!=": NEQ,
	">":  GREATER,
	">=": GEQ,
	"<":  LESSER,
	"<=": LEQ,
}

type Lexer struct {
	tokens chan Token
	reader *bufio.Reader
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		tokens: make(chan Token),
		reader: bufio.NewReader(strings.NewReader(input)),
	}
	go l.run()
	return l
}

func (l *Lexer) run() {
	defer close(l.tokens)
	for {
		r, _, err := l.reader.ReadRune()
		if err == io.EOF {
			l.emit(EOF, "")
			return
		}
		if unicode.IsSpace(r) {
			continue
		}
		switch {
		case r == '"':
			l.lexString()
		case unicode.IsDigit(r):
			l.lexNumber(r)
		case unicode.IsLetter(r):
			l.lexIdentifier(r)
		default:
			l.lexSymbol(r)
		}
	}
}

func (l *Lexer) emit(t TokenType, value string) {
	l.tokens <- Token{Type: t, Value: value}
}

func (l *Lexer) lexString() {
	var sb strings.Builder
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			log.Fatal("Unexpected EOF in string")
		}
		if r == '"' {
			break
		}
		if r == '\\' {
			next, _, err := l.reader.ReadRune()
			if err != nil {
				log.Fatal("Unexpected EOF in escape sequence")
			}
			if next == '"' {
				sb.WriteRune('"')
			} else {
				sb.WriteRune('\\')
				sb.WriteRune(next)
			}
		} else {
			sb.WriteRune(r)
		}
	}
	l.emit(STRING, sb.String())
}

func (l *Lexer) lexIdentifier(first rune) {
	var sb strings.Builder
	sb.WriteRune(first)
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil || !isIdentChar(r) {
			if err != io.EOF {
				l.reader.UnreadRune()
			}
			break
		}
		sb.WriteRune(r)
	}
	ident := sb.String()
	if t, ok := keywords[ident]; ok {
		l.emit(t, ident)
	} else {
		l.emit(IDENT, ident)
	}
}

func isIdentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (l *Lexer) lexNumber(first rune) {
	var sb strings.Builder
	sb.WriteRune(first)
	t := INT
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil || (!unicode.IsDigit(r) && r != '.') {
			if err != io.EOF {
				l.reader.UnreadRune()
			}
			break
		}
		if r == '.' {
			if t == FLOAT {
				l.reader.UnreadRune()
				break
			}
			t = FLOAT
		}
		sb.WriteRune(r)
	}
	l.emit(t, sb.String())
}

func (l *Lexer) lexSymbol(first rune) {
	single := string(first)
	if single == "/" {
		if l.lexComment() {
			return
		}
	}
	double := single
	if next, _, err := l.reader.ReadRune(); err == nil {
		double += string(next)
		if t, ok := symbols[double]; ok {
			l.emit(t, double)
			return
		}
		l.reader.UnreadRune()
	}
	if t, ok := symbols[single]; ok {
		l.emit(t, single)
	} else {
		log.Fatalf("Unknown symbol: %s", single)
	}
}

func (l *Lexer) lexComment() bool {
	next, _, err := l.reader.ReadRune()
	if err != nil {
		l.reader.UnreadRune()
		return false
	}
	switch next {
	case '/':
		l.reader.ReadLine()
		return true
	case '*':
		for {
			r, _, err := l.reader.ReadRune()
			if err != nil {
				log.Fatal("Unclosed multi-line comment")
			}
			if r == '*' {
				if next, _, err := l.reader.ReadRune(); err == nil && next == '/' {
					return true
				}
				l.reader.UnreadRune()
			}
		}
	}
	l.reader.UnreadRune()
	return false
}
