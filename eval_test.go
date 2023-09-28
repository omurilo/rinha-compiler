package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestExamples(t *testing.T) {
	files, _ := os.ReadDir("examples")
	for _, file := range files {
		fileName := file.Name()
		t.Log(fileName)
		if !strings.Contains(fileName, ".rinha.json") {
			rinhaToJson(fileName)
			continue
		}
	}

	files, _ = os.ReadDir("examples")
	for _, file := range files {
		fileName := file.Name()
		t.Log(fileName)
		if !strings.Contains(fileName, ".rinha.json") {
			continue
		}

		stdin, err := os.ReadFile("examples/" + fileName)

		if err != nil {
			panic(err)
		}

		var ast File
		if err := json.Unmarshal(stdin, &ast); err != nil {
			fmt.Println(stdin)
			t.Error("Error decoding JSON:", err)
			return
		}
		SCOPE_DEFAULT_SIZE := 8
		scope := make(Scope, SCOPE_DEFAULT_SIZE)
		t.Log(">>>>>>>>>>>> ", fileName)
		Eval(scope, ast.Expression)
		t.Log()
	}
}

func rinhaToJson(fileName string) {
	cmd := exec.Command("rinha", fmt.Sprintf("examples/%s", fileName))
	value, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(fmt.Sprintf("examples/%s.json", fileName))
	if err != nil {
		log.Fatal(err)
	}

	a, err := f.WriteString(string(value))
	if err != nil {
		log.Fatal(a, err)
	}
	f.Sync()
}
