package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Environment struct {
	variables map[string]any
	functions map[string]any
	parent    *Environment
}

func CreateEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables: make(map[string]any),
		functions: make(map[string]any),
		parent:    parent,
	}
}

func (env *Environment) SetVariable(k Node, v any) {
	switch node := k.(type) {
	case Ident:
		env.variables[node.Lexeme.Text] = v
	case IndexExpr:
		name := node.Collection.(Ident).Lexeme.Text
		index := node.Index.Evaluate(env)
		t, _ := env.GetVariable(name)
		switch arrMap := t.(type) {
		case []any:
			arrMap[index.(int)] = v
		case map[any]any:
			arrMap[index] = v
		}
		env.variables[name] = t
	}
}

func (env *Environment) GetVariable(s string) (any, bool) {
	v, ok := env.variables[s]
	if !ok && env.parent != nil {
		return env.parent.GetVariable(s)
	}
	return v, ok
}

func (env *Environment) SetFunction(s string, v any) {
	env.functions[s] = v
}

func (env *Environment) GetFunction(s string) (any, bool) {
	v, ok := env.functions[s]
	if !ok && env.parent != nil {
		return env.parent.GetFunction(s)
	}
	return v, ok
}

func (n StringLiteral) Evaluate(env *Environment) any {
	return n.Value
}

func (n BooleanLiteral) Evaluate(env *Environment) any {
	return n.Value
}

func (n IntegerLiteral) Evaluate(env *Environment) any {
	return n.Value
}

func (n FloatLiteral) Evaluate(env *Environment) any {
	return n.Value
}

func (n ReturnStmt) Evaluate(env *Environment) any {
	return n.Expr.Evaluate(env)
}

func (n Ident) Evaluate(env *Environment) any {
	if n.IsFunc {
		v, _ := env.GetFunction(n.Lexeme.Text)
		return v
	}
	v, _ := env.GetVariable(n.Lexeme.Text)
	return v
}

func (n VarAssign) Evaluate(env *Environment) any {
	v := n.Value.Evaluate(env)
	env.SetVariable(n.Name, v)
	return v
}

func (n PrefixOp) Evaluate(env *Environment) any {
	v := n.Expr.Evaluate(env)
	return evalPrefix(n.Lexeme.Text, v)
}

func evalPrefix(prefix string, v any) any {
	switch v := v.(type) {
	case bool:
		if prefix == "!" {
			return !v
		}
	case int:
		return v * -1
	case float64:
		return v * -1
	}
	return nil
}

func (n InfixOp) Evaluate(env *Environment) any {
	l := n.Left.Evaluate(env)
	r := n.Right.Evaluate(env)
	operator := n.Lexeme.Text

	if l == nil {
		return r
	}
	if r == nil {
		return l
	}

	switch l.(type) {
	case int:
		switch r.(type) {
		case int:
			return evalIntInt(l, r, operator)
		case string:
			return evalIntString(l, r, operator)
		case float64:
			return evalIntFloat(l, r, operator)
		default:
			return nil
		}
	case float64:
		switch r.(type) {
		case int:
			return evalIntFloat(r, l, operator)
		case string:
			return evalFloatString(l, r, operator)
		case float64:
			return evalFloatFloat(l, r, operator)
		default:
			return nil
		}
	case string:
		switch r.(type) {
		case string:
			return evalStringString(l, r, operator)
		case int:
			return evalStringInt(l, r, operator)
		case float64:
			return evalStringFloat(l, r, operator)
		default:
			return nil
		}
	case bool:
		switch r.(type) {
		case bool:
			return evalBoolBool(l, r, operator)
		default:
			return nil
		}
	default:
		return nil
	}
}

func evalIntInt(ll, rr any, operator string) any {
	l := ll.(int)
	r := rr.(int)
	switch operator {
	case ">":
		return l > r
	case ">=":
		return l >= r
	case "<":
		return l < r
	case "<=":
		return l <= r
	case "==":
		return l == r
	case "!=":
		return l != r
	case "+":
		return l + r
	case "-":
		return l - r
	case "*":
		return l * r
	case "/":
		return l / r
	default:
		return nil
	}
}

