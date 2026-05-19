package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/omurilo/rinha-compiler/ast"
	"github.com/omurilo/rinha-compiler/runtime"
)

type Scope map[string]ast.Term

// TupleVal is the runtime representation of a tuple (holds evaluated values).
type TupleVal struct {
	First  interface{}
	Second interface{}
}

var cache_scope = make(map[string]ast.Term)

func Eval(scope Scope, termData ast.Term) ast.Term {
	kind := termData.(map[string]interface{})["kind"].(string)

	switch ast.TermKind(kind) {
	case ast.KindInt:
		var intValue ast.Int
		decode(termData, &intValue)
		return intValue.Value

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
			runtime.Error(binaryValue.Location, "invalid add operation")
		case ast.Sub:
			lhsInt, rhsInt := toInt(lhs, rhs, "sub", binaryValue.Location)
			return lhsInt - rhsInt
		case ast.Mul:
			lhsInt, rhsInt := toInt(lhs, rhs, "mul", binaryValue.Location)
			return lhsInt * rhsInt
		case ast.Div:
			lhsInt, rhsInt := toInt(lhs, rhs, "div", binaryValue.Location)
			if rhsInt == 0 {
				runtime.Error(binaryValue.Location, "division by zero")
			}
			return lhsInt / rhsInt
		case ast.Rem:
			lhsInt, rhsInt := toInt(lhs, rhs, "rem", binaryValue.Location)
			return lhsInt % rhsInt
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
			return lhsInt < rhsInt
		case ast.Gt:
			lhsInt, rhsInt := toInt(lhs, rhs, "gt", binaryValue.Location)
			return lhsInt > rhsInt
		case ast.Lte:
			lhsInt, rhsInt := toInt(lhs, rhs, "lte", binaryValue.Location)
			return lhsInt <= rhsInt
		case ast.Gte:
			lhsInt, rhsInt := toInt(lhs, rhs, "gte", binaryValue.Location)
			return lhsInt >= rhsInt
		}

	case ast.KindPrint:
		var printValue ast.Print
		decode(termData, &printValue)

		value := Eval(scope, printValue.Value)
		if reflect.TypeOf(value).Kind() == reflect.Func {
			fmt.Println("<#closure>")
		} else if tv, ok := value.(TupleVal); ok {
			fmt.Printf("(%v, %v)\n", toString(tv.First), toString(tv.Second))
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
		boolean, _ := toBool(value, value)
		if boolean {
			return Eval(scope, ifValue.Then)
		}
		if ifValue.Otherwise == nil {
			return nil
		}
		return Eval(scope, ifValue.Otherwise)

	case ast.KindFirst:
		var firstValue ast.First
		decode(termData, &firstValue)

		value := Eval(scope, firstValue.Value)
		if tv, ok := value.(TupleVal); ok {
			return tv.First.(ast.Term)
		}
		runtime.Error(firstValue.Location, "Runtime error: first requires a tuple")

	case ast.KindSecond:
		var secondValue ast.Second
		decode(termData, &secondValue)

		value := Eval(scope, secondValue.Value)
		if tv, ok := value.(TupleVal); ok {
			return tv.Second.(ast.Term)
		}
		runtime.Error(secondValue.Location, "Runtime error: second requires a tuple")

	case ast.KindTuple:
		var tupleValue ast.Tuple
		decode(termData, &tupleValue)

		first := Eval(scope, tupleValue.First)
		second := Eval(scope, tupleValue.Second)
		return TupleVal{First: first, Second: second}

	case ast.KindCall:
		var callValue ast.Call
		decode(termData, &callValue)

		impure := containsPrint(termData)

		var evalArgs []ast.Term
		for _, v := range callValue.Arguments {
			evalArgs = append(evalArgs, Eval(scope, v))
		}

		args_str := argsToString(evalArgs).String()
		fn_name := callValue.Callee.(map[string]interface{})["text"]
		if _, ok := fn_name.(string); !ok {
			fn_name = "anonymous"
		}
		if args_str == "" {
			n, _ := rand.Int(rand.Reader, big.NewInt(1e6))
			args_str = n.String() + fmt.Sprintf("%d", len(evalArgs))
		}

		cacheKey := fmt.Sprintf("%s#%s", fn_name.(string), args_str)
		if cached := cache_scope[cacheKey]; cached != nil {
			return cached
		}

		fn := Eval(scope, callValue.Callee)
		if reflect.TypeOf(fn).Kind() != reflect.Func {
			return fn
		}

		result := reflect.ValueOf(fn).Call([]reflect.Value{reflect.ValueOf(evalArgs)})[0].Interface().(ast.Term)
		if !impure {
			cache_scope[cacheKey] = result
		}
		return result

	case ast.KindFunction:
		var functionValue ast.Function
		decode(termData, &functionValue)

		return func(args []ast.Term) ast.Term {
			if len(args) != len(functionValue.Parameters) {
				runtime.Error(functionValue.Location, fmt.Sprintf("Expected %d arguments, but got %d", len(functionValue.Parameters), len(args)))
			}
			isolatedScope := Scope{}
			for k, v := range scope {
				isolatedScope[k] = v
			}
			for i, v := range functionValue.Parameters {
				isolatedScope[v.Text] = args[i]
			}
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

		value, ok := scope[varValue.Text]
		if !ok {
			runtime.Error(varValue.Location, fmt.Sprintf("undefined variable %s", varValue.Text))
		}
		return value
	}

	return nil
}

func containsPrint(term ast.Term) bool {
	if term == nil {
		return false
	}
	node, ok := term.(map[string]interface{})
	if !ok {
		return false
	}
	if node["kind"] == "Print" {
		return true
	}
	for _, v := range node {
		if child, ok := v.(map[string]interface{}); ok {
			if containsPrint(child) {
				return true
			}
		}
		if children, ok := v.([]interface{}); ok {
			for _, c := range children {
				if containsPrint(c) {
					return true
				}
			}
		}
	}
	return false
}

func toInt(lhs interface{}, rhs interface{}, operation string, loc ast.Location) (int32, int32) {
	var lhsInt, rhsInt int32
	var okLhs, okRhs bool

	if v, ok := lhs.(int32); ok {
		lhsInt = v
		okLhs = true
	}
	if v, ok := rhs.(int32); ok {
		rhsInt = v
		okRhs = true
	}

	if !okLhs || !okRhs {
		runtime.Error(loc, fmt.Sprintf("Invalid %s operation", operation))
	}
	return lhsInt, rhsInt
}

func toBool(lhs interface{}, rhs interface{}) (bool, bool) {
	toB := func(v interface{}) bool {
		switch val := v.(type) {
		case bool:
			return val
		case int32:
			return val != 0
		case string:
			return val != ""
		}
		return false
	}
	return toB(lhs), toB(rhs)
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case int32:
		return strconv.Itoa(int(v))
	case bool:
		return strconv.FormatBool(v)
	case string:
		return v
	case TupleVal:
		return fmt.Sprintf("(%v, %v)", toString(v.First), toString(v.Second))
	}
	if reflect.TypeOf(value).Kind() == reflect.Func {
		return "<#closure>"
	}
	return fmt.Sprintf("%v", value)
}

func argsToString(args []ast.Term) *bytes.Buffer {
	var buf bytes.Buffer
	for _, arg := range args {
		buf.WriteString(toString(arg))
	}
	return &buf
}

func decode(term ast.Term, value ast.Term) ast.Term {
	if err := mapstructure.Decode(term, &value); err != nil {
		return nil
	}
	return value
}
