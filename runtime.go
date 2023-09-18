package main

import "fmt"

func Error(loc Location, msg string) {
	panic(fmt.Errorf("%s:%d:%d: %s", loc.Filename, loc.Start, loc.End, msg))
}
