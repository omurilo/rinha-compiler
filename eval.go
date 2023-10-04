package main

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

type Scope map[string]Term

func Eval(scope Scope, termData Term) Term {
	kind := termData.(map[string]interface{})["kind"].(string)

	switch TermKind(kind) {
	case KindInt:
		return Int{Value: int32(termData.(map[string]interface{})["value"].(float64))}
	case KindStr:
		return Str{Value: termData.(map[string]interface{})["value"].(string)}
	case KindBinary:
		var binaryValue = termData.(map[string]interface{})

		lhs := Eval(scope, binaryValue["lhs"])
		op := BinaryOp(binaryValue["op"].(string))
		rhs := Eval(scope, binaryValue["rhs"])
		switch op {
		case Add:
			if _, ok := lhs.(Str); ok {
				return Str{Value: toString(lhs) + toString(rhs)}
			}

			if _, ok := rhs.(Str); ok {
				return Str{Value: toString(lhs) + toString(rhs)}
			}

			if lhs, ok := lhs.(Int); ok {
				if rhs, ok := rhs.(Int); ok {
					return Int{Value: lhs.Value + rhs.Value}
				}
			}

			Error(binaryValue["location"], "invalid add operation")
		case Sub:
			lhsInt, rhsInt := toInt(lhs, rhs, "sub", binaryValue["location"])
			return Int{Value: lhsInt - rhsInt}
		case Mul:
			lhsInt, rhsInt := toInt(lhs, rhs, "mul", binaryValue["location"])
			return Int{Value: lhsInt * rhsInt}
		case Div:
			lhsInt, rhsInt := toInt(lhs, rhs, "div", binaryValue["location"])
			if rhsInt == 0 {
				Error(binaryValue["location"], "division by zero")
			}
			return Int{Value: lhsInt / rhsInt}
		case Rem:
			lhsInt, rhsInt := toInt(lhs, rhs, "rem", binaryValue["location"])
			return Int{Value: lhsInt % rhsInt}
		case Eq:
			return Bool{Value: isEqual(lhs, rhs, "eq", binaryValue["location"])}
		case Neq:
			return Bool{Value: !isEqual(lhs, rhs, "neq", binaryValue["location"])}
		case And:
			lhsBool, rhsBool := toBool(lhs, rhs)
			return Bool{Value: lhsBool && rhsBool}
		case Or:
			lhsBool, rhsBool := toBool(lhs, rhs)
			return Bool{Value: lhsBool || rhsBool}
		case Lt:
			lhsInt, rhsInt := toInt(lhs, rhs, "lt", binaryValue["location"])
			return Bool{Value: lhsInt < rhsInt}
		case Gt:
			lhsInt, rhsInt := toInt(lhs, rhs, "gt", binaryValue["location"])
			return Bool{Value: lhsInt > rhsInt}
		case Lte:
			lhsInt, rhsInt := toInt(lhs, rhs, "lte", binaryValue["location"])
			return Bool{Value: lhsInt <= rhsInt}
		case Gte:
			lhsInt, rhsInt := toInt(lhs, rhs, "gte", binaryValue["location"])
			return Bool{Value: lhsInt >= rhsInt}
		}
	case KindPrint:
		value := Eval(scope, termData.(map[string]interface{})["value"])
		fmt.Println(toString(value))
		return value
	case KindBool:
		return Bool{Value: termData.(map[string]interface{})["value"].(bool)}
	case KindIf:
		value := Eval(scope, termData.(map[string]interface{})["condition"])
		boolean, _ := toBool(value, value)
		if boolean {
			return Eval(scope, termData.(map[string]interface{})["then"])
		} else {
			return Eval(scope, termData.(map[string]interface{})["otherwise"])
		}
	case KindFirst:
		value := Eval(scope, termData.(map[string]interface{})["value"])

		if tuple, ok := value.(Tuple); ok {
			return tuple.First
		}

		Error(termData.(map[string]interface{})["location"], "Runtime error")
	case KindSecond:
		value := Eval(scope, termData.(map[string]interface{})["value"])

		if tuple, ok := value.(Tuple); ok {
			return tuple.Second
		}

		Error(termData.(map[string]interface{})["location"], "Runtime error")
	case KindTuple:
		first := Eval(scope, termData.(map[string]interface{})["first"])
		second := Eval(scope, termData.(map[string]interface{})["second"])

		return Tuple{First: first, Second: second}
	case KindCall:
		callValue := termData.(map[string]interface{})
		fn := Eval(scope, callValue["callee"])

		closure := fn.(Closure)

		if len(callValue["arguments"].([]interface{})) != len(closure.Value.Parameters.([]interface{})) {
			Error(callValue["location"], fmt.Sprintf("Expected %d arguments, but got %d", len(closure.Value.Parameters.([]interface{})), len(callValue["arguments"].([]interface{}))))
		}

		isolatedScope := make(Scope, len(scope))

		for i, c := range scope {
			isolatedScope[i] = c
		}

		for i, v := range closure.Value.Parameters.([]interface{}) {
			isolatedScope[v.(map[string]interface{})["text"].(string)] = Eval(scope, callValue["arguments"].([]interface{})[i])
		}

		return Eval(isolatedScope, closure.Value.Body)
	case KindFunction:
		functionValue := termData.(map[string]interface{})

		return Closure{
			Kind: "Closure",
			Value: ClosureValue{
				Body:       functionValue["value"],
				Parameters: functionValue["parameters"],
			},
		}
	case KindLet:
		scope[termData.(map[string]interface{})["name"].(map[string]interface{})["text"].(string)] = Eval(scope, termData.(map[string]interface{})["value"])
		return Eval(scope, termData.(map[string]interface{})["next"])
	case KindVar:
		var (
			value Term
			ok    bool
		)
		if value, ok = scope[termData.(map[string]interface{})["text"].(string)]; !ok {
			Error(termData.(map[string]interface{})["location"], fmt.Sprintf("undefined variable %s", termData.(map[string]interface{})["text"].(string)))
		}

		return value
	}

	return nil
}

