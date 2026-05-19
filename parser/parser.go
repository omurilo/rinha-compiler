package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/omurilo/rinha-compiler/ast"
	"github.com/omurilo/rinha-compiler/lexer"
	"github.com/omurilo/rinha-compiler/runtime"
)

var BINARY_OPERATIONS = map[string]string{
	"+":  "Add",
	"-":  "Sub",
	"*":  "Mul",
	"/":  "Div",
	"%":  "Rem",
	"==": "Eq",
	"<":  "Lt",
	">":  "Gt",
	">=": "Gte",
	"<=": "Lte",
	"!=": "Neq",
	"&&": "And",
	"||": "Or",
}

var PRECEDENCE = map[string]int{
	"||": 1,
	"&&": 2,
	"==": 3, "!=": 3,
	"<": 4, ">": 4, "<=": 4, ">=": 4,
	"+": 5, "-": 5,
	"*": 6, "/": 6, "%": 6,
}

var tree map[string]interface{}
var current_token lexer.Token
var p *lexer.CustomParser
var last_end uint32 // byte offset after the last consumed token

func Main(program string, filename string) string {
	p = lexer.Initialize(program, filename)

	advance()
	expr := parser()
	tree = map[string]interface{}{
		"name":       filename,
		"expression": expr,
		"location":   ast.Location{Start: 0, End: uint32(len(program)), Filename: filename},
	}
	treeJson, err := json.Marshal(tree)

	if err != nil {
		runtime.Error(current_token.Location, "unexpected error occurred at parsing tree to json")
	}

	return string(treeJson)
}

func parser() map[string]interface{} {
	return parse_binary(parse_atom(), 0)
}

func parse_atom() map[string]interface{} {
	switch current_token.Type {
	case ":PRINT":
		return parse_print()
	case ":STRING":
		return parse_string()
	case ":NUMBER":
		return parse_number()
	case ":IDENTIFIER":
		return parse_identifier()
	case ":LET":
		return parse_let()
	case ":FUNCTION":
		return parse_function()
	case ":IF":
		return parse_if()
	case ":TRUE":
		return parse_bool()
	case ":FALSE":
		return parse_bool()
	case ":FIRST":
		return parse_first()
	case ":SECOND":
		return parse_second()
	case ":LPAREN":
		start := current_token.Location.Start
		filename := current_token.Location.Filename
		expr := parse_paren_or_tuple()
		// Handle IIFE: (fn ...)(args)
		if current_token.Type == ":LPAREN" {
			consume(":LPAREN")
			call := parse_function_call(expr)
			consume(":RPAREN")
			call["location"] = ast.Location{Start: start, End: last_end, Filename: filename}
			return call
		}
		return expr
	default:
		return nil
	}
}

// parse_binary implements precedence climbing.
// "+" is right-associative so that "str" + int + int evaluates as "str" + (int + int).
// All other operators are left-associative.
func parse_binary(lhs map[string]interface{}, minPrec int) map[string]interface{} {
	for current_token.Type == ":BINARY_OP" {
		prec, ok := PRECEDENCE[current_token.Value]
		if !ok || prec < minPrec {
			break
		}
		op := current_token.Value
		opLoc := current_token.Location
		consume(":BINARY_OP")
		nextPrec := prec + 1 // left-associative by default
		if op == "+" {
			nextPrec = prec // right-associative for +
		}
		rhs := parse_binary(parse_atom(), nextPrec)
		lhsStart := nodStart(lhs)
		rhsEnd := nodeEnd(rhs)
		lhs = map[string]interface{}{
			"kind": "Binary", "op": BINARY_OPERATIONS[op],
			"lhs": lhs, "rhs": rhs,
			"location": ast.Location{Start: lhsStart, End: rhsEnd, Filename: opLoc.Filename},
		}
	}
	return lhs
}

func parse_print() map[string]interface{} {
	start := current_token.Location.Start
	filename := current_token.Location.Filename
	consume(":PRINT")
	consume(":LPAREN")
	value := parser()
	consume(":RPAREN")
	return map[string]interface{}{
		"kind":     "Print",
		"value":    value,
		"location": ast.Location{Start: start, End: last_end, Filename: filename},
	}
}

func parse_string() map[string]interface{} {
	node := map[string]interface{}{"kind": "Str", "value": current_token.Value, "location": current_token.Location}
	consume(":STRING")
	return node
}

func parse_number() map[string]interface{} {
	number, _ := strconv.Atoi(current_token.Value)
	node := map[string]interface{}{"kind": "Int", "value": number, "location": current_token.Location}
	consume(":NUMBER")
	return node
}

func parse_let() map[string]interface{} {
	start := current_token.Location.Start
	filename := current_token.Location.Filename
	consume(":LET")

	nameText := current_token.Value
	nameLoc := current_token.Location
	consume(":IDENTIFIER")
	consume(":ASSIGNMENT")

	value := parser()
	consume(":SEMICOLON")
	next := parser()

	return map[string]interface{}{
		"kind":  "Let",
		"name":  map[string]interface{}{"text": nameText, "location": nameLoc},
		"value": value,
		"next":  next,
		"location": ast.Location{Start: start, End: last_end, Filename: filename},
	}
}

