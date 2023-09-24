package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

var tree map[string]interface{}
var current_token lexer.Token
var p *lexer.CustomParser

func Main(program string, filename string) string {
	p = lexer.Initialize(program, filename)

	advance()
	tree = map[string]interface{}{"name": filename, "expression": parser()}
	treeJson, _ := json.Marshal(tree)

	// err != nil {
	//   runtime.Error(current_token.Location, "unexpected error occurred at parsing tree to json")
	// }

	// fmt.Println(string(treeJson))
	return string(treeJson)
}

func parser() map[string]interface{} {
	switch current_token.Type {
	case ":PRINT":
		return maybe_binary_op(parse_print())
	case ":STRING":
		return maybe_binary_op(parse_string())
	case ":NUMBER":
		return maybe_binary_op(parse_number())
	case ":IDENTIFIER":
		return maybe_binary_op(parse_identifier())
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
	default:
		// consume(current_token.Type)
		return nil
	}
}

func parse_print() map[string]interface{} {
	node := map[string]interface{}{"kind": "Print"}

	consume(":PRINT")
	consume(":LPAREN")
	node["value"] = parser()

	consume(":RPAREN")

	return node
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
	node := map[string]interface{}{"kind": "Let", "location": current_token.Location}
	node["name"] = map[string]interface{}{"text": nil}
	node["value"] = make(map[string]interface{})
	node["next"] = make(map[string]interface{})
	consume(":LET")

	node["name"].(map[string]interface{})["text"] = current_token.Value
	consume(":IDENTIFIER")
	consume(":ASSIGNMENT")

	node["value"] = parser()
	consume(":SEMICOLON")
	node["next"] = parser()

	return node
}

func parse_function() map[string]interface{} {
	node := map[string]interface{}{"kind": "Function", "location": current_token.Location}
	node["parameters"] = []map[string]interface{}{}
	node["value"] = make(map[string]interface{})

	consume(":FUNCTION")
	consume(":LPAREN")

	for current_token.Type != ":RPAREN" {
		parameter := map[string]interface{}{"text": current_token.Value, "location": current_token.Location}
		node["parameters"] = append(node["parameters"].([]map[string]interface{}), parameter)
		consume(":IDENTIFIER")
		if current_token.Type == ":COMMA" {
			consume(":COMMA")
		}
	}

	consume(":RPAREN")
	consume(":ARROW")
	consume(":LBRACE")

	node["value"] = parser()

	consume(":RBRACE")

	return node
}

func parse_if() map[string]interface{} {
	node := map[string]interface{}{"kind": "If", "location": current_token.Location}
	node["then"] = map[string]interface{}{}
	node["otherwise"] = map[string]interface{}{}
	node["condition"] = map[string]interface{}{}

	consume(":IF")
	consume(":LPAREN")

	node["condition"] = parser()

	consume(":RPAREN")
	consume(":LBRACE")

	node["then"] = parser()

	consume(":RBRACE")

	if current_token.Type == ":ELSE" {
		consume(":ELSE")
		consume(":LBRACE")

		node["otherwise"] = parser()

		consume(":RBRACE")
	}

	return node
}

func parse_bool() map[string]interface{} {
	node := map[string]interface{}{"kind": "Bool", "value": current_token.Value == "true", "location": current_token.Location}
	consume(fmt.Sprintf(":%s", strings.ToUpper(current_token.Value)))

	return node
}

func parse_identifier() map[string]interface{} {
	node := map[string]interface{}{"kind": "Var", "text": current_token.Value, "location": current_token.Location}
	consume(":IDENTIFIER")

	if current_token.Type == ":LPAREN" {
		consume(":LPAREN")
		function_call := maybe_binary_op(parse_function_call(node))
		consume(":RPAREN")

		return function_call
	}

	return node
}

func parse_function_call(callee map[string]interface{}) map[string]interface{} {
	node := map[string]interface{}{"kind": "Call", "callee": callee, "location": current_token.Location}
	node["arguments"] = []map[string]interface{}{}

	for current_token.Type != ":RPAREN" {
		argument := parser()
		node["arguments"] = append(node["arguments"].([]map[string]interface{}), argument)
		if current_token.Type == ":COMMA" {
			consume(":COMMA")
		}
	}

	return node
}

func maybe_binary_op(lhs map[string]interface{}) map[string]interface{} {
	if current_token.Type != ":BINARY_OP" {
		return lhs
	}

	node := map[string]interface{}{"kind": "Binary", "op": BINARY_OPERATIONS[current_token.Value], "lhs": lhs, "location": current_token.Location}

	consume(":BINARY_OP")
	node["rhs"] = parser()

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
	advance()
}
