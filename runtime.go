package main

import "fmt"

func Error(loc interface{}, msg string) {
	location := loc.(map[string]interface{})
	panic(fmt.Errorf("%s:%d:%d: %s", location["filename"], location["start"], location["end"], msg))
}
