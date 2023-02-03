package main

import (
	"fmt"

	"github.com/ipoluianov/runlang-go/runlang"

	_ "embed"
)

//go:embed main.run
var code string

func main() {
	p := runlang.NewProgram()
	p.Compile(code)

	p.Run()

	fmt.Println("123")
}
