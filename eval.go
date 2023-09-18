package main

import (
	"fmt"
)

type Scope map[string]Term

func Eval(scope Scope, termData Term) Term {
	kind := termData.(map[string]interface{})["kind"].(string)

	switch TermKind(kind) {
	case KindInt:
		var intValue Int
		err := mapstructure.Decode(termData, &intValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		return big.NewInt(intValue.Value)
	case KindStr:
		var strValue Str
		err := mapstructure.Decode(termData, &strValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		return strValue.Value
	case KindBinary:
		var binaryValue Binary

		err := mapstructure.Decode(termData, &binaryValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		lhs := Eval(scope, binaryValue.LHS)
		op := BinaryOp(binaryValue.Op)
		rhs := Eval(scope, binaryValue.RHS)
		switch op {
		case Add:
			if lhsInt, ok := lhs.(*big.Int); ok {
				if rhsInt, ok := rhs.(*big.Int); ok {
					return new(big.Int).Add(lhsInt, rhsInt)
				}
			} else {
				if rhsInt, ok := rhs.(int64); ok {
					return fmt.Sprintf("%s%d", lhs, rhsInt)
				} else {
					return fmt.Sprintf("%s%s", lhs, rhs)
				}
			}
		case Sub:
			lhsInt, rhsInt := toInt(lhs, rhs, "sub")
			return new(big.Int).Sub(lhsInt, rhsInt)
		case Mul:
			lhsInt, rhsInt := toInt(lhs, rhs, "mul")
			return new(big.Int).Mul(lhsInt, rhsInt)
		case Div:
			lhsInt, rhsInt := toInt(lhs, rhs, "div")
			return new(big.Int).Div(lhsInt, rhsInt)
		case Rem:
			lhsInt, rhsInt := toInt(lhs, rhs, "rem")
			return new(big.Int).Rem(lhsInt, rhsInt)
		case Eq:
			lhsInt, rhsInt := toInt(lhs, rhs, "eq")
			result := lhsInt.Cmp(rhsInt)
			return result == 0
		case Neq:
			lhsInt, rhsInt := toInt(lhs, rhs, "eq")
			result := lhsInt.Cmp(rhsInt)
			return result != 0
		case And:
			lhsBool, rhsBool := toBool(lhs, rhs, "and")
			return lhsBool && rhsBool
		case Or:
			lhsBool, rhsBool := toBool(lhs, rhs, "or")
			return lhsBool || rhsBool
		case Lt:
			lhsInt, rhsInt := toInt(lhs, rhs, "lt")
			// return lhsInt < rhsInt
			result := lhsInt.Cmp(rhsInt)
			return result < 0
		case Gt:
			lhsInt, rhsInt := toInt(lhs, rhs, "gt")
			// return lhsInt > rhsInt
			result := lhsInt.Cmp(rhsInt)
			return result > 0
		case Lte:
			lhsInt, rhsInt := toInt(lhs, rhs, "lte")
			// return lhsInt <= rhsInt
			result := lhsInt.Cmp(rhsInt)
			return result <= 0
		case Gte:
			lhsInt, rhsInt := toInt(lhs, rhs, "lte")
			// return lhsInt >= rhsInt
			result := lhsInt.Cmp(rhsInt)
			return result >= 0
		}
	case KindPrint:
		var printValue Print
		err := mapstructure.Decode(termData, &printValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		value := Eval(scope, printValue.Value)
		if reflect.TypeOf(value).Kind().String() == "func" {
			fmt.Println("<#closure>")
		} else {
			fmt.Println(value)
		}
		return value
	case KindBool:
		var boolValue Print
		err := mapstructure.Decode(termData, &boolValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		return boolValue.Value
	case KindIf:
		var ifValue If
		err := mapstructure.Decode(termData, &ifValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		value := Eval(scope, ifValue.Condition)
		if bool(value.(bool)) {
			return Eval(scope, ifValue.Then)
		} else {
			return Eval(scope, ifValue.Otherwise)
		}
	case KindFirst:
		var firstValue First
		var firstValueValue Tuple

		err := mapstructure.Decode(termData, &firstValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		mapstructure.Decode(firstValue.Value, &firstValueValue)

		if firstValueValue.Kind != KindTuple {
			panic("Runtime error")
		}
		first := firstValueValue.First
		value := Eval(scope, first)
		return value
	case KindSecond:
		var secondValue Second
		var secondValueValue Tuple

		err := mapstructure.Decode(termData, &secondValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		mapstructure.Decode(secondValue.Value, &secondValueValue)

		if secondValueValue.Kind != KindTuple {
			panic("Runtime error")
		}
		second := secondValueValue.Second
		value := Eval(scope, second)
		return value
	case KindTuple:
		var tupleValue Tuple

		err := mapstructure.Decode(termData, &tupleValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		first := Eval(scope, tupleValue.First)
		second := Eval(scope, tupleValue.Second)

		return fmt.Sprintf("(%v, %v)", first, second)
	case KindCall:
		var callValue Call

		err := mapstructure.Decode(termData, &callValue)

		if err != nil {
			fmt.Println("Error: ", err)
			return nil
		}

		fn := reflect.ValueOf(Eval(scope, callValue.Callee))

		var evalArgs []Term

		for _, v := range callValue.Arguments {
			evalArgs = append(evalArgs, Eval(scope, v))
		}

		return fn.Call([]reflect.Value{reflect.ValueOf(evalArgs), reflect.ValueOf(scope)})[0].Interface().(Term)
	case KindFunction:
		var functionValue Function

		err := mapstructure.Decode(termData, &functionValue)

		if err != nil {
			fmt.Println("Error: ", err)
			return nil
		}

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

		err := mapstructure.Decode(termData, &letValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		scope[letValue.Name.Text] = Eval(scope, letValue.Value)
		return Eval(scope, letValue.Next)
	case KindVar:
		var varValue Var

		err := mapstructure.Decode(termData, &varValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		var (
			value Term
			ok    bool
		)
		if value, ok = scope[varValue.Text]; !ok {
			panic(fmt.Sprintf("undefined variable %s", varValue.Text))
		}
		return value
	}

	return nil
}

func toInt(lhs interface{}, rhs interface{}, operation string) (*big.Int, *big.Int) {
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
		message := fmt.Sprintf("Invalid %s operation", operation)
		panic(message)
	}

	return big.NewInt(lhsInt), big.NewInt(rhsInt)
}

func toBool(lhs interface{}, rhs interface{}, operation string) (bool, bool) {
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
