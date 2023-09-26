package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestExamples(t *testing.T) {
	files, _ := os.ReadDir("examples")
	for _, file := range files {
		t.Log(file.Name())
		if strings.Contains(file.Name(), ".rinha") {
			continue
		}

		stdin, err := os.ReadFile("examples/" + file.Name())

		if err != nil {
			panic(err)
		}

		var ast File
		if err := json.Unmarshal(stdin, &ast); err != nil {
			t.Error("Error decoding JSON:", err)
			return
		}
		SCOPE_DEFAULT_SIZE := 8
		scope := make(Scope, SCOPE_DEFAULT_SIZE)
		t.Log(">>>>>>>>>>>> ", file.Name())
		Eval(scope, ast.Expression)
		t.Log()
	}
}
