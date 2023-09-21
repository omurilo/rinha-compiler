package main

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/omurilo/rinha-compiler/ast"
	"github.com/omurilo/rinha-compiler/runtime"
)

type Scope map[string]ast.Term

func Eval(scope Scope, termData ast.Term) ast.Term {
	kind := termData.(map[string]interface{})["kind"].(string)

	// fmt.Println(kind, termData)

	switch ast.TermKind(kind) {
	case ast.KindInt:
		var intValue ast.Int
		decode(termData, &intValue)

		return big.NewInt(intValue.Value)
	case ast.KindStr:
		var strValue ast.Str
		decode(termData, &strValue)

		return strValue.Value
	case ast.KindBinary:
		var binaryValue ast.Binary

		decode(termData, &binaryValue)

		lhs := Eval(scope, binaryValue.LHS)
		op := ast.BinaryOp(binaryValue.Op)
		rhs := Eval(scope, binaryValue.RHS)

		switch op {
		case ast.Add:
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

			runtime.Error(binaryValue.Location, "invalid add operation")
		case ast.Sub:
			lhsInt, rhsInt := toInt(lhs, rhs, "sub", binaryValue.Location)
			return new(big.Int).Sub(lhsInt, rhsInt)
		case ast.Mul:
			lhsInt, rhsInt := toInt(lhs, rhs, "mul", binaryValue.Location)
			return new(big.Int).Mul(lhsInt, rhsInt)
		case ast.Div:
			lhsInt, rhsInt := toInt(lhs, rhs, "div", binaryValue.Location)
			if rhsInt.Cmp(big.NewInt(0)) == 0 {
				runtime.Error(binaryValue.Location, "division by zero")
			}
			return new(big.Int).Div(lhsInt, rhsInt)
		case ast.Rem:
			lhsInt, rhsInt := toInt(lhs, rhs, "rem", binaryValue.Location)
			return new(big.Int).Rem(lhsInt, rhsInt)
		case ast.Eq:
			return fmt.Sprintf("%v", lhs) == fmt.Sprintf("%v", rhs)
		case ast.Neq:
			return fmt.Sprintf("%v", lhs) != fmt.Sprintf("%v", rhs)
		case ast.And:
			lhsBool, rhsBool := toBool(lhs, rhs)
			return lhsBool && rhsBool
		case ast.Or:
			lhsBool, rhsBool := toBool(lhs, rhs)
			return lhsBool || rhsBool
		case ast.Lt:
			lhsInt, rhsInt := toInt(lhs, rhs, "lt", binaryValue.Location)
			result := lhsInt.Cmp(rhsInt)
			return result < 0
		case ast.Gt:
			lhsInt, rhsInt := toInt(lhs, rhs, "gt", binaryValue.Location)
			result := lhsInt.Cmp(rhsInt)
			return result > 0
		case ast.Lte:
			lhsInt, rhsInt := toInt(lhs, rhs, "lte", binaryValue.Location)
			result := lhsInt.Cmp(rhsInt)
			return result <= 0
		case ast.Gte:
			lhsInt, rhsInt := toInt(lhs, rhs, "gte", binaryValue.Location)
			result := lhsInt.Cmp(rhsInt)
			return result >= 0
		}
	case ast.KindPrint:
		var printValue ast.Print
		decode(termData, &printValue)

		value := Eval(scope, printValue.Value)
		if reflect.TypeOf(value).Kind().String() == "func" {
			fmt.Println("<#closure>")
			// fn := reflect.ValueOf(value)
			//
			// var evalArgs []ast.Term
			//
			// return fn.Call([]reflect.Value{reflect.ValueOf(evalArgs), reflect.ValueOf(scope)})[0].Interface().(ast.Term)

		} else {
			fmt.Println(value)
		}
		return value
	case ast.KindBool:
		var boolValue ast.Bool
		decode(termData, &boolValue)

		return boolValue.Value
	case ast.KindIf:
		var ifValue ast.If
		decode(termData, &ifValue)

		value := Eval(scope, ifValue.Condition)
		// fmt.Println("value da condition", value, ifValue.Condition)
		if bool(value.(bool)) {
			return Eval(scope, ifValue.Then)
		} else {
			return Eval(scope, ifValue.Otherwise)
		}
	case ast.KindFirst:
		var firstValue ast.First
		var firstValueValue ast.Tuple

		decode(termData, &firstValue)
		decode(firstValue.Value, &firstValueValue)

		if firstValueValue.Kind != ast.KindTuple {
			runtime.Error(firstValueValue.Location, "Runtime error")
		}
		first := firstValueValue.First
		value := Eval(scope, first)
		return value
	case ast.KindSecond:
		var secondValue ast.Second
		var secondValueValue ast.Tuple

		decode(termData, &secondValue)
		decode(secondValue.Value, &secondValueValue)

		if secondValueValue.Kind != ast.KindTuple {
			runtime.Error(secondValueValue.Location, "Runtime error")
		}
		second := secondValueValue.Second
		value := Eval(scope, second)
		return value
	case ast.KindTuple:
		var tupleValue ast.Tuple

		decode(termData, &tupleValue)

		first := Eval(scope, tupleValue.First)
		second := Eval(scope, tupleValue.Second)

		return fmt.Sprintf("(%v, %v)", first, second)
	case ast.KindCall:
		var callValue ast.Call

		decode(termData, &callValue)
		// fmt.Println("call function", scope, callValue.Arguments)
		fn := reflect.ValueOf(Eval(scope, callValue.Callee))

		var evalArgs []ast.Term

		for _, v := range callValue.Arguments {
			// fmt.Println("call value arguments", scope, v)
			evalArgs = append(evalArgs, Eval(scope, v))
		}

		return fn.Call([]reflect.Value{reflect.ValueOf(evalArgs), reflect.ValueOf(scope)})[0].Interface().(ast.Term)
	case ast.KindFunction:
		var functionValue ast.Function

		decode(termData, &functionValue)

		return func(args []ast.Term, fScope Scope) ast.Term {
			if len(args) != len(functionValue.Parameters) {
				runtime.Error(functionValue.Location, fmt.Sprintf("Expected %d arguments, but got %d", len(functionValue.Parameters), len(args)))
			}
			isolatedScope := Scope{}
			for k, v := range fScope {
				isolatedScope[k] = v
			}
			for i, v := range functionValue.Parameters {
				isolatedScope[v.Text] = args[i]
			}

			// fmt.Println("isolatedScope", isolatedScope, "value", functionValue)

			return Eval(isolatedScope, functionValue.Value)
		}
	case ast.KindLet:
		var letValue ast.Let

		decode(termData, &letValue)

		scope[letValue.Name.Text] = Eval(scope, letValue.Value)
		return Eval(scope, letValue.Next)
	case ast.KindVar:
		var varValue ast.Var

		decode(termData, &varValue)
		// fmt.Println("varValue", varValue.Text, scope)
		var (
			value ast.Term
			ok    bool
		)
		if value, ok = scope[varValue.Text]; !ok {
			runtime.Error(varValue.Location, fmt.Sprintf("undefined variable %s", varValue.Text))
		}
		// fmt.Println("value", value)
		return value
	}

	return nil
}

func toInt(lhs interface{}, rhs interface{}, operation string, loc ast.Location) (*big.Int, *big.Int) {
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
		runtime.Error(loc, fmt.Sprintf("Invalid %s operation", operation))
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

func decode(term ast.Term, value ast.Term) ast.Term {
	err := mapstructure.Decode(term, &value)

	if err != nil {
		// fmt.Println("Error:", err)
		return nil
	}

	return value
}
