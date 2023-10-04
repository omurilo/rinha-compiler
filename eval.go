package main

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

type Scope map[string]Term

func Eval(scope Scope, termData Term) Term {
	kind := termData.(map[string]interface{})["kind"].(string)

	switch TermKind(kind) {
	case KindInt:
		var intValue Int
		decode(termData, &intValue)

		return intValue.Value
	case KindStr:
		var strValue Str
		decode(termData, &strValue)

		return strValue.Value
	case KindBinary:
		var binaryValue Binary

		decode(termData, &binaryValue)

		lhs := Eval(scope, binaryValue.LHS)
		op := BinaryOp(binaryValue.Op)
		rhs := Eval(scope, binaryValue.RHS)
		switch op {
		case Add:
			lhsType := reflect.TypeOf(lhs).Kind()
			rhsType := reflect.TypeOf(rhs).Kind()

			if lhsType == reflect.String || rhsType == reflect.String {
				return toString(lhs) + toString(rhs)
			}

			if lhsType == reflect.Int32 && rhsType == reflect.Int32 {
				return lhs.(int32) + rhs.(int32)
			}

			Error(binaryValue.Location, "invalid add operation")
		case Sub:
			lhsInt, rhsInt := toInt(lhs, rhs, "sub", binaryValue.Location)
			return lhsInt - rhsInt
		case Mul:
			lhsInt, rhsInt := toInt(lhs, rhs, "mul", binaryValue.Location)
			return lhsInt * rhsInt
		case Div:
			lhsInt, rhsInt := toInt(lhs, rhs, "div", binaryValue.Location)
			if rhsInt == 0 {
				Error(binaryValue.Location, "division by zero")
			}
			return lhsInt / rhsInt
		case Rem:
			lhsInt, rhsInt := toInt(lhs, rhs, "rem", binaryValue.Location)
			return lhsInt % rhsInt
		case Eq:
			return isEqual(lhs, rhs, "eq", binaryValue.Location)
		case Neq:
			return !isEqual(lhs, rhs, "neq", binaryValue.Location)
		case And:
			lhsBool, rhsBool := toBool(lhs, rhs)
			return lhsBool && rhsBool
		case Or:
			lhsBool, rhsBool := toBool(lhs, rhs)
			return lhsBool || rhsBool
		case Lt:
			lhsInt, rhsInt := toInt(lhs, rhs, "lt", binaryValue.Location)
			return lhsInt < rhsInt
		case Gt:
			lhsInt, rhsInt := toInt(lhs, rhs, "gt", binaryValue.Location)
			return lhsInt > rhsInt
		case Lte:
			lhsInt, rhsInt := toInt(lhs, rhs, "lte", binaryValue.Location)
			return lhsInt <= rhsInt
		case Gte:
			lhsInt, rhsInt := toInt(lhs, rhs, "gte", binaryValue.Location)
			return lhsInt >= rhsInt
		}
	case KindPrint:
		var printValue Print
		decode(termData, &printValue)

		value := Eval(scope, printValue.Value)
		fmt.Println(toString(value))
		return value
	case KindBool:
		var boolValue Print
		decode(termData, &boolValue)

		return boolValue.Value
	case KindIf:
		var ifValue If
		decode(termData, &ifValue)

		value := Eval(scope, ifValue.Condition)
		boolean, _ := toBool(value, value)
		if boolean {
			return Eval(scope, ifValue.Then)
		} else {
			return Eval(scope, ifValue.Otherwise)
		}
	case KindFirst:
		var firstValue First

		decode(termData, &firstValue)

		value := Eval(scope, firstValue.Value)

		if tuple, ok := value.(Tuple); ok {
			return tuple.First
		}

		Error(firstValue.Location, "Runtime error")
	case KindSecond:
		var secondValue Second

		decode(termData, &secondValue)

		value := Eval(scope, secondValue.Value)

		if tuple, ok := value.(Tuple); ok {
			return tuple.Second
		}

		Error(secondValue.Location, "Runtime error")
	case KindTuple:
		var tupleValue Tuple

		decode(termData, &tupleValue)

		first := Eval(scope, tupleValue.First)
		second := Eval(scope, tupleValue.Second)

		return Tuple{First: first, Second: second}
	case KindCall:
		var callValue Call
		var closure Closure

		decode(termData, &callValue)
		fn := Eval(scope, callValue.Callee)

		decode(fn.(Closure), &closure)

		if len(callValue.Arguments) != len(closure.Value.Parameters) {
			Error(callValue.Location, fmt.Sprintf("Expected %d arguments, but got %d", len(closure.Value.Parameters), len(callValue.Arguments)))
		}

		isolatedScope := Scope{}

		for i, v := range scope {
			isolatedScope[i] = v
		}

		for i, v := range closure.Value.Parameters {
			isolatedScope[v.Text] = Eval(scope, callValue.Arguments[i])
		}

		return Eval(isolatedScope, closure.Value.Body)
	case KindFunction:
		var functionValue Function

		decode(termData, &functionValue)

		return Closure{
			Kind: "Closure",
			Value: ClosureValue{
				Body:       functionValue.Value,
				Parameters: functionValue.Parameters,
			},
		}
	case KindLet:
		var letValue Let

		decode(termData, &letValue)

		scope[letValue.Name.Text] = Eval(scope, letValue.Value)
		return Eval(scope, letValue.Next)
	case KindVar:
		var varValue Var

		decode(termData, &varValue)

		var (
			value Term
			ok    bool
		)
		if value, ok = scope[varValue.Text]; !ok {
			Error(varValue.Location, fmt.Sprintf("undefined variable %s", varValue.Text))
		}

		return value
	}

	return nil
}

