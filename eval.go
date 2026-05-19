package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

type Scope map[string]Term

var cache_scope map[string]Term = make(map[string]Term, 0)

// closureSeq gives each Closure a unique identity, since Go closures from
// the same literal share the same code pointer and can't be told apart via %p.
var closureSeq uint64

// Closure wraps a function with its impurity flag and a unique ID for caching.
type Closure struct {
	fn     func([]Term) Term
	impure bool
	id     uint64
}

// TailCall is returned by Eval when a call is in tail position.
// The trampoline in KindCall resolves it without growing the Go stack.
type TailCall struct {
	closure Closure
	args    []Term
}

// trampoline drives a tail-call chain to completion.
// It keeps calling closures as long as they return TailCall, then returns
// the final non-TailCall value.
func trampoline(closure Closure, args []Term) Term {
	result := closure.fn(args)
	for {
		tc, ok := result.(TailCall)
		if !ok {
			return result
		}
		result = tc.closure.fn(tc.args)
	}
}

// Eval evaluates a raw AST node.
// tail signals that the current expression is in tail position — a KindCall in
// tail position returns TailCall instead of executing, so the caller's
// trampoline can reuse the current stack frame.
func Eval(scope Scope, termData Term, tail bool) Term {
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

		lhs := Eval(scope, binaryValue.LHS, false)
		op := BinaryOp(binaryValue.Op)
		rhs := Eval(scope, binaryValue.RHS, false)
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
			return equalTerms(lhs, rhs)
		case Neq:
			return !equalTerms(lhs, rhs)
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

		value := Eval(scope, printValue.Value, false)
		if _, ok := value.(Closure); ok {
			fmt.Println("<#closure>")
		} else if _, ok := value.(Tuple); ok {
			fmt.Printf("(%v, %v)\n", toString(value.(Tuple).First), toString(value.(Tuple).Second))
		} else {
			fmt.Println(value)
		}
		return value
	case KindBool:
		var boolValue Bool
		decode(termData, &boolValue)

		return boolValue.Value
	case KindIf:
		var ifValue If
		decode(termData, &ifValue)

		condition := Eval(scope, ifValue.Condition, false)
		boolVal, ok := condition.(bool)
		if !ok {
			Error(ifValue.Location, "if condition must be a boolean")
		}
		if boolVal {
			return Eval(scope, ifValue.Then, tail)
		}
		return Eval(scope, ifValue.Otherwise, tail)
	case KindFirst:
		var firstValue First

		decode(termData, &firstValue)

		value := Eval(scope, firstValue.Value, false)

		if tuple, ok := value.(Tuple); ok {
			return tuple.First
		}

		Error(firstValue.Location, "Runtime error")
	case KindSecond:
		var secondValue Second

		decode(termData, &secondValue)

		value := Eval(scope, secondValue.Value, false)

		if tuple, ok := value.(Tuple); ok {
			return tuple.Second
		}

		Error(secondValue.Location, "Runtime error")
	case KindTuple:
		var tupleValue Tuple

		decode(termData, &tupleValue)

		first := Eval(scope, tupleValue.First, false)
		second := Eval(scope, tupleValue.Second, false)

		return Tuple{First: first, Second: second}
	case KindCall:
		var callValue Call

		decode(termData, &callValue)

		var evalArgs []Term

		for _, v := range callValue.Arguments {
			evalArgs = append(evalArgs, Eval(scope, v, false))
		}

		fn := Eval(scope, callValue.Callee, false)

		closure, ok := fn.(Closure)
		if !ok {
			return fn
		}

		cache_key := fmt.Sprintf("%d#%s", closure.id, argsToString(evalArgs).String())

		// Always check the cache first — a hit is valid regardless of tail position.
		if !closure.impure {
			if cached, hit := cache_scope[cache_key]; hit {
				return cached
			}
		}

		// In tail position: hand the call back to the nearest trampoline instead
		// of adding another stack frame.
		if tail {
			return TailCall{closure: closure, args: evalArgs}
		}

		// Non-tail position: run the trampoline here and cache the final result.
		result := trampoline(closure, evalArgs)

		if !closure.impure {
			cache_scope[cache_key] = result
		}

		return result
	case KindFunction:
		var functionValue Function

		decode(termData, &functionValue)

		impure := containsPrint(functionValue.Value)
		closureSeq++
		id := closureSeq

		fn := func(args []Term) Term {
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

			// The function body is always in tail position.
			return Eval(isolatedScope, functionValue.Value, true)
		}

		return Closure{fn: fn, impure: impure, id: id}
	case KindLet:
		// Iterate over the let-chain instead of recursing so that long
		// sequences of let bindings don't exhaust the goroutine stack.
		for {
			var letValue Let
			decode(termData, &letValue)
			scope[letValue.Name.Text] = Eval(scope, letValue.Value, false)
			termData = letValue.Next
			next, ok := termData.(map[string]interface{})
			if !ok || next["kind"] != string(KindLet) {
				break
			}
		}
		return Eval(scope, termData, tail)
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

// equalTerms compares two runtime values by type and value, never by string representation.
func equalTerms(lhs, rhs interface{}) bool {
	switch l := lhs.(type) {
	case int32:
		r, ok := rhs.(int32)
		return ok && l == r
	case string:
		r, ok := rhs.(string)
		return ok && l == r
	case bool:
		r, ok := rhs.(bool)
		return ok && l == r
	case Tuple:
		r, ok := rhs.(Tuple)
		return ok && equalTerms(l.First, r.First) && equalTerms(l.Second, r.Second)
	default:
		return false
	}
}

// containsPrint recursively checks whether a raw AST node contains a Print node.
func containsPrint(term Term) bool {
	m, ok := term.(map[string]interface{})
	if !ok {
		return false
	}
	if kind, _ := m["kind"].(string); kind == "Print" {
		return true
	}
	for _, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			if containsPrint(child) {
				return true
			}
		case []interface{}:
			for _, elem := range child {
				if childMap, ok := elem.(map[string]interface{}); ok {
					if containsPrint(childMap) {
						return true
					}
				}
			}
		}
	}
	return false
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
	} else if _, ok := value.(Closure); ok {
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

func argsToString(args []Term) *bytes.Buffer {
	var buffer bytes.Buffer
	for i := 0; i < len(args); i++ {
		var value string
		if reflect.TypeOf(args[i]).Kind() == reflect.Int32 {
			value = strconv.Itoa(int(args[i].(int32)))
		} else if c, ok := args[i].(Closure); ok {
			value = fmt.Sprintf("closure%d", c.id)
		} else if reflect.TypeOf(args[i]) == reflect.TypeOf(Tuple{}) {
			value = fmt.Sprintf("(%v, %v)", toString(args[i].(Tuple).First), toString(args[i].(Tuple).Second))
		} else if reflect.TypeOf(args[i]).Kind() == reflect.Bool {
			value = strconv.FormatBool(args[i].(bool))
		} else {
			value = args[i].(string)
		}

		buffer.WriteString(value)
	}

	return &buffer
}
