package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Scope holds variables and functions. (Using pointer receivers as in your second code)
type Scope struct {
	variables map[string]any
	functions map[string]any
	parent    *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		variables: make(map[string]any),
		functions: make(map[string]any),
		parent:    parent,
	}
}

func (s *Scope) SetVariable(name string, value any) {
	s.variables[name] = value
}

func (s *Scope) GetVariable(name string) (any, bool) {
	if v, ok := s.variables[name]; ok {
		return v, true
	}
	if s.parent != nil {
		return s.parent.GetVariable(name)
	}
	return nil, false
}

func (s *Scope) SetFunction(name string, fn any) {
	s.functions[name] = fn
}

func (s *Scope) GetFunction(name string) (any, bool) {
	if f, ok := s.functions[name]; ok {
		return f, true
	}
	if s.parent != nil {
		return s.parent.GetFunction(name)
	}
	return nil, false
}

// ---------- Node Evaluations ----------

// StringNode
func (n *StringNode) Eval(s *Scope) any {
	return n.Value
}

// IntNode
func (n *IntNode) Eval(s *Scope) any {
	return n.Value
}

// FloatNode
func (n *FloatNode) Eval(s *Scope) any {
	return n.Value
}

// BoolNode
func (n *BoolNode) Eval(s *Scope) any {
	return n.Value
}

// ReturnNode
func (n *ReturnNode) Eval(s *Scope) any {
	return n.Value.Eval(s)
}

// IdentifierNode: First try to get a function, then a variable.
func (n *IdentifierNode) Eval(s *Scope) any {
	if f, ok := s.GetFunction(n.Name); ok {
		return f
	}
	if v, ok := s.GetVariable(n.Name); ok {
		return v
	}
	log.Fatalf("Undefined identifier: %s", n.Name)
	return nil
}

// VariableNode: if the name is an IdentifierNode, assign directly; if it’s an IndexNode, update the underlying collection.
func (n *VariableNode) Eval(s *Scope) any {
	value := n.Value.Eval(s)
	switch name := n.Name.(type) {
	case *IdentifierNode:
		s.SetVariable(name.Name, value)
	case *IndexNode:
		subject := name.Subject.Eval(s)
		index := name.Index.Eval(s)
		switch subj := subject.(type) {
		case []any:
			subj[index.(int)] = value
		case map[any]any:
			subj[index] = value
		default:
			log.Fatalf("Cannot index into type %T", subject)
		}
	default:
		log.Fatalf("Unsupported variable target type %T", name)
	}
	return value
}

// UnaryOpNode
func (n *UnaryOpNode) Eval(s *Scope) any {
	v := n.Right.Eval(s)
	switch n.Op {
	case "-":
		switch v := v.(type) {
		case int:
			return -v
		case float64:
			return -v
		}
	case "!":
		if b, ok := v.(bool); ok {
			return !b
		}
	}
	log.Fatalf("Invalid unary operation: %s on %T", n.Op, v)
	return nil
}

// BinaryOpNode
func (n *BinaryOpNode) Eval(s *Scope) any {
	l := n.Left.Eval(s)
	r := n.Right.Eval(s)
	switch lv := l.(type) {
	case int:
		return evalIntBinary(lv, r, n.Op)
	case float64:
		return evalFloatBinary(lv, r, n.Op)
	case string:
		return evalStringBinary(lv, r, n.Op)
	case bool:
		return evalBoolBinary(lv, r, n.Op)
	}
	log.Fatalf("Unsupported types for operator %s: %T and %T", n.Op, l, r)
	return nil
}