func toInt(lhs interface{}, rhs interface{}, operation string, loc Location) (int32, int32) {
	if _, ok := lhs.(int32); !ok {
		Error(loc, fmt.Sprintf("Invalid %s operation", operation))
	}

	if _, ok := rhs.(int32); !ok {
		Error(loc, fmt.Sprintf("Invalid %s operation", operation))
	}

	return lhs.(int32), rhs.(int32)
}

func toBool(lhs interface{}, rhs interface{}) (bool, bool) {
	var okLhs bool = false
	var okRhs bool = false

	if _, ok := lhs.(int32); ok {
		if lhs != 0 {
			okLhs = true
		}
	}

	if _, ok := rhs.(int32); ok {
		if rhs != 0 {
			okRhs = true
		}
	}

	if _, ok := lhs.(string); ok {
		if lhs != "" {
			okLhs = true
		}
	}

	if _, ok := rhs.(string); ok {
		if rhs != "" {
			okRhs = true
		}
	}

	if _, ok := lhs.(bool); ok {
		okLhs = lhs.(bool)
	}

	if _, ok := rhs.(bool); ok {
		okRhs = rhs.(bool)
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
	if reflect.TypeOf(value).Kind() == reflect.Int32 {
		return strconv.Itoa(int(value.(int32)))
	} else if reflect.TypeOf(value) == reflect.TypeOf(Closure{}) && value.(Closure).Kind == "Closure" {
		return "<#closure>"
	} else if reflect.TypeOf(value) == reflect.TypeOf(Tuple{}) {
		return fmt.Sprintf("(%v, %v)", toString(value.(Tuple).First), toString(value.(Tuple).Second))
	} else if reflect.TypeOf(value).Kind() == reflect.Bool {
		return strconv.FormatBool(value.(bool))
	}

	return value.(string)
}

func isEqual(lhs interface{}, rhs interface{}, operation string, loc Location) bool {
	if reflect.TypeOf(lhs) == reflect.TypeOf(rhs) {
		return lhs == rhs
	} else {
		return false
	}
}

func decode(term Term, value Term) Term {
	err := mapstructure.Decode(term, &value)

	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return value
}
