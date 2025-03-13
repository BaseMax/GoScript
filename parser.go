package main

import (
	"strconv"
)

const (
	LOWEST_PREC = iota + 1
	PREC_EQUALS
	PREC_LESSGREATER
	PREC_SUM
	PREC_PRODUCT
	PREC_PREFIX
	PREC_CALL
	PREC_INDEX
	PREC_RANGE
)

var precedences = map[TokKind]int{
	ASSIGN:       PREC_EQUALS,
	EQ_OP:        PREC_EQUALS,
	NEQ_OP:       PREC_EQUALS,
	LESS_THAN:    PREC_LESSGREATER,
	GREATER_THAN: PREC_LESSGREATER,
	LESS_EQ:      PREC_LESSGREATER,
	GREATER_EQ:   PREC_LESSGREATER,
	PLUS_SYM:     PREC_SUM,
	MINUS_SYM:    PREC_SUM,
	DIVIDE_SYM:   PREC_PRODUCT,
	MULTIPLY_SYM: PREC_PRODUCT,
	OPEN_PAREN:   PREC_CALL,
	OPEN_BRACKET: PREC_INDEX,
	DOTDOT_SYM:   PREC_RANGE,
}

type Node interface {
	Evaluate(env *Environment) any
}

type prefixParseFunc func() Node

type infixParseFunc func(Node) Node

type ArrayLiteral struct {
	Elements []Node
}

type StringLiteral struct {
	Value string
}

type Ident struct {
	Lexeme Lexeme
	IsFunc bool
}

type IntegerLiteral struct {
	Lexeme Lexeme
	Value  int
}

type FloatLiteral struct {
	Lexeme Lexeme
	Value  float64
}
type BooleanLiteral struct {
	Lexeme Lexeme
	Value  bool
}

type ReturnStmt struct {
	Expr Node
}
type VarAssign struct {
	Name  Node
	Value Node
}
type PrefixOp struct {
	Lexeme Lexeme
	Expr   Node
}

type InfixOp struct {
	Lexeme Lexeme
	Left   Node
	Right  Node
}

type BlockStmt struct {
	Stmts []Node
}

type IfStmt struct {
	Condition Node
	Then      *BlockStmt
	Else      *BlockStmt
}

type ForStmt struct {
	Key    *Ident
	Value  *Ident
	Target Node
	Body   *BlockStmt
}

type RangeExpr struct {
	From Node
	To   Node
	Step Node
}

type PrintStmt struct {
	Args    []Node
	NewLine bool
}

type IndexExpr struct {
	Collection Node
	Index      Node
}

type MapLiteral struct {
	Pairs map[Node]Node
}

type FunctionLiteral struct {
	Name   string
	Params []*Ident
	Body   *BlockStmt
	Env    *Environment
}

type CallExpr struct {
	Function Node
	Args     []Node
}

type SwapStmt struct {
	A Node
	B Node
}

type ImportStmt struct {
	File Node
}

type InputStmt struct {
	Prompt Node
}

type LengthExpr struct {
	Target Node
}

type analyzer struct {
	astNodes      chan Node
	lexemes       chan Lexeme
	curLex        *Lexeme
	nxtLex        *Lexeme
	prefixParsers map[TokKind]prefixParseFunc
	infixParsers  map[TokKind]infixParseFunc
}

func CreateParser(lexemes chan Lexeme) *analyzer {
	a := &analyzer{
		lexemes:      lexemes,
		infixParsers: make(map[TokKind]infixParseFunc),
		astNodes:     make(chan Node),
		curLex:       &Lexeme{},
		nxtLex:       &Lexeme{},
	}

	a.prefixParsers = map[TokKind]prefixParseFunc{
		IDENTIFIER:   a.parseIdent,
		STRING_T:     a.parseStr,
		INTEGER_T:    a.parseInteger,
		FLOAT_T:      a.parseFloating,
		MINUS_SYM:    a.parsePrefixOperator,
		EXCLAMATION:  a.parsePrefixOperator,
		TRUE_T:       a.parseBool,
		FALSE_T:      a.parseBool,
		OPEN_PAREN:   a.parseGroup,
		IF_T:         a.parseIf,
		FUNCTION_T:   a.parseFunction,
		PRINT_T:      a.parsePrint,
		PRINTLN_T:    a.parsePrint,
		OPEN_BRACKET: a.parseArray,
		OPEN_CURLY:   a.parseMap,
		FOR_T:        a.parseFor,
		RETURN_T:     a.parseRet,
		SWAP_T:       a.parseSwap,
		INPUT_T:      a.parseInput,
		LENGTH_T:     a.parseLen,
		IMPORT_T:     a.parseImport,
	}

	for _, kind := range []TokKind{OR_T, AND_T, PLUS_SYM, MINUS_SYM, MULTIPLY_SYM, DIVIDE_SYM, EQ_OP, NEQ_OP, GREATER_THAN, GREATER_EQ, LESS_THAN, LESS_EQ} {
		a.infixParsers[kind] = a.parseInfixOperator
	}

	a.infixParsers[OPEN_PAREN] = a.parseCall
	a.infixParsers[OPEN_BRACKET] = a.parseIndex
	a.infixParsers[DOTDOT_SYM] = a.parseRange
	a.infixParsers[ASSIGN] = a.parseVarAssign

	*a.curLex = <-a.lexemes
	*a.nxtLex = <-a.lexemes

	go a.processParsing()

	return a
}

