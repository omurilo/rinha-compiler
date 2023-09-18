package main

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type Scope map[string]Term

func Eval(scope Scope, termData Term) Term {
	kind := termData.(map[string]interface{})["kind"].(string)

	switch TermKind(kind) {
	case KindInt:
		var intValue Int
		decode(termData, &intValue)

		return big.NewInt(intValue.Value)
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
			if lhsInt, ok := lhs.(*big.Int); ok {
				if rhsInt, ok := rhs.(*big.Int); ok {
					return new(big.Int).Add(lhsInt, rhsInt)
				} else if rhsStr, ok := rhs.(string); ok {
					return fmt.Sprintf("%d%s", lhsInt, rhsStr)
				}
			} else if lhsStr, ok := lhs.(string); ok {
				if rhsInt, ok := rhs.(*big.Int); ok {
					return fmt.Sprintf("%s%d", lhsStr, rhsInt)
				} else if rhsStr, ok := rhs.(string); ok {
					return fmt.Sprintf("%s%s", lhsStr, rhsStr)
				}
			}

			Error(binaryValue.Location, "invalid add operation")
		case Sub:
			lhsInt, rhsInt := toInt(lhs, rhs, "sub", binaryValue.Location)
			return new(big.Int).Sub(lhsInt, rhsInt)
		case Mul:
			lhsInt, rhsInt := toInt(lhs, rhs, "mul", binaryValue.Location)
			return new(big.Int).Mul(lhsInt, rhsInt)
		case Div:
			lhsInt, rhsInt := toInt(lhs, rhs, "div", binaryValue.Location)
			if rhsInt.Cmp(big.NewInt(0)) == 0 {
				Error(binaryValue.Location, "division by zero")
			}
			return new(big.Int).Div(lhsInt, rhsInt)
		case Rem:
			lhsInt, rhsInt := toInt(lhs, rhs, "rem", binaryValue.Location)
			return new(big.Int).Rem(lhsInt, rhsInt)
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
			result := lhsInt.Cmp(rhsInt)
			return result < 0
		case Gt:
			lhsInt, rhsInt := toInt(lhs, rhs, "gt", binaryValue.Location)
			result := lhsInt.Cmp(rhsInt)
			return result > 0
		case Lte:
			lhsInt, rhsInt := toInt(lhs, rhs, "lte", binaryValue.Location)
			result := lhsInt.Cmp(rhsInt)
			return result <= 0
		case Gte:
			lhsInt, rhsInt := toInt(lhs, rhs, "gte", binaryValue.Location)
			result := lhsInt.Cmp(rhsInt)
			return result >= 0
		}
	case KindPrint:
		var printValue Print
		decode(termData, &printValue)

		value := Eval(scope, printValue.Value)
		if reflect.TypeOf(value).Kind().String() == "func" {
			fmt.Println("<#closure>")
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
		if bool(value.(bool)) {
			return Eval(scope, ifValue.Then)
		} else {
			return Eval(scope, ifValue.Otherwise)
		}
	case KindFirst:
		var firstValue First
		var firstValueValue Tuple

		decode(termData, &firstValue)
		decode(firstValue.Value, &firstValueValue)

		if firstValueValue.Kind != KindTuple {
			Error(firstValueValue.Location, "Runtime error")
		}
		first := firstValueValue.First
		value := Eval(scope, first)
		return value
	case KindSecond:
		var secondValue Second
		var secondValueValue Tuple

		decode(termData, &secondValue)
		decode(secondValue.Value, &secondValueValue)

		if secondValueValue.Kind != KindTuple {
			Error(secondValueValue.Location, "Runtime error")
		}
		second := secondValueValue.Second
		value := Eval(scope, second)
		return value
	case KindTuple:
		var tupleValue Tuple

		decode(termData, &tupleValue)

		first := Eval(scope, tupleValue.First)
		second := Eval(scope, tupleValue.Second)

		return fmt.Sprintf("(%v, %v)", first, second)
	case KindCall:
		var callValue Call

		decode(termData, &callValue)

		fn := reflect.ValueOf(Eval(scope, callValue.Callee))

		var evalArgs []Term

		for _, v := range callValue.Arguments {
			evalArgs = append(evalArgs, Eval(scope, v))
		}

		return fn.Call([]reflect.Value{reflect.ValueOf(evalArgs), reflect.ValueOf(scope)})[0].Interface().(Term)
	case KindFunction:
		var functionValue Function

		decode(termData, &functionValue)

		return func(args []Term, fScope Scope) Term {
			isolatedScope := Scope{}
			for k, v := range fScope {
				isolatedScope[k] = v
			}
			for i, v := range functionValue.Parameters {
				isolatedScope[v.Text] = args[i]
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

func toInt(lhs interface{}, rhs interface{}, operation string, loc Location) (*big.Int, *big.Int) {
	var lhsInt int64
	var rhsInt int64
	var okLhs bool = false
	var okRhs bool = false

	if _, ok := lhs.(int64); ok {
		lhsInt = lhs.(int64)
		okLhs = true
	}

	if _, ok := rhs.(int64); ok {
		rhsInt = rhs.(int64)
		okRhs = true
	}

	if _, ok := lhs.(*big.Int); ok {
		lhsInt = lhs.(*big.Int).Int64()
		okLhs = true
	}

	if _, ok := rhs.(*big.Int); ok {
		rhsInt = rhs.(*big.Int).Int64()
		okRhs = true
	}

	if !okLhs || !okRhs {
		Error(loc, fmt.Sprintf("Invalid %s operation", operation))
	}

	return big.NewInt(lhsInt), big.NewInt(rhsInt)
}

func toBool(lhs interface{}, rhs interface{}) (bool, bool) {
	var okLhs bool = false
	var okRhs bool = false

	if _, ok := lhs.(int64); ok {
		if lhs != 0 {
			okLhs = true
		}
	}

	if _, ok := rhs.(int64); ok {
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

func decode(term Term, value Term) Term {
	err := mapstructure.Decode(term, &value)

	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return value
}
