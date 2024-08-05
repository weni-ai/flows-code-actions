package main

import (
	"flag"
	"fmt"
	"strings"
)

type stringMapFlag map[string]string

func (f *stringMapFlag) String() string {
	return "map[string]string"
}

func (f *stringMapFlag) Set(value string) error {
	parts := strings.Split(value, "===")
	if len(parts) != 2 {
		return fmt.Errorf("invalid flag format: %s", value)
	}
	(*f)[parts[0]] = parts[1]
	return nil
}

var QueryParams = map[string]string{}

func main() {
	params := make(stringMapFlag)

	flag.Var(&params, "arg", "Add an argument in the form of key=value")
	flag.Parse()

	for k, v := range params {
		fmt.Printf("%s=%s\n", k, v)
	}
}
