package main

import (
	"bufio"
	"io"
	"log"
	"strings"
	"unicode"
)

type TokKind string

type Lexeme struct {
	Kind TokKind
	Text string
}

type scanner struct {
	lexemes chan Lexeme
	rdr     *bufio.Reader
}

const (
	END_OF_FILE   TokKind = "EOF"
	OPEN_PAREN    TokKind = "("
	CLOSE_PAREN   TokKind = ")"
	OPEN_CURLY    TokKind = "{"
	CLOSE_CURLY   TokKind = "}"
	OPEN_BRACKET  TokKind = "["
	CLOSE_BRACKET TokKind = "]"
	COMMA_SYM     TokKind = ","
	DOT_SYM       TokKind = "."
	PLUS_SYM      TokKind = "+"
	MINUS_SYM     TokKind = "-"
	MULTIPLY_SYM  TokKind = "*"
	DIVIDE_SYM    TokKind = "/"
	COLON_SYM     TokKind = ":"
	SEMICOLON_SYM TokKind = ";"
	EXCLAMATION   TokKind = "!"
	QUESTION_MARK TokKind = "?"
	ASSIGN        TokKind = "="
	EQ_OP         TokKind = "=="
	NEQ_OP        TokKind = "!="
	GREATER_THAN  TokKind = ">"
	LESS_THAN     TokKind = "<"
	GREATER_EQ    TokKind = ">="
	LESS_EQ       TokKind = "<="
	STRING_T      TokKind = "STRING"
	INTEGER_T     TokKind = "INT"
	FLOAT_T       TokKind = "FLOAT"
	IDENTIFIER    TokKind = "IDENT"
	ERROR_T       TokKind = "ERROR"
	TRUE_T        TokKind = "TRUE"
	FALSE_T       TokKind = "FALSE"
	IF_T          TokKind = "IF"
	ELSE_T        TokKind = "ELSE"
	FUNCTION_T    TokKind = "FN"
	PRINT_T       TokKind = "PRINT"
	PRINTLN_T     TokKind = "PRINTLN"
	RETURN_T      TokKind = "RETURN"
	FOR_T         TokKind = "FOR"
	DOTDOT_SYM    TokKind = ".."
	SWAP_T        TokKind = "SWAP"
	INPUT_T       TokKind = "INPUT"
	LENGTH_T      TokKind = "LEN"
	IMPORT_T      TokKind = "IMPORT"
	OR_T          TokKind = "OR"
	AND_T         TokKind = "AND"
)

var reservedWords = map[string]TokKind{
	"fn":      FUNCTION_T,
	"print":   PRINT_T,
	"println": PRINTLN_T,
	"return":  RETURN_T,
	"true":    TRUE_T,
	"false":   FALSE_T,
	"if":      IF_T,
	"else":    ELSE_T,
	"for":     FOR_T,
	"swap":    SWAP_T,
	"input":   INPUT_T,
	"len":     LENGTH_T,
	"import":  IMPORT_T,
	"or":      OR_T,
	"and":     AND_T,
}

var symbolMap = map[string]TokKind{
	"(":  OPEN_PAREN,
	")":  CLOSE_PAREN,
	"{":  OPEN_CURLY,
	"}":  CLOSE_CURLY,
	"[":  OPEN_BRACKET,
	"]":  CLOSE_BRACKET,
	",":  COMMA_SYM,
	"..": DOTDOT_SYM,
	".":  DOT_SYM,
	"+":  PLUS_SYM,
	"-":  MINUS_SYM,
	"*":  MULTIPLY_SYM,
	"/":  DIVIDE_SYM,
	":":  COLON_SYM,
	"!":  EXCLAMATION,
	"?":  QUESTION_MARK,
	"=":  ASSIGN,
	"==": EQ_OP,
	"!=": NEQ_OP,
	">":  GREATER_THAN,
	">=": GREATER_EQ,
	"<":  LESS_THAN,
	"<=": LESS_EQ,
}

func CreateScanner(src string) *scanner {
	src = strings.NewReplacer(`\n`, "\n", `\t`, "\t", `\r`, "\r").Replace(src)

	rdr := bufio.NewReader(strings.NewReader(src))
	s := &scanner{
		rdr:     rdr,
		lexemes: make(chan Lexeme),
	}
	go s.scanTokens()
	return s
}

func (s *scanner) scanTokens() {
	r, _, err := s.rdr.ReadRune()
	if err == io.EOF {
		s.sendToken(END_OF_FILE, "")
		close(s.lexemes)
		return
	}
	switch {
	case r == '"':
		s.scanString()
	case unicode.IsDigit(r):
		s.scanNumber(r)
	case unicode.IsLetter(r):
		s.scanIdentifier(r)
	default:
		s.scanSymbol(r)
	}
	s.scanTokens()
}

func (s *scanner) scanMultiLineComment() {
	for {
		str, err := s.rdr.ReadString('*')
		_ = str
		if err != nil {
			log.Fatal(err)
			return
		}
		nextChar, _, err := s.rdr.ReadRune()
		if err != nil {
			log.Fatal(err)
			return
		}
		if nextChar == '/' {
			break
		}
		s.rdr.UnreadRune()
	}
}

func (s *scanner) sendToken(kind TokKind, txt string) {
	token := Lexeme{Kind: kind, Text: txt}
	s.lexemes <- token
}

func (s *scanner) scanString() {
	accumulated, _ := s.rdr.ReadString('"')
	str := accumulated
	for strings.HasSuffix(str, "\\\"") {
		str, _ = s.rdr.ReadString('"')
		accumulated += str
	}
	accumulated = strings.TrimRight(accumulated, `"`)
	accumulated = strings.ReplaceAll(accumulated, "\\\"", "\"")
	s.sendToken(STRING_T, accumulated)
}

func (s *scanner) scanIdentifier(r rune) {
	var text string
	for {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			text += string(r)
		} else {
			s.rdr.UnreadRune()
			break
		}
		r, _, _ = s.rdr.ReadRune()
	}

	if kw, ok := reservedWords[text]; ok {
		s.sendToken(kw, text)
	} else {
		s.sendToken(IDENTIFIER, text)
	}
}

func (s *scanner) scanNumber(r rune) {
	tokType := INTEGER_T
	var text string
	for {
		if r == '.' {
			r2, _, _ := s.rdr.ReadRune()
			if r2 == '.' {
				s.sendToken(INTEGER_T, text)
				s.sendToken(DOTDOT_SYM, "..")
				return
			} else {
				s.rdr.UnreadRune()
			}
			tokType = FLOAT_T
		}
		if unicode.IsDigit(r) || r == '.' {
			text += string(r)
		} else {
			s.rdr.UnreadRune()
			break
		}
		r, _, _ = s.rdr.ReadRune()
	}
	s.sendToken(tokType, text)
}

func (s *scanner) scanSymbol(r rune) {
	single := string(r)

	r, _, _ = s.rdr.ReadRune()
	double := single + string(r)
	if t, ok := symbolMap[double]; ok {
		s.sendToken(t, double)
		return
	}
	if double == "//" {
		s.rdr.ReadLine()
		return
	}
	if double == "/*" {
		s.scanMultiLineComment()
		return
	}
	if t, ok := symbolMap[single]; ok {
		s.rdr.UnreadRune()
		s.sendToken(t, single)
		return
	}
	s.rdr.UnreadRune()
}
