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
		}
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
		return value
	}

	return nil
}