func (a *analyzer) getPrecedence(kind TokKind) int {
	if prec, ok := precedences[kind]; ok {
		return prec
	}
	return LOWEST_PREC
}

func (a *analyzer) checkNext(expected TokKind) bool {
	if a.nxtLex.Kind == expected {
		a.advance()
		return true
	}
	return false
}

func (a *analyzer) advance() {
	a.curLex = a.nxtLex
	next, ok := <-a.lexemes
	if !ok {
		a.nxtLex = &Lexeme{Kind: END_OF_FILE, Text: ""}
	} else {
		a.nxtLex = &next
	}
}

func (a *analyzer) processParsing() {
	for a.curLex.Kind != END_OF_FILE {
		node := a.parseExpr(LOWEST_PREC)
		if node != nil {
			a.astNodes <- node
		}
		a.advance()
	}
	close(a.astNodes)
}

func (a *analyzer) parseExpr(prec int) Node {
	var left Node
	if prefix, ok := a.prefixParsers[a.curLex.Kind]; ok {
		left = prefix()
	}
	nextPrec := a.getPrecedence(a.nxtLex.Kind)
	for nextPrec > prec {
		infix, ok := a.infixParsers[a.nxtLex.Kind]
		if !ok {
			return left
		}
		a.advance()
		left = infix(left)
		nextPrec = a.getPrecedence(a.nxtLex.Kind)
	}
	return left
}

func (a *analyzer) parseStr() Node {
	return StringLiteral{Value: a.curLex.Text}
}

func (a *analyzer) parseIdent() Node {
	return Ident{Lexeme: *a.curLex, IsFunc: a.nxtLex.Kind == OPEN_PAREN}
}
func (a *analyzer) parseInteger() Node {
	v, _ := strconv.Atoi(a.curLex.Text)
	return IntegerLiteral{
		Lexeme: *a.curLex,
		Value:  v,
	}
}

func (a *analyzer) parseFloating() Node {
	v, _ := strconv.ParseFloat(a.curLex.Text, 64)
	return FloatLiteral{
		Lexeme: *a.curLex,
		Value:  v,
	}
}

func (a *analyzer) parseBool() Node {
	b, _ := strconv.ParseBool(a.curLex.Text)
	return BooleanLiteral{
		Lexeme: *a.curLex,
		Value:  b,
	}
}

func (a *analyzer) parseRet() Node {
	a.advance()
	return ReturnStmt{
		Expr: a.parseExpr(LOWEST_PREC),
	}
}

func (a *analyzer) parseVarAssign(left Node) Node {
	a.advance()
	assign := VarAssign{
		Name: left,
	}
	assign.Value = a.parseExpr(LOWEST_PREC)
	return assign
}

func (a *analyzer) parseGroup() Node {
	a.advance()
	exp := a.parseExpr(LOWEST_PREC)
	a.advance()
	return exp
}

func (a *analyzer) parsePrefixOperator() Node {
	op := PrefixOp{
		Lexeme: *a.curLex,
	}
	a.advance()
	op.Expr = a.parseExpr(PREC_PREFIX)
	return op
}

func (a *analyzer) parseInfixOperator(left Node) Node {
	op := InfixOp{
		Lexeme: *a.curLex,
		Left:   left,
	}
	prec := a.getPrecedence(a.curLex.Kind)
	a.advance()
	op.Right = a.parseExpr(prec)
	return op
}

func (a *analyzer) parseBlock() *BlockStmt {
	block := &BlockStmt{
		Stmts: []Node{},
	}
	for a.nxtLex.Kind != CLOSE_CURLY {
		a.advance()
		exp := a.parseExpr(LOWEST_PREC)
		block.Stmts = append(block.Stmts, exp)
	}
	a.advance()
	return block
}