func toInt(lhs interface{}, rhs interface{}, operation string, loc interface{}) (int32, int32) {
	if _, ok := lhs.(Int); !ok {
		Error(loc, fmt.Sprintf("Invalid %s operation", operation))
	}

	if _, ok := rhs.(Int); !ok {
		Error(loc, fmt.Sprintf("Invalid %s operation", operation))
	}

	return lhs.(Int).Value, rhs.(Int).Value
}

func toBool(lhs interface{}, rhs interface{}) (bool, bool) {
	var okLhs bool = false
	var okRhs bool = false

	if _, ok := lhs.(Int); ok {
		if lhs.(Int).Value != 0 {
			okLhs = true
		}
	}

	if _, ok := rhs.(Int); ok {
		if rhs.(Int).Value != 0 {
			okRhs = true
		}
	}

	if _, ok := lhs.(Str); ok {
		if lhs.(Str).Value != "" {
			okLhs = true
		}
	}

	if _, ok := rhs.(Str); ok {
		if rhs.(Str).Value != "" {
			okRhs = true
		}
	}

	if _, ok := lhs.(Bool); ok {
		okLhs = lhs.(Bool).Value
	}

	if _, ok := rhs.(Bool); ok {
		okRhs = rhs.(Bool).Value
	}

	if lhs == nil {
		okLhs = false
	}

	if rhs == nil {
		okRhs = false
	}

	return okLhs, okRhs
}

func toString(value interface{}) string {
	if _, ok := value.(Int); ok {
		return strconv.Itoa(int(value.(Int).Value))
	} else if _, ok := value.(Closure); ok {
		return "<#closure>"
	} else if _, ok := value.(Tuple); ok {
		return fmt.Sprintf("(%v, %v)", toString(value.(Tuple).First), toString(value.(Tuple).Second))
	} else if _, ok := value.(Bool); ok {
		return strconv.FormatBool(value.(Bool).Value)
	}

	return value.(Str).Value
}

func isEqual(lhs interface{}, rhs interface{}, operation string, loc interface{}) bool {
	if _, ok := lhs.(Int); ok {
		if _, ok := rhs.(Int); ok {
			return lhs.(Int).Value == rhs.(Int).Value
		}
	} else if _, ok := lhs.(Str); ok {
		if _, ok := rhs.(Str); ok {
			return lhs.(Str).Value == rhs.(Str).Value
		}
	}

	return false
}

func decode(term Term, value Term) Term {
	err := mapstructure.Decode(term, &value)

	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return value
}