func evalIntBinary(l int, r any, op string) any {
	switch rv := r.(type) {
	case int:
		switch op {
		case "+":
			return l + rv
		case "-":
			return l - rv
		case "*":
			return l * rv
		case "/":
			return l / rv
		case ">":
			return l > rv
		case ">=":
			return l >= rv
		case "<":
			return l < rv
		case "<=":
			return l <= rv
		case "==":
			return l == rv
		case "!=":
			return l != rv
		}
	case float64:
		lv := float64(l)
		switch op {
		case "+":
			return lv + rv
		case "-":
			return lv - rv
		case "*":
			return lv * rv
		case "/":
			return lv / rv
		case ">":
			return lv > rv
		case ">=":
			return lv >= rv
		case "<":
			return lv < rv
		case "<=":
			return lv <= rv
		case "==":
			return lv == rv
		case "!=":
			return lv != rv
		}
	}
	log.Fatalf("Invalid operation %s between int and %T", op, r)
	return nil
}

func evalFloatBinary(l float64, r any, op string) any {
	switch rv := r.(type) {
	case int:
		rvFloat := float64(rv)
		switch op {
		case "+":
			return l + rvFloat
		case "-":
			return l - rvFloat
		case "*":
			return l * rvFloat
		case "/":
			return l / rvFloat
		case ">":
			return l > rvFloat
		case ">=":
			return l >= rvFloat
		case "<":
			return l < rvFloat
		case "<=":
			return l <= rvFloat
		case "==":
			return l == rvFloat
		case "!=":
			return l != rvFloat
		}
	case float64:
		switch op {
		case "+":
			return l + rv
		case "-":
			return l - rv
		case "*":
			return l * rv
		case "/":
			return l / rv
		case ">":
			return l > rv
		case ">=":
			return l >= rv
		case "<":
			return l < rv
		case "<=":
			return l <= rv
		case "==":
			return l == rv
		case "!=":
			return l != rv
		}
	}
	log.Fatalf("Invalid operation %s between float and %T", op, r)
	return nil
}

func evalStringBinary(l string, r any, op string) any {
	if op != "+" {
		log.Fatalf("Invalid operation %s on string", op)
	}
	switch rv := r.(type) {
	case string:
		return l + rv
	case int:
		return l + strconv.Itoa(rv)
	case float64:
		return l + strconv.FormatFloat(rv, 'f', -1, 64)
	}
	log.Fatalf("Cannot concatenate string with %T", r)
	return nil
}

func evalBoolBinary(l bool, r any, op string) any {
	if rv, ok := r.(bool); ok {
		switch op {
		case "==":
			return l == rv
		case "!=":
			return l != rv
		case "or":
			return l || rv
		case "and":
			return l && rv
		}
	}
	log.Fatalf("Invalid operation %s between bool and %T", op, r)
	return nil
}

// BlockNode: Evaluate each statement and return early on *IfNode or *ReturnNode (as in your original logic)
func (n *BlockNode) Eval(s *Scope) any {
	var result any
	for _, stmt := range n.Statements {
		result = stmt.Eval(s)
		switch stmt.(type) {
		case *IfNode, *ReturnNode:
			return result
		}
	}
	return result
}

// IfNode
func (n *IfNode) Eval(s *Scope) any {
	if n.Condition.Eval(s).(bool) {
		return n.True.Eval(s)
	} else if n.Else != nil {
		return n.Else.Eval(s)
	}
	return nil
}

// ForNode: creates a function node for each iteration and applies it.
func (n *ForNode) Eval(s *Scope) any {
	subject := n.Subject.Eval(s)
	fn := &FunctionNode{
		Params: []*IdentifierNode{n.Key},
		Body:   n.Body,
		Scope:  s,
	}
	if n.Value != nil {
		fn.Params = append(fn.Params, n.Value)
	}
	switch subj := subject.(type) {
	case map[any]any:
		for k, v := range subj {
			args := []any{k}
			if n.Value != nil {
				args = append(args, v)
			}
			applyFunction(fn, args, false)
		}
	case string:
		for _, c := range subj {
			applyFunction(fn, []any{string(c)}, false)
		}
	case []any:
		for i, v := range subj {
			args := []any{v}
			if n.Value != nil {
				args = []any{i, v}
			}
			applyFunction(fn, args, false)
		}
	}
	return nil
}

