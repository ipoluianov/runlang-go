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
	v := 0
	fmt.Println("begin")
	dt1 := time.Now()
	for i := 0; i < 10000000; i++ {
		v += i
	}
	dt2 := time.Now()
	fmt.Println("end", dt2.Sub(dt1).Milliseconds())

	p := runlang.NewProgram()
	p.Compile(code)

	dt3 := time.Now()
	err := p.Run()
	dt4 := time.Now()
	fmt.Println("end", dt4.Sub(dt3).Milliseconds())

	// 9
	// 3809

	if err != nil {
		fmt.Println("*********************")
		fmt.Println("ERROR:", err.Error())
		fmt.Println("*********************")
	}

	fmt.Println("--------------")
}
