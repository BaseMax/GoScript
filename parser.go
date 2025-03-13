package main

import (
	"log"
	"strconv"
)

const (
	LowestPriority int = iota + 1
	Equals             // ==
	LessGreater        // > or <
	Sum                // +
	Product            // *
	Prefix             // -X or !X
	Call               // myFunction(X)
	Index              // array[index]
	RangePriority      // 1..10
)

var precedence = map[TokenType]int{
	EQ:       Equals,
	EQEQ:     Equals,
	NEQ:      Equals,
	LESSER:   LessGreater,
	GREATER:  LessGreater,
	LEQ:      LessGreater,
	GEQ:      LessGreater,
	PLUS:     Sum,
	MINUS:    Sum,
	SLASH:    Product,
	STAR:     Product,
	LPARENT:  Call,
	LBRACKET: Index,
	DOTDOT:   RangePriority,
}

type Node interface {
	Eval(scope *Scope) any
}

type Parser struct {
	tokens          chan Token
	currentToken    Token
	peekToken       Token
	nodes           chan Node
	prefixParselets map[TokenType]func() Node
	infixParselets  map[TokenType]func(Node) Node
}

func NewParser(tokens chan Token) *Parser {
	p := &Parser{
		tokens:          tokens,
		nodes:           make(chan Node),
		prefixParselets: make(map[TokenType]func() Node),
		infixParselets:  make(map[TokenType]func(Node) Node),
	}
	p.registerParselets()
	p.currentToken = <-tokens
	p.peekToken = <-tokens
	go p.parse()
	return p
}

func (p *Parser) registerParselets() {
	p.prefixParselets = map[TokenType]func() Node{
		IDENT:    p.parseIdentifier,
		STRING:   p.parseString,
		INT:      p.parseInt,
		FLOAT:    p.parseFloat,
		MINUS:    p.parseUnaryOp,
		NOT:      p.parseUnaryOp,
		TRUE:     p.parseBool,
		FALSE:    p.parseBool,
		LPARENT:  p.parseGrouped,
		IF:       p.parseIf,
		FN:       p.parseFunction,
		PRINT:    p.parsePrint,
		PRINTLN:  p.parsePrint,
		LBRACKET: p.parseArray,
		LCURLY:   p.parseMap,
		FOR:      p.parseFor,
		RETURN:   p.parseReturn,
		SWAP:     p.parseSwap,
		INPUT:    p.parseInput,
		LEN:      p.parseLen,
		IMPORT:   p.parseImport,
	}
	for _, t := range []TokenType{OR, AND, PLUS, MINUS, STAR, SLASH, EQEQ, NEQ, GREATER, GEQ, LESSER, LEQ} {
		p.infixParselets[t] = p.parseBinaryOp
	}
	p.infixParselets[LPARENT] = p.parseFunctionCall
	p.infixParselets[LBRACKET] = p.parseArrayIndex
	p.infixParselets[DOTDOT] = p.parseRange
	p.infixParselets[EQ] = p.parseVariable
}

func (p *Parser) parse() {
	defer close(p.nodes)
	for p.currentToken.Type != EOF {
		if node := p.parseExpression(LowestPriority); node != nil {
			p.nodes <- node
		}
		p.advance()
	}
}

func (p *Parser) advance() {
	p.currentToken = p.peekToken
	p.peekToken = <-p.tokens
}

func (p *Parser) expect(t TokenType) {
	if p.currentToken.Type != t {
		log.Fatalf("Expected %s, got %s", t, p.currentToken.Type)
	}
	p.advance()
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedence[p.peekToken.Type]; ok {
		return p
	}
	return LowestPriority
}

func (p *Parser) parseExpression(precedence int) Node {
	prefix, ok := p.prefixParselets[p.currentToken.Type]
	if !ok {
		return nil
	}
	left := prefix()
	for p.peekPrecedence() > precedence {
		infix, ok := p.infixParselets[p.peekToken.Type]
		if !ok {
			break
		}
		p.advance()
		left = infix(left)
	}
	return left
}

type StringNode struct{ Value string }

func (p *Parser) parseString() Node {
	s := StringNode{Value: p.currentToken.Value}
	p.advance()
	return &s
}