func evalFloatFloat(ll, rr any, operator string) any {
	l := ll.(float64)
	r := rr.(float64)
	switch operator {
	case ">":
		return l > r
	case ">=":
		return l >= r
	case "<":
		return l < r
	case "<=":
		return l <= r
	case "==":
		return l == r
	case "!=":
		return l != r
	case "+":
		return l + r
	case "-":
		return l - r
	case "*":
		return l * r
	case "/":
		return l / r
	default:
		return nil
	}
}

func evalIntString(ll, rr any, operator string) int {
	l := ll.(int)
	r, _ := strconv.Atoi(rr.(string))
	switch operator {
	case "+":
		return l + r
	case "-":
		return l - r
	case "*":
		return l * r
	case "/":
		return l / r
	default:
		return 0
	}
}

func evalFloatString(ll, rr any, operator string) float64 {
	l := ll.(float64)
	r, _ := strconv.ParseFloat(rr.(string), 64)
	switch operator {
	case "+":
		return l + r
	case "-":
		return l - r
	case "*":
		return l * r
	case "/":
		return l / r
	default:
		return 0
	}
}

func evalIntFloat(ll, rr any, operator string) any {
	l := float64(ll.(int))
	r := rr.(float64)
	switch operator {
	case ">":
		return l > r
	case ">=":
		return l >= r
	case "<":
		return l < r
	case "<=":
		return l <= r
	case "==":
		return l == r
	case "!=":
		return l != r
	case "+":
		return l + r
	case "-":
		return l - r
	case "*":
		return l * r
	case "/":
		return l / r
	default:
		return nil
	}
}

func evalStringString(ll, rr any, operator string) string {
	l := ll.(string)
	r := rr.(string)
	switch operator {
	case "+":
		return l + r
	default:
		return ""
	}
}

func evalStringInt(ll, rr any, operator string) string {
	l := ll.(string)
	r := strconv.Itoa(rr.(int))
	switch operator {
	case "+":
		return l + r
	default:
		return ""
	}
}

func evalStringFloat(ll, rr any, operator string) string {
	l := ll.(string)
	r := strconv.FormatFloat(rr.(float64), 'f', -1, 64)
	switch operator {
	case "+":
		return l + r
	default:
		return ""
	}
}

func evalBoolBool(ll, rr any, operator string) bool {
	l := ll.(bool)
	r := rr.(bool)
	switch operator {
	case "==":
		return l == r
	case "!=":
		return l != r
	case "or":
		return l || r
	case "and":
		return l && r
	}
	return false
}

func (n BlockStmt) Evaluate(env *Environment) any {
	var result any
	for _, stm := range n.Stmts {
		if stm == nil {
			continue
		}
		result = stm.Evaluate(env)
		if result != nil {
			switch stm.(type) {
			case IfStmt, ReturnStmt:
				return result
			}
		}
	}
	return result
}

func (n IfStmt) Evaluate(env *Environment) any {
	condition := n.Condition.Evaluate(env).(bool)
	if condition {
		return n.Then.Evaluate(env)
	} else if n.Else != nil {
		return n.Else.Evaluate(env)
	}
	return nil
}

func (n ArrayLiteral) Evaluate(env *Environment) any {
	ret := []any{}
	for _, node := range n.Elements {
		ret = append(ret, node.Evaluate(env))
	}
	return ret
}

func (n MapLiteral) Evaluate(env *Environment) any {
	m := map[any]any{}
	for k, v := range n.Pairs {
		m[k.Evaluate(env)] = v.Evaluate(env)
	}
	return m
}

func (n IndexExpr) Evaluate(env *Environment) any {
	index := n.Index.Evaluate(env)
	arrMap := n.Collection.Evaluate(env)
	switch arrMap.(type) {
	case map[any]any:
		return arrMap.(map[any]any)[index]
	default:
		return arrMap.([]any)[index.(int)]
	}
}

