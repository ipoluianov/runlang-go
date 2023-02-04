package runlang

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Line struct {
	Tp     string
	Lexems []string
}

type Context struct {
	returnToLine int
	vars         map[string]interface{}
	stackIfWhile []*Block
}

func NewContext(returnToLine int) *Context {
	var c Context
	c.returnToLine = returnToLine
	c.vars = make(map[string]interface{})
	return &c
}

type Block struct {
	tp         string
	beginIndex int
}

func NewBlock(tp string, beginIndex int) *Block {
	var c Block
	c.tp = tp
	c.beginIndex = beginIndex
	return &c
}

type Program struct {
	currentLine     int
	lines           []*Line
	context         *Context
	functions       map[string]int
	stack           []*Context
	parentFunctions map[string]func(args ...interface{})
}

func NewProgram() *Program {
	var c Program
	c.context = NewContext(-1)
	c.functions = make(map[string]int)
	c.parentFunctions = make(map[string]func(args ...interface{}))
	c.parentFunctions["print"] = c.funcPrint
	return &c
}

func (c *Program) funcPrint(args ...interface{}) {
	fmt.Println(">", args)
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

func (c *Program) Run() {
	c.currentLine = 0
	for c.currentLine >= 0 && c.currentLine < len(c.lines) {
		c.ExecLine()
	}
}

func (c *Program) ExecLine() {
	time.Sleep(100 * time.Millisecond)
	if len(c.lines[c.currentLine].Lexems) < 1 {
		c.currentLine++
		return
	}
	l0 := c.lines[c.currentLine].Lexems[0]
	fmt.Println("ExecLine", c.currentLine, c.lines[c.currentLine].Lexems)
	_, isFunctionCall := c.functions[l0]
	if isFunctionCall {
		c.fnCall(c.lines[c.currentLine].Lexems)
		return
	}
	_, isExpFunctionCall := c.parentFunctions[l0]
	if isExpFunctionCall {
		c.fnExtCall(c.lines[c.currentLine].Lexems)
		return
	}
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

	c.fnSet()
}

func (c *Program) fnCall(funcCallBody []string) {
	functionName := funcCallBody[0]
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
}

func (c *Program) parseParameters(parts []string) []interface{} {
	parameters := make([]interface{}, len(parts))
	for i := 0; i < len(parameters); i++ {
		parameters[i] = c.get(parts[i])
	}
	return parameters
}

func (c *Program) fnExtCall(funcCallBody []string) {
	functionName := funcCallBody[0]
	parameters := c.parseParameters(funcCallBody[1:])
	c.parentFunctions[functionName](parameters...)
	c.currentLine++
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
	cond, err := c.calcCondition(line)
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

func (c *Program) calcCondition(cond []string) (result bool, err error) {
	if len(cond) != 3 {
		err = errors.New("wrong condition length")
		return
	}

	p1 := cond[0]
	op := cond[1]
	p2 := cond[2]
	pv1 := c.get(p1)
	pv2 := c.get(p2)

	// int64
	pv1int, pv1int_ok := pv1.(int64)
	pv2int, pv2int_ok := pv2.(int64)
	if pv1int_ok && pv2int_ok {
		switch op {
		case "<":
			result = pv1int < pv2int
			return
		case "<=":
			result = pv1int <= pv2int
			return
		case "==":
			result = pv1int == pv2int
			return
		case ">=":
			result = pv1int >= pv2int
			return
		case ">":
			result = pv1int > pv2int
			return
		}
	}

	// double
	pv1double, pv1double_ok := pv1.(float64)
	pv2double, pv2double_ok := pv2.(float64)
	if pv1double_ok && pv2double_ok {
		switch op {
		case "<":
			result = pv1double < pv2double
		case "<=":
			result = pv1double <= pv2double
		case "==":
			result = pv1double == pv2double
		case ">=":
			result = pv1double >= pv2double
		case ">":
			result = pv1double > pv2double
		}
	}

	err = errors.New("wrong contition")

	return
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
	cond, err := c.calcCondition(line)
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

func (c *Program) fnSet() {
	// type of set
	setLine := c.lines[c.currentLine].Lexems
	leftPart := make([]string, 0)
	for i := 0; i < len(setLine); i++ {
		if setLine[i] == "=" {
			break
		}
		leftPart = append(leftPart, setLine[i])
	}
	rigthPart := setLine[len(leftPart)+1:]
	fmt.Println("SET left:", leftPart, "right:", rigthPart)
	if len(rigthPart) == 0 {
		c.currentLine++
		return
	}

	l0 := rigthPart[0]

	_, isFunctionCall := c.functions[l0]
	if isFunctionCall {
		c.fnCall(rigthPart)
		return
	}
	_, isExpFunctionCall := c.parentFunctions[l0]
	if isExpFunctionCall {
		c.fnExtCall(rigthPart)
		return
	}

	if len(rigthPart) == 1 {
		c.set(leftPart[0], c.get(rigthPart[0]))
	}

	if len(rigthPart) == 3 {
		result := int64(0)
		p1 := rigthPart[0]
		op := rigthPart[1]
		p2 := rigthPart[2]
		pv1 := c.get(p1)
		pv2 := c.get(p2)

		// int64
		pv1int, pv1int_ok := pv1.(int64)
		pv2int, pv2int_ok := pv2.(int64)
		if pv1int_ok && pv2int_ok {
			switch op {
			case "+":
				result = pv1int + pv2int
			case "-":
				result = pv1int - pv2int
			case "*":
				result = pv1int * pv2int
			case "/":
				result = pv1int / pv2int
			}
		}
		c.set(leftPart[0], result)
	}

	c.currentLine++
}

func (c *Program) set(name string, value interface{}) {
	c.context.vars[name] = value
}

func (c *Program) get(name string) interface{} {
	iVal, err := strconv.ParseInt(name, 10, 64)
	if err == nil {
		return iVal
	}
	return c.context.vars[name]
}
