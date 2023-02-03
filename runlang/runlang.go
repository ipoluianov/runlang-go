package runlang

import (
	"fmt"
	"strconv"
	"strings"
)

type Line struct {
	Tp     string
	Lexems []string
}

type Context struct {
	returnToLine int
	vars         map[string]interface{}
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
	stackIfWhile    []*Block
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
	if len(c.lines[c.currentLine].Lexems) < 1 {
		c.currentLine++
		return
	}
	l0 := c.lines[c.currentLine].Lexems[0]
	fmt.Println("ExecLine:", c.lines[c.currentLine].Lexems)
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
	if l0 == "end" {
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
	c.currentLine = c.context.returnToLine
	c.context = c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *Program) fnFn() {
	for c.currentLine < len(c.lines) {
		ls := c.lines[c.currentLine].Lexems
		if len(ls) > 0 {
			if ls[0] == "return" {
				break
			}
		}
		c.currentLine++
	}
	c.currentLine++
}

func (c *Program) fnIf() {
	c.stackIfWhile = append(c.stackIfWhile, NewBlock("if", c.currentLine+1))
	cond := c.calcCondition()
	if cond {
		c.currentLine++
		return
	}
	for c.currentLine < len(c.lines) && c.lines[c.currentLine].Lexems[0] != "end" && c.lines[c.currentLine].Lexems[0] != "else" {
		c.currentLine++
	}
	c.currentLine++
}

func (c *Program) calcCondition() bool {
	return true
}

func (c *Program) fnWhile() {
	c.stackIfWhile = append(c.stackIfWhile, NewBlock("while", c.currentLine+1))
}

func (c *Program) fnEnd() {
	el := c.stackIfWhile[len(c.stackIfWhile)-1]
	if el.tp == "while" {
		c.currentLine = el.beginIndex
		return
	}
	c.stackIfWhile = c.stackIfWhile[:len(c.stackIfWhile)-1]
	c.currentLine++
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
		p1, _ := strconv.ParseInt(rigthPart[0], 10, 64)
		op := rigthPart[1]
		p2, _ := strconv.ParseInt(rigthPart[2], 10, 64)

		if op == "+" {
			result = p1 + p2
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