func parse_function() map[string]interface{} {
	start := current_token.Location.Start
	filename := current_token.Location.Filename
	consume(":FUNCTION")
	consume(":LPAREN")

	var parameters []map[string]interface{}
	for current_token.Type != ":RPAREN" {
		parameter := map[string]interface{}{"text": current_token.Value, "location": current_token.Location}
		parameters = append(parameters, parameter)
		consume(":IDENTIFIER")
		if current_token.Type == ":COMMA" {
			consume(":COMMA")
		}
	}

	consume(":RPAREN")
	consume(":ARROW")
	consume(":LBRACE")
	value := parser()
	consume(":RBRACE")

	return map[string]interface{}{
		"kind":       "Function",
		"parameters": parameters,
		"value":      value,
		"location":   ast.Location{Start: start, End: last_end, Filename: filename},
	}
}

func parse_if() map[string]interface{} {
	start := current_token.Location.Start
	filename := current_token.Location.Filename
	consume(":IF")
	consume(":LPAREN")
	condition := parser()
	consume(":RPAREN")
	consume(":LBRACE")
	then := parser()
	consume(":RBRACE")

	node := map[string]interface{}{
		"kind":      "If",
		"condition": condition,
		"then":      then,
	}

	if current_token.Type == ":ELSE" {
		consume(":ELSE")
		consume(":LBRACE")
		node["otherwise"] = parser()
		consume(":RBRACE")
	}

	node["location"] = ast.Location{Start: start, End: last_end, Filename: filename}
	return node
}

// parse_paren_or_tuple distinguishes (expr, expr) tuples from (expr) grouping parens.
func parse_paren_or_tuple() map[string]interface{} {
	start := current_token.Location.Start
	filename := current_token.Location.Filename
	consume(":LPAREN")
	first := parser()

	if current_token.Type == ":COMMA" {
		consume(":COMMA")
		second := parser()
		consume(":RPAREN")
		return map[string]interface{}{
			"kind":     "Tuple",
			"first":    first,
			"second":   second,
			"location": ast.Location{Start: start, End: last_end, Filename: filename},
		}
	}

	consume(":RPAREN")
	return first
}

func parse_first() map[string]interface{} {
	start := current_token.Location.Start
	filename := current_token.Location.Filename
	consume(":FIRST")
	consume(":LPAREN")
	value := parser()
	consume(":RPAREN")
	return map[string]interface{}{
		"kind":     "First",
		"value":    value,
		"location": ast.Location{Start: start, End: last_end, Filename: filename},
	}
}

func parse_second() map[string]interface{} {
	start := current_token.Location.Start
	filename := current_token.Location.Filename
	consume(":SECOND")
	consume(":LPAREN")
	value := parser()
	consume(":RPAREN")
	return map[string]interface{}{
		"kind":     "Second",
		"value":    value,
		"location": ast.Location{Start: start, End: last_end, Filename: filename},
	}
}

func parse_bool() map[string]interface{} {
	node := map[string]interface{}{"kind": "Bool", "value": current_token.Value == "true", "location": current_token.Location}
	consume(fmt.Sprintf(":%s", strings.ToUpper(current_token.Value)))
	return node
}

func parse_identifier() map[string]interface{} {
	start := current_token.Location.Start
	filename := current_token.Location.Filename
	node := map[string]interface{}{"kind": "Var", "text": current_token.Value, "location": current_token.Location}
	consume(":IDENTIFIER")

	if current_token.Type == ":LPAREN" {
		consume(":LPAREN")
		call := parse_function_call(node)
		consume(":RPAREN")
		call["location"] = ast.Location{Start: start, End: last_end, Filename: filename}
		return call
	}

	return node
}

func parse_function_call(callee map[string]interface{}) map[string]interface{} {
	node := map[string]interface{}{"kind": "Call", "callee": callee}
	var arguments []map[string]interface{}

	for current_token.Type != ":RPAREN" {
		argument := parser()
		arguments = append(arguments, argument)
		if current_token.Type == ":COMMA" {
			consume(":COMMA")
		}
	}

	node["arguments"] = arguments
	return node
}

func advance() {
	current_token = p.Next()
}

func consume(token_type string) {
	if current_token.Value == "" {
		runtime.Error(current_token.Location, fmt.Sprintf("Expected %v but found nil", token_type))
	}

	if current_token.Type != token_type {
		runtime.Error(current_token.Location, fmt.Sprintf("Expected %v but found %v in %v", token_type, current_token.Type, current_token.Value))
	}
	last_end = current_token.Location.End
	advance()
}

// nodStart extracts the Start byte offset from a parsed node's location.
func nodStart(node map[string]interface{}) uint32 {
	if node == nil {
		return 0
	}
	if loc, ok := node["location"].(ast.Location); ok {
		return loc.Start
	}
	return 0
}

// nodeEnd extracts the End byte offset from a parsed node's location.
func nodeEnd(node map[string]interface{}) uint32 {
	if node == nil {
		return 0
	}
	if loc, ok := node["location"].(ast.Location); ok {
		return loc.End
	}
	return 0
}
