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
		return value
	}

	return nil
}
