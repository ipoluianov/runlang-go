package runlang

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ipoluianov/runlang-go/lib"
)

type Program struct {
	currentLine     int
	lines           []*Line
	context         *Context
	functions       map[string]int
	stack           []*Context
	parentFunctions map[string]func(args ...interface{}) ([]interface{}, error)
}

func NewProgram() *Program {
	var c Program
	c.context = NewContext(-1)
	c.functions = make(map[string]int)
	c.parentFunctions = make(map[string]func(args ...interface{}) ([]interface{}, error))
	c.parentFunctions["run.add"] = lib.Add
	c.parentFunctions["run.print"] = lib.Print
	return &c
}

func (c *Program) Compile(code string) {
	lines := strings.FieldsFunc(code, func(r rune) bool {
		return r == 10 || r == 13
	})
	for lineIndex, line := range lines {
		lexems := strings.FieldsFunc(line, func(r rune) bool {
			return r <= 20 || r == ' ' || r == ',' || r == '(' || r == ')'
		})
		var l Line
		l.Lexems = lexems
		c.lines = append(c.lines, &l)
		if len(l.Lexems) > 0 && l.Lexems[0] == "fn" {
			c.functions[l.Lexems[1]] = lineIndex
		}
	}
}

func (c *Program) Run() (err error) {
	c.currentLine = 0
	for c.currentLine >= 0 && c.currentLine < len(c.lines) {
		err = c.ExecLine()
		if err != nil {
			return
		}
	}
	return
}

func (c *Program) ExecLine() (err error) {
	time.Sleep(100 * time.Millisecond)
	if len(c.lines[c.currentLine].Lexems) < 1 {
		c.currentLine++
		return
	}
	l0 := c.lines[c.currentLine].Lexems[0]
	fmt.Println("ExecLine", c.currentLine, c.lines[c.currentLine].Lexems)
	if l0 == "return" {
		c.fnReturn()
		return
	}
	if l0 == "fn" {
		c.fnFn()
		return
	}
	if l0 == "if" {
		c.fnIf()
		return
	}
	if l0 == "while" {
		c.fnWhile()
		return
	}
	if l0 == "break" {
		c.fnBreak()
		return
	}
	if l0 == "}" {
		c.fnEnd()
		return
	}

	err = c.fnSet()
	return
}

func (c *Program) fnCall(funcCallBody []string) (err error) {
	functionName := funcCallBody[0]
	_, internalExists := c.functions[functionName]
	_, externalExists := c.parentFunctions[functionName]
	if !internalExists && !externalExists {
		err = errors.New("unknown function " + functionName)
		return
	}

	if internalExists {
		functionLineIndex := c.functions[functionName]
		functionLineParameters := c.lines[functionLineIndex].Lexems[2:]
		parameters := c.parseParameters(funcCallBody[1:])
		ctx := NewContext(c.currentLine + 1)
		for i := range functionLineParameters {
			ctx.vars[functionLineParameters[i]] = nil
			if i < len(parameters) {
				ctx.vars[functionLineParameters[i]] = parameters[i]
			}
		}
		c.stack = append(c.stack, c.context)
		c.context = ctx
		c.currentLine = functionLineIndex + 1
		c.context.stackIfWhile = append(c.context.stackIfWhile, NewBlock("fn", -1))
	} else {
		if externalExists {
			parameters := c.parseParameters(funcCallBody[1:])
			extFunctinon := c.parentFunctions[functionName]
			extFunctinon(parameters...)
			c.currentLine++
		}
	}
	return
}

func (c *Program) parseParameters(parts []string) []interface{} {
	parameters := make([]interface{}, len(parts))
	for i := 0; i < len(parameters); i++ {
		parameters[i] = c.get(parts[i])
	}
	return parameters
}

func (c *Program) fnReturn() {
	c.exitFromFunction(nil)
}

func (c *Program) exitFromFunction(results []string) {
	c.currentLine = c.context.returnToLine
	c.context = c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *Program) skipBlock() {
	opened := 1
	c.currentLine++
	for c.currentLine < len(c.lines) {
		ls := c.lines[c.currentLine].Lexems
		for i := range ls {
			if ls[i] == "{" {
				opened++
			}
			if opened == 0 {
				break
			}
			if ls[i] == "}" {
				opened--
			}
		}
		if opened == 0 {
			break
		}
		c.currentLine++
	}
	c.currentLine++
}

