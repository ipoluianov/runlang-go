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

	err := p.Run()
	if err != nil {
		fmt.Println("*********************")
		fmt.Println("ERROR:", err.Error())
		fmt.Println("*********************")
	}

	fmt.Println("--------------")
}