type IdentifierNode struct{ Name string }

func (p *Parser) parseIdentifier() Node {
	id := IdentifierNode{Name: p.currentToken.Value}
	p.advance()
	return &id
}

type IntNode struct{ Value int }

func (p *Parser) parseInt() Node {
	v, err := strconv.Atoi(p.currentToken.Value)
	if err != nil {
		log.Fatalf("Invalid integer: %s", p.currentToken.Value)
	}
	n := IntNode{Value: v}
	p.advance()
	return &n
}

type FloatNode struct{ Value float64 }

func (p *Parser) parseFloat() Node {
	v, err := strconv.ParseFloat(p.currentToken.Value, 64)
	if err != nil {
		log.Fatalf("Invalid float: %s", p.currentToken.Value)
	}
	n := FloatNode{Value: v}
	p.advance()
	return &n
}

type BoolNode struct{ Value bool }

func (p *Parser) parseBool() Node {
	v, err := strconv.ParseBool(p.currentToken.Value)
	if err != nil {
		log.Fatalf("Invalid boolean: %s", p.currentToken.Value)
	}
	n := BoolNode{Value: v}
	p.advance()
	return &n
}

type ReturnNode struct{ Value Node }

func (p *Parser) parseReturn() Node {
	p.advance()
	return &ReturnNode{Value: p.parseExpression(LowestPriority)}
}

type VariableNode struct {
	Name  Node
	Value Node
}

func (p *Parser) parseVariable(left Node) Node {
	p.advance()
	return &VariableNode{Name: left, Value: p.parseExpression(LowestPriority)}
}

type UnaryOpNode struct {
	Op    string
	Right Node
}

func (p *Parser) parseUnaryOp() Node {
	op := p.currentToken.Value
	p.advance()
	return &UnaryOpNode{Op: op, Right: p.parseExpression(Prefix)}
}

type BinaryOpNode struct {
	Op    string
	Left  Node
	Right Node
}

func (p *Parser) parseBinaryOp(left Node) Node {
	op := p.currentToken.Value
	priority := precedence[p.currentToken.Type]
	p.advance()
	return &BinaryOpNode{Op: op, Left: left, Right: p.parseExpression(priority)}
}

type BlockNode struct{ Statements []Node }

func (p *Parser) parseBlock() *BlockNode {
	p.expect(LCURLY)
	block := &BlockNode{}
	for p.currentToken.Type != RCURLY && p.currentToken.Type != EOF {
		if stmt := p.parseExpression(LowestPriority); stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.advance()
	}
	p.expect(RCURLY)
	return block
}

type IfNode struct {
	Condition Node
	True      *BlockNode
	Else      *BlockNode
}

func (p *Parser) parseIf() Node {
	p.advance()
	condition := p.parseExpression(LowestPriority)
	p.advance()
	trueBlock := p.parseBlock()
	var elseBlock *BlockNode
	if p.currentToken.Type == ELSE {
		p.advance()
		elseBlock = p.parseBlock()
	}
	return &IfNode{Condition: condition, True: trueBlock, Else: elseBlock}
}

type ForNode struct {
	Key     *IdentifierNode
	Value   *IdentifierNode
	Subject Node
	Body    *BlockNode
}

func (p *Parser) parseFor() Node {
	p.advance()
	key := p.parseIdentifier().(*IdentifierNode)
	var value *IdentifierNode
	if p.currentToken.Type == COMMA {
		p.advance()
		value = p.parseIdentifier().(*IdentifierNode)
	}
	p.expect(FOR)
	p.advance()
	subject := p.parseExpression(LowestPriority)
	p.advance()
	body := p.parseBlock()
	return &ForNode{Key: key, Value: value, Subject: subject, Body: body}
}

type RangeNode struct {
	From Node
	To   Node
	Step Node
}

func (p *Parser) parseRange(left Node) Node {
	p.advance()
	to := p.parseExpression(LowestPriority)
	var step Node
	if p.currentToken.Type == COLON {
		p.advance()
		step = p.parseExpression(LowestPriority)
	}
	return &RangeNode{From: left, To: to, Step: step}
}

