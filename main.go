package main

import (
	"fmt"
	"time"

	"github.com/ipoluianov/runlang-go/runlang"

	_ "embed"
)

//go:embed main.run
var code string

func main() {
	p := runlang.NewProgram()
	p.Compile(code)

	fmt.Println("-----------begin")
	dt3 := time.Now()
	res, err := p.RunFn("init", 5, "qwe")
	fmt.Println("Result:", res)
	dt4 := time.Now()
	fmt.Println("-----------end", dt4.Sub(dt3).Milliseconds())

	// 9
	// 3809

	if err != nil {
		fmt.Println("*********************")
		fmt.Println("ERROR:", err.Error())
		fmt.Println("*********************")
	}
}
