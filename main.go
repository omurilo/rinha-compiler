package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/omurilo/rinha-compiler/ast"
	"github.com/omurilo/rinha-compiler/parser"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: rinha-compiler [--json] <file.rinha>")
		os.Exit(1)
	}

	jsonOnly := false
	filename := args[0]
	if args[0] == "--json" {
		jsonOnly = true
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: rinha-compiler --json <file.rinha>")
			os.Exit(1)
		}
		filename = args[1]
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	treeString := parser.Main(string(file), filename)

	if jsonOnly {
		fmt.Println(treeString)
		return
	}

	var tree ast.File
	if err := json.Unmarshal([]byte(treeString), &tree); err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding JSON:", err)
		os.Exit(1)
	}

	scope := make(Scope, 8)
	Eval(scope, tree.Expression)
}
