package main

import (
	"fmt"

	. "github.com/CodingPet-jpg/go-src/flag"
)

func main() {
	Bool("test_bool", false, "bool value")
	Int("test_int", 0, "int value")
	Int64("test_int64", 0, "int64 value")
	Uint("test_uint", 0, "uint value")
	Uint64("test_uint64", 0, "uint64 value")
	String("test_string", "jokoi", "string value")
	Float64("test_float64", 0, "float64 value")
	Duration("test_duration", 0, "time.Duration value")
	//PrintDefaults()
	Parse()
	Usage = func() {
		fmt.Println("Hello World")
	}
	CommandLine.Usage()
}