func (n PrintStmt) Evaluate(env *Environment) any {
	args := evalExpressions(n.Args, env)
	if n.NewLine {
		fmt.Println(args...)
	} else {
		fmt.Print(args...)
	}
	return nil
}

func evalExpressions(exps []Node, env *Environment) []any {
	res := []any{}
	for _, exp := range exps {
		r := exp.Evaluate(env)
		res = append(res, r)
	}
	return res
}

func (n FunctionLiteral) Evaluate(env *Environment) any {
	n.Env = env
	env.SetFunction(n.Name, n)
	return n
}

func (n CallExpr) Evaluate(env *Environment) any {
	fn := n.Function.Evaluate(env).(FunctionLiteral)
	args := evalExpressions(n.Args, env)
	return applyFunction(fn, args, true)
}

func argsToEnvironment(fn FunctionLiteral, args []any, fresh bool) *Environment {
	env := fn.Env
	if fresh {
		env = CreateEnvironment(fn.Env)
	}
	for i, param := range fn.Params {
		env.SetVariable(*param, args[i])
	}
	return env
}

func applyFunction(fn FunctionLiteral, args []any, fresh bool) any {
	newEnv := argsToEnvironment(fn, args, fresh)
	return fn.Body.Evaluate(newEnv)
}

func (n ForStmt) Evaluate(env *Environment) any {
	subject := n.Target.Evaluate(env)
	fn := FunctionLiteral{
		Body:   n.Body,
		Params: []*Ident{n.Key},
		Env:    env,
	}
	if n.Value != nil {
		fn.Params = append(fn.Params, n.Value)
	}
	switch subject.(type) {
	case string:
		for _, v := range subject.(string) {
			args := []any{string(v)}
			applyFunction(fn, args, false)
		}
	case map[any]any:
		for k, v := range subject.(map[any]any) {
			args := []any{k}
			if n.Value != nil {
				args = []any{k, v}
			}
			applyFunction(fn, args, false)
		}
	case []any:
		for k, v := range subject.([]any) {
			args := []any{v}
			if n.Value != nil {
				args = []any{k, v}
			}
			applyFunction(fn, args, false)
		}
	}
	return nil
}

func (n RangeExpr) Evaluate(env *Environment) any {
	from := n.From.Evaluate(env).(int)
	to := n.To.Evaluate(env).(int)
	step := 1
	if n.Step != nil {
		step = n.Step.Evaluate(env).(int)
	}
	ret := []any{}
	if from < to {
		for i := from; i <= to; i += step {
			ret = append(ret, i)
		}
		return ret
	}
	if step > 0 {
		step *= -1
	}
	for i := from; i >= to; i += step {
		ret = append(ret, i)
	}
	return ret
}

func (n SwapStmt) Evaluate(env *Environment) any {
	t := n.A.Evaluate(env)
	env.SetVariable(n.A, n.B.Evaluate(env))
	env.SetVariable(n.B, t)
	return nil
}

func (n ImportStmt) Evaluate(env *Environment) any {
	t := n.File.Evaluate(env).(string)
	input, err := os.ReadFile(t)
	if err != nil {
		log.Fatal(err)
	}
	l := CreateScanner(string(input))
	p := CreateParser(l.lexemes)
	EvaluateNodes(p.astNodes, env)
	return nil
}

func (n InputStmt) Evaluate(env *Environment) any {
	reader := bufio.NewReader(os.Stdin)
	prompt := n.Prompt.Evaluate(env)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return text
}

func (n LengthExpr) Evaluate(env *Environment) any {
	v := n.Target.Evaluate(env)
	switch t := v.(type) {
	case string:
		return len(t)
	case map[any]any:
		return len(t)
	case []any:
		return len(t)
	}
	return nil
}

func EvaluateNodes(nodes chan Node, env *Environment) any {
	var result any
	for node := range nodes {
		result = node.Evaluate(env)
	}
	return result
}
