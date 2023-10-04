package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
)

const DEFAULT_FILE_PATH = "/var/rinha/source.rinha.json"

func main() {
	var stdin []byte
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	} else {
		args := os.Args
		var file []byte
		var err error

		if len(args) < 2 {
			file, err = os.ReadFile(DEFAULT_FILE_PATH)
		} else {
			file, err = os.ReadFile(os.Args[1])
		}

		if err != nil {
			panic(err)
		}
		stdin = file
	}
	var ast File
	if err := json.Unmarshal(stdin, &ast); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	SCOPE_DEFAULT_SIZE := 8
	scope := make(Scope, SCOPE_DEFAULT_SIZE)
	runtime.GOMAXPROCS(2)
	Eval(scope, ast.Expression)
}