func (a *analyzer) parseIf() Node {
	a.advance()
	cond := a.parseExpr(LOWEST_PREC)
	ifStmt := IfStmt{
		Condition: cond,
	}
	a.advance()
	ifStmt.Then = a.parseBlock()
	if !a.checkNext(ELSE_T) {
		return ifStmt
	}
	a.advance()
	ifStmt.Else = a.parseBlock()
	return ifStmt
}

func (a *analyzer) parseFor() Node {
	a.advance()
	keyIdent := a.parseIdent().(Ident)
	forStmt := ForStmt{
		Key: &keyIdent,
	}
	if a.checkNext(COMMA_SYM) {
		a.advance()
		valIdent := a.parseIdent().(Ident)
		forStmt.Value = &valIdent
	}
	a.advance()
	a.advance()
	forStmt.Target = a.parseExpr(LOWEST_PREC)
	a.advance()
	forStmt.Body = a.parseBlock()
	return forStmt
}

func (a *analyzer) parseRange(left Node) Node {
	rnge := RangeExpr{
		From: left,
	}
	a.advance()
	rnge.To = a.parseExpr(LOWEST_PREC)
	if a.checkNext(COLON_SYM) {
		a.advance()
		rnge.Step = a.parseExpr(LOWEST_PREC)
	}
	return rnge
}

func (a *analyzer) parsePrint() Node {
	ps := PrintStmt{
		Args:    []Node{},
		NewLine: a.curLex.Kind == PRINTLN_T,
	}
	a.advance()
	ps.Args = a.parseArgList()
	return ps
}

func (a *analyzer) parseArgList() []Node {
	args := []Node{}
	for a.nxtLex.Kind != CLOSE_PAREN {
		a.advance()
		if a.curLex.Kind == COMMA_SYM {
			a.advance()
		}
		args = append(args, a.parseExpr(LOWEST_PREC))
	}
	a.advance()
	return args
}

func (a *analyzer) parseArray() Node {
	a.advance()
	arr := ArrayLiteral{
		Elements: make([]Node, 0),
	}
	for a.curLex.Kind != CLOSE_BRACKET {
		arr.Elements = append(arr.Elements, a.parseExpr(LOWEST_PREC))
		a.advance()
		if a.curLex.Kind == COMMA_SYM {
			a.advance()
		}
	}
	return arr
}

func (a *analyzer) parseIndex(left Node) Node {
	a.advance()
	idx := IndexExpr{
		Collection: left,
		Index:      a.parseExpr(LOWEST_PREC),
	}
	a.advance()
	return idx
}

func (a *analyzer) parseMap() Node {
	a.advance()
	m := MapLiteral{Pairs: map[Node]Node{}}
	for {
		key := a.parseExpr(LOWEST_PREC)
		a.advance()
		a.advance()
		val := a.parseExpr(LOWEST_PREC)
		m.Pairs[key] = val

		if a.nxtLex.Kind == COMMA_SYM {
			a.advance()
			a.advance()
		}

		if a.nxtLex.Kind == CLOSE_CURLY {
			a.advance()
			break
		}
	}
	return m
}

func (a *analyzer) parseFunction() Node {
	a.advance()
	fn := FunctionLiteral{Name: a.curLex.Text}
	a.advance()
	fn.Params = a.parseParamList()
	fn.Body = a.parseBlock()
	return fn
}

func (a *analyzer) parseParamList() []*Ident {
	params := []*Ident{}
	for a.nxtLex.Kind != CLOSE_PAREN {
		a.advance()
		if a.curLex.Kind == COMMA_SYM {
			a.advance()
		}
		params = append(params, &Ident{Lexeme: *a.curLex})
	}
	a.advance()
	a.advance()
	return params
}

func (a *analyzer) parseCall(function Node) Node {
	return CallExpr{
		Function: function,
		Args:     a.parseArgList(),
	}
}

func (a *analyzer) parseSwap() Node {
	a.advance()
	a.advance()
	swap := SwapStmt{
		A: a.parseExpr(LOWEST_PREC),
	}
	a.advance()
	a.advance()
	swap.B = a.parseExpr(LOWEST_PREC)
	return swap
}

func (a *analyzer) parseImport() Node {
	a.advance()
	a.advance()
	imp := ImportStmt{
		File: a.parseExpr(LOWEST_PREC),
	}
	a.advance()
	return imp
}

func (a *analyzer) parseInput() Node {
	a.advance()
	a.advance()
	inp := InputStmt{
		Prompt: a.parseExpr(LOWEST_PREC),
	}
	a.advance()
	return inp
}

func (a *analyzer) parseLen() Node {
	a.advance()
	a.advance()
	ln := LengthExpr{Target: a.parseExpr(LOWEST_PREC)}
	a.advance()
	return ln
}
