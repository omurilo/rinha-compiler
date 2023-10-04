package main

import (
	"bytes"
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
				var str bytes.Buffer
				str.WriteString(toString(lhs))
				str.WriteString(toString(rhs))
				return str.String()
			}

			if lhsType == reflect.Int32 && rhsType == reflect.Int32 {
				return lhs.(int32) + rhs.(int32)
			}

			if lhsType == reflect.Int && rhsType == reflect.Int {
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
			return fmt.Sprintf("%v", lhs) == fmt.Sprintf("%v", rhs)
		case Neq:
			return fmt.Sprintf("%v", lhs) != fmt.Sprintf("%v", rhs)
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
		if reflect.TypeOf(value).Kind().String() == "func" {
			fmt.Println("<#closure>")
		} else if _, ok := value.(Tuple); ok {
			fmt.Printf("(%v, %v)\n", toString(value.(Tuple).First), toString(value.(Tuple).Second))
		} else {
			fmt.Println(value)
		}
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

		decode(termData, &callValue)

		var evalArgs []Term

		for _, v := range callValue.Arguments {
			arg := Eval(scope, v)
			evalArgs = append(evalArgs, arg)
		}

		fn_name := callValue.Callee.(map[string]interface{})["text"]

		if _, ok := fn_name.(string); !ok {
			fn_name = "anonymous"
		}

		fn := Eval(scope, callValue.Callee)

		if reflect.TypeOf(fn).Kind().String() != "func" {
			return fn
		}

		result := reflect.ValueOf(fn).Call([]reflect.Value{reflect.ValueOf(evalArgs)})[0].Interface().(Term)

		return result
	case KindFunction:
		var functionValue Function

		decode(termData, &functionValue)

		return func(args []Term) Term {
			if len(args) != len(functionValue.Parameters) {
				Error(functionValue.Location, fmt.Sprintf("Expected %d arguments, but got %d", len(functionValue.Parameters), len(args)))
			}

			isolatedScope := Scope{}
			for k, v := range scope {
				isolatedScope[k] = v
			}

			for i, v := range functionValue.Parameters {
				isolatedScope[v.Text] = args[i]
				isolatedScope[fmt.Sprintf("%s#%v", v.Text, i+1)] = args[i]
			}

			return Eval(isolatedScope, functionValue.Value)
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
	var lhsInt int32
	var rhsInt int32
	var okLhs bool = false
	var okRhs bool = false

	if _, ok := lhs.(int32); ok {
		lhsInt = lhs.(int32)
		okLhs = true
	}

	if _, ok := rhs.(int32); ok {
		rhsInt = rhs.(int32)
		okRhs = true
	}

	if !okLhs || !okRhs {
		Error(loc, fmt.Sprintf("Invalid %s operation", operation))
	}

	return lhsInt, rhsInt
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
	} else if reflect.TypeOf(value).Kind().String() == "func" {
		return "<#closure>"
	} else if reflect.TypeOf(value) == reflect.TypeOf(Tuple{}) {
		return fmt.Sprintf("(%v, %v)", toString(value.(Tuple).First), toString(value.(Tuple).Second))
	} else if reflect.TypeOf(value).Kind() == reflect.Bool {
		return strconv.FormatBool(value.(bool))
	}

	return value.(string)
}

func decode(term Term, value Term) Term {
	err := mapstructure.Decode(term, &value)

	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return value
}