// RangeNode
func (n *RangeNode) Eval(s *Scope) any {
	from := n.From.Eval(s).(int)
	to := n.To.Eval(s).(int)
	step := 1
	if n.Step != nil {
		step = n.Step.Eval(s).(int)
	}
	var result []any
	if from <= to {
		for i := from; i <= to; i += step {
			result = append(result, i)
		}
	} else {
		if step > 0 {
			step = -step
		}
		for i := from; i >= to; i += step {
			result = append(result, i)
		}
	}
	return result
}

// PrintNode
func (n *PrintNode) Eval(s *Scope) any {
	args := evalArgs(n.Args, s)
	if n.Newline {
		fmt.Println(args...)
	} else {
		fmt.Print(args...)
	}
	return nil
}

// ArrayNode: evaluate each element
func (n *ArrayNode) Eval(s *Scope) any {
	return evalArgs(n.Elements, s)
}

// MapNode: evaluate each key-value pair
func (n *MapNode) Eval(s *Scope) any {
	m := make(map[any]any)
	for k, v := range n.Pairs {
		m[k.Eval(s)] = v.Eval(s)
	}
	return m
}

// IndexNode: access element of an array or map
func (n *IndexNode) Eval(s *Scope) any {
	subject := n.Subject.Eval(s)
	index := n.Index.Eval(s)
	switch subj := subject.(type) {
	case []any:
		return subj[index.(int)]
	case map[any]any:
		return subj[index]
	}
	log.Fatalf("Cannot index into type %T", subject)
	return nil
}

// FunctionNode: store the current scope and register the function in the scope.
func (n *FunctionNode) Eval(s *Scope) any {
	n.Scope = s
	s.SetFunction(n.Name, n)
	return n
}

// CallNode: call a function node with evaluated arguments.
func (n *CallNode) Eval(s *Scope) any {
	fn, ok := n.Function.Eval(s).(*FunctionNode)
	if !ok {
		log.Fatalf("Attempted to call a non-function")
	}
	args := evalArgs(n.Args, s)
	return applyFunction(fn, args, true)
}

// SwapNode: swap the values of the left and right targets.
// Supports both IdentifierNode and IndexNode.
func (n *SwapNode) Eval(s *Scope) any {
	leftVal := n.Left.Eval(s)
	rightVal := n.Right.Eval(s)

	setValue := func(node any, value any) {
		switch target := node.(type) {
		case *IdentifierNode:
			s.SetVariable(target.Name, value)
		case *IndexNode:
			subject := target.Subject.Eval(s)
			index := target.Index.Eval(s)
			switch subj := subject.(type) {
			case []any:
				subj[index.(int)] = value
			case map[any]any:
				subj[index] = value
			default:
				log.Fatalf("Cannot index into type %T", subject)
			}
		default:
			log.Fatalf("Unsupported swap target type %T", node)
		}
	}
	setValue(n.Left, rightVal)
	setValue(n.Right, leftVal)
	return nil
}

// ImportNode: read, lex, parse, and evaluate a file.
func (n *ImportNode) Eval(s *Scope) any {
	filename := n.Filename.Eval(s).(string)
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file %s: %v", filename, err)
	}
	lexer := NewLexer(string(data))
	parser := NewParser(lexer.tokens)
	Evaluate(parser.nodes, s)
	return nil
}

// InputNode: prompt the user and return input.
// Now the prompt is evaluated with the current scope.
func (n *InputNode) Eval(s *Scope) any {
	prompt := n.Prompt.Eval(s)
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

// LenNode: return the length of a string, array, or map.
func (n *LenNode) Eval(s *Scope) any {
	subject := n.Subject.Eval(s)
	switch subj := subject.(type) {
	case string:
		return len(subj)
	case []any:
		return len(subj)
	case map[any]any:
		return len(subj)
	}
	log.Fatalf("Len not applicable to %T", subject)
	return nil
}
