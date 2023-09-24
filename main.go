package main

import (
	// "bufio"
	"encoding/json"
	"fmt"

	// "encoding/json"
	// "fmt"
	// "log"
	"os"

	// "github.com/omurilo/rinha-compiler/ast"
	"github.com/omurilo/rinha-compiler/ast"
	"github.com/omurilo/rinha-compiler/parser"
)

func main() {
	var stdin []byte
	filename := os.Args[1]
	file, err := os.ReadFile(filename)

	if err != nil {
		panic(err)
	}
	stdin = file
	SCOPE_DEFAULT_SIZE := 8
	scope := make(Scope, SCOPE_DEFAULT_SIZE)
	// Eval(scope, ast.Expression)
	treeString := parser.Main(string(stdin), filename)

	var tree ast.File
	if err := json.Unmarshal([]byte(treeString), &tree); err != nil {
		fmt.Println("Error decoding JSON: ", err)
	}

	Eval(scope, tree.Expression)
}