type PrintNode struct {
	Args    []Node
	Newline bool
}

func (p *Parser) parsePrint() Node {
	newline := p.currentToken.Type == PRINTLN
	p.advance()
	p.expect(LPARENT)
	args := p.parseArgs(RPARENT)
	return &PrintNode{Args: args, Newline: newline}
}

type ArrayNode struct{ Elements []Node }

func (p *Parser) parseArray() Node {
	p.advance()
	elements := p.parseArgs(RBRACKET)
	return &ArrayNode{Elements: elements}
}

type MapNode struct{ Pairs map[Node]Node }

func (p *Parser) parseMap() Node {
	p.advance()
	pairs := make(map[Node]Node)
	for p.currentToken.Type != RCURLY {
		key := p.parseExpression(LowestPriority)
		p.expect(COLON)
		p.advance()
		value := p.parseExpression(LowestPriority)
		pairs[key] = value
		if p.currentToken.Type == COMMA {
			p.advance()
		}
	}
	p.advance()
	return &MapNode{Pairs: pairs}
}

type IndexNode struct {
	Subject Node
	Index   Node
}

func (p *Parser) parseArrayIndex(left Node) Node {
	p.advance()
	index := p.parseExpression(LowestPriority)
	p.expect(RBRACKET)
	return &IndexNode{Subject: left, Index: index}
}

type FunctionNode struct {
	Name   string
	Params []*IdentifierNode
	Body   *BlockNode
	Scope  *Scope
}

func (p *Parser) parseFunction() Node {
	p.advance()
	name := p.currentToken.Value
	p.advance()
	p.expect(LPARENT)
	params := p.parseParams()
	p.advance()
	body := p.parseBlock()
	return &FunctionNode{Name: name, Params: params, Body: body}
}

type CallNode struct {
	Function Node
	Args     []Node
}

func (p *Parser) parseFunctionCall(left Node) Node {
	p.advance()
	args := p.parseArgs(RPARENT)
	return &CallNode{Function: left, Args: args}
}

type SwapNode struct {
	Left  Node
	Right Node
}

func (p *Parser) parseSwap() Node {
	p.advance()
	p.expect(LPARENT)
	left := p.parseExpression(LowestPriority)
	p.expect(COMMA)
	p.advance()
	right := p.parseExpression(LowestPriority)
	p.expect(RPARENT)
	return &SwapNode{Left: left, Right: right}
}

type ImportNode struct{ Filename Node }

func (p *Parser) parseImport() Node {
	p.advance()
	p.expect(LPARENT)
	filename := p.parseExpression(LowestPriority)
	p.expect(RPARENT)
	return &ImportNode{Filename: filename}
}

type InputNode struct{ Prompt Node }

func (p *Parser) parseInput() Node {
	p.advance()
	p.expect(LPARENT)
	prompt := p.parseExpression(LowestPriority)
	p.expect(RPARENT)
	return &InputNode{Prompt: prompt}
}

type LenNode struct{ Subject Node }

func (p *Parser) parseLen() Node {
	p.advance()
	p.expect(LPARENT)
	subject := p.parseExpression(LowestPriority)
	p.expect(RPARENT)
	return &LenNode{Subject: subject}
}

func (p *Parser) parseArgs(end TokenType) []Node {
	var args []Node
	for p.currentToken.Type != end {
		if len(args) > 0 {
			p.expect(COMMA)
		}
		p.advance()
		args = append(args, p.parseExpression(LowestPriority))
	}
	p.advance()
	return args
}

func (p *Parser) parseParams() []*IdentifierNode {
	var params []*IdentifierNode
	for p.currentToken.Type != RPARENT {
		if len(params) > 0 {
			p.expect(COMMA)
		}
		p.advance()
		if p.currentToken.Type != IDENT {
			log.Fatalf("Expected identifier, got %s", p.currentToken.Type)
		}
		params = append(params, &IdentifierNode{Name: p.currentToken.Value})
	}
	return params
}

func (p *Parser) parseGrouped() Node {
	p.advance()
	expr := p.parseExpression(LowestPriority)
	p.expect(RPARENT)
	return expr
}
