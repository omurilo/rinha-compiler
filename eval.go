package main

import (
	"fmt"
)

type Scope map[string]Term

func Eval(scope Scope, termData Term) Term {
	kind := termData.(map[string]interface{})["kind"].(string)

	switch TermKind(kind) {
	case KindStr:
		var strValue Str
		err := mapstructure.Decode(termData, &strValue)

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		return strValue.Value
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
