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
		lexemes: make(chan Lexeme, 256),
	}
	go s.scanTokens()
	return s
}

func (s *scanner) scanTokens() {
	for {
		r, _, err := s.rdr.ReadRune()
		if err == io.EOF {
			s.sendToken(END_OF_FILE, "")
			close(s.lexemes)
			break
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
	}
}

func (s *scanner) scanMultiLineComment() {
	for {
		_, err := s.rdr.ReadString('*')
		if err != nil {
			log.Fatal(err)
		}
		nextRune, _, err := s.rdr.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		if nextRune == '/' {
			break
		}
		s.rdr.UnreadRune()
	}
}

func (s *scanner) sendToken(kind TokKind, txt string) {
	s.lexemes <- Lexeme{Kind: kind, Text: txt}
}

func (s *scanner) scanString() {
	var builder strings.Builder
	for {
		r, _, err := s.rdr.ReadRune()
		if err == io.EOF {
			break
		}
		if r == '"' {
			current := builder.String()
			if len(current) > 0 && current[len(current)-1] == '\\' {
				builder.WriteRune(r)
				continue
			}
			break
		}
		builder.WriteRune(r)
	}
	result := strings.ReplaceAll(builder.String(), `\"`, `"`)
	s.sendToken(STRING_T, result)
}

func (s *scanner) scanIdentifier(initial rune) {
	var builder strings.Builder
	builder.WriteRune(initial)
	for {
		r, _, err := s.rdr.ReadRune()
		if err != nil {
			break
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		} else {
			s.rdr.UnreadRune()
			break
		}
	}
	text := builder.String()
	if kw, ok := reservedWords[text]; ok {
		s.sendToken(kw, text)
	} else {
		s.sendToken(IDENTIFIER, text)
	}
}

func (s *scanner) scanNumber(initial rune) {
	tokType := INTEGER_T
	var builder strings.Builder
	builder.WriteRune(initial)
	for {
		r, _, err := s.rdr.ReadRune()
		if err != nil {
			break
		}
		if r == '.' {
			nextRune, _, err := s.rdr.ReadRune()
			if err == nil && nextRune == '.' {
				s.sendToken(tokType, builder.String())
				s.sendToken(DOTDOT_SYM, "..")
				return
			} else if err == nil {
				s.rdr.UnreadRune()
			}
			tokType = FLOAT_T
			builder.WriteRune(r)
		} else if unicode.IsDigit(r) {
			builder.WriteRune(r)
		} else {
			s.rdr.UnreadRune()
			break
		}
	}
	s.sendToken(tokType, builder.String())
}

func (s *scanner) scanSymbol(r rune) {
	single := string(r)
	if peek, err := s.rdr.Peek(1); err == nil {
		double := single + string(peek)
		if t, ok := symbolMap[double]; ok {
			s.rdr.ReadRune()
			s.sendToken(t, double)
			return
		}
		if double == "//" {
			s.rdr.ReadString('\n')
			return
		}
		if double == "/*" {
			s.rdr.ReadRune()
			s.scanMultiLineComment()
			return
		}
	}
	if t, ok := symbolMap[single]; ok {
		s.sendToken(t, single)
	}
}