func (c *Program) fnFn() {
	c.skipBlock()
}

func (c *Program) fnIf() {
	// if a > b {
	c.context.stackIfWhile = append(c.context.stackIfWhile, NewBlock("if", c.currentLine+1))
	line := c.lines[c.currentLine].Lexems[1:]
	if line[len(line)-1] == "{" {
		line = line[:len(line)-1]
	}
	cond, err := c.context.calcCondition(line)
	if err != nil {
		panic(err)
	}
	if cond {
		c.currentLine++
		return
	}
	c.skipBlock()
	/*
	   ls := c.lines[c.currentLine].Lexems

	   	if len(ls) == 3 && ls[0] == "}" && ls[1] == "else" && ls[2] == "{" {
	   		c.currentLine++
	   	}
	*/
}

func (c *Program) fnWhile() {
	last := c.context.stackIfWhile[len(c.context.stackIfWhile)-1]
	if last.beginIndex != c.currentLine {
		c.context.stackIfWhile = append(c.context.stackIfWhile, NewBlock("while", c.currentLine))
	}
	line := c.lines[c.currentLine].Lexems[1:]
	if line[len(line)-1] == "{" {
		line = line[:len(line)-1]
	}
	cond, err := c.context.calcCondition(line)
	if err != nil {
		panic("wrong condition")
	}

	if !cond {
		c.fnBreak()
		return
	}

	c.currentLine++
}

func (c *Program) fnBreak() {
	for len(c.context.stackIfWhile) > 0 {
		last := c.context.stackIfWhile[len(c.context.stackIfWhile)-1]
		c.context.stackIfWhile = c.context.stackIfWhile[:len(c.context.stackIfWhile)-1]
		if last.tp == "while" {
			c.currentLine = last.beginIndex
			c.skipBlock()
			break
		}
	}
}

func (c *Program) fnEnd() {
	if len(c.context.stackIfWhile) == 0 {
		panic("wrong block")
	}
	el := c.context.stackIfWhile[len(c.context.stackIfWhile)-1]
	if el.tp == "while" {
		c.currentLine = el.beginIndex
		return
	}
	if el.tp == "fn" {
		c.context.stackIfWhile = c.context.stackIfWhile[:len(c.context.stackIfWhile)-1]
		c.exitFromFunction(nil)
		return
	}
	if el.tp == "if" {
		c.context.stackIfWhile = c.context.stackIfWhile[:len(c.context.stackIfWhile)-1]
		ls := c.lines[c.currentLine].Lexems
		if len(ls) == 3 && ls[0] == "}" && ls[1] == "else" && ls[2] == "{" {
			c.skipBlock()
		} else {
			c.currentLine++
		}
	}
}

func (c *Program) fnElse() {
	for c.currentLine < len(c.lines) && c.lines[c.currentLine].Lexems[0] != "end" {
		c.currentLine++
	}

}

func (c *Program) fnSet() (err error) {
	// type of set
	setLine := c.lines[c.currentLine].Lexems
	leftPart := make([]string, 0)
	for i := 0; i < len(setLine); i++ {
		if setLine[i] == "=" {
			break
		}
		leftPart = append(leftPart, setLine[i])
	}
	var rightPart []string
	if len(leftPart) == len(setLine) {
		leftPart = nil
		rightPart = setLine
	} else {
		rightPart = setLine[len(leftPart)+1:]
	}

	fmt.Println("SET left:", leftPart, "right:", rightPart)
	if len(rightPart) == 0 {
		err = errors.New("no right part on operation")
		return
	}

	err = c.fnCall(rightPart)

	/*if len(rigthPart) == 1 {
		c.set(leftPart[0], c.get(rigthPart[0]))
	}*/

	c.currentLine++
	return
}

func (c *Program) set(name string, value interface{}) {
	c.context.set(name, value)
}

func (c *Program) get(name string) interface{} {
	return c.context.get(name)
}
