package runlang

import (
	"errors"
	"fmt"
	"strings"

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
	c.parentFunctions["run.sub"] = lib.Sub
	c.parentFunctions["run.mul"] = lib.Mul
	c.parentFunctions["run.div"] = lib.Div

	c.parentFunctions["run.print"] = lib.Print

	c.parentFunctions["run.string"] = lib.TypeString
	c.parentFunctions["run.int64"] = lib.TypeInt64
	c.parentFunctions["run.double"] = lib.TypeDouble
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
	//time.Sleep(10 * time.Millisecond)
	if len(c.lines[c.currentLine].Lexems) < 1 {
		c.currentLine++
		return
	}
	l0 := c.lines[c.currentLine].Lexems[0]
	//fmt.Println("ExecLine", c.currentLine+1, c.lines[c.currentLine].Lexems)
	if l0 == "return" {
		err = c.fnReturn()
		return
	}
	if l0 == "fn" {
		err = c.fnFn()
		return
	}
	if l0 == "if" {
		err = c.fnIf()
		return
	}
	if l0 == "while" {
		err = c.fnWhile()
		return
	}
	if l0 == "break" {
		err = c.fnBreak()
		return
	}
	if l0 == "}" {
		err = c.fnEnd(true)
		return
	}
	if l0 == "dump" {
		err = c.fnDump()
		return
	}

	err = c.fnSet()
	return
}

func (c *Program) fnCall(resultPlaces []string, funcCallBody []string) (err error) {
	functionName := funcCallBody[0]
	_, internalExists := c.functions[functionName]
	_, externalExists := c.parentFunctions[functionName]
	if !internalExists && !externalExists {
		err = errors.New("unknown function " + functionName)
		return
	}

	if internalExists {
		functionLineIndex := c.functions[functionName]
		ls := c.lines[functionLineIndex].Lexems
		functionLineParameters := ls[2 : len(ls)-1]
		parameters, err := c.parseParameters(funcCallBody[1:])
		if err != nil {
			return err
		}
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
		block := NewBlock("fn", -1)
		ctx.resultPlaces = resultPlaces
		c.context.stackIfWhile = append(c.context.stackIfWhile, block)
	} else {
		if externalExists {
			parameters, err := c.parseParameters(funcCallBody[1:])
			if err != nil {
				return err
			}
			extFunctinon := c.parentFunctions[functionName]
			var resultValues []interface{}
			resultValues, err = extFunctinon(parameters...)
			// Put results into local variables
			for i := range resultPlaces {
				if i < len(resultValues) {
					c.set(resultPlaces[i], resultValues[i])
				}
			}
			c.currentLine++
		}
	}
	return
}

func (c *Program) parseParameters(parts []string) ([]interface{}, error) {
	parameters := make([]interface{}, len(parts))
	for i := 0; i < len(parameters); i++ {
		v, err := c.get(parts[i])
		if err != nil {
			return nil, err
		}
		parameters[i] = v
	}
	return parameters, nil
}

func (c *Program) fnReturn() error {
	results := make([]interface{}, 0)
	ls := c.lines[c.currentLine].Lexems
	for i := 1; i < len(ls); i++ {
		v, err := c.get(ls[i])
		if err != nil {
			return err
		}
		results = append(results, v)
	}
	c.exitFromFunction(results)
	return nil
}

func (c *Program) exitFromFunction(results []interface{}) error {
	c.currentLine = c.context.returnToLine
	contextOfFunction := c.context
	c.context = c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
	for i := range contextOfFunction.resultPlaces {
		if i < len(results) {
			c.set(contextOfFunction.resultPlaces[i], results[i])
		}
	}
	return nil
}

func (c *Program) skipBlock() error {
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
	return nil
}

func (c *Program) fnFn() error {
	c.skipBlock()
	return nil
}

func (c *Program) fnIf() error {
	c.context.stackIfWhile = append(c.context.stackIfWhile, NewBlock("if", c.currentLine+1))
	// if a > b {
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
		return nil
	}
	c.skipBlock()
	c.currentLine-- // to end }
	c.fnEnd(false)
	/*
	   ls := c.lines[c.currentLine].Lexems

	   	if len(ls) == 3 && ls[0] == "}" && ls[1] == "else" && ls[2] == "{" {
	   		c.currentLine++
	   	}
	*/
	return nil
}

func (c *Program) fnWhile() error {
	firstExecution := true
	if len(c.context.stackIfWhile) > 0 {
		last := c.context.stackIfWhile[len(c.context.stackIfWhile)-1]
		if last.beginIndex == c.currentLine {
			firstExecution = false
		}
	}

	if firstExecution {
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
		return nil
	}

	c.currentLine++
	return nil
}

func (c *Program) fnBreak() error {
	for len(c.context.stackIfWhile) > 0 {
		last := c.context.stackIfWhile[len(c.context.stackIfWhile)-1]
		c.context.stackIfWhile = c.context.stackIfWhile[:len(c.context.stackIfWhile)-1]
		if last.tp == "while" {
			c.currentLine = last.beginIndex
			c.skipBlock()
			break
		}
	}
	return nil
}

func (c *Program) fnDump() error {
	fmt.Println("--------------------")
	fmt.Println("DUMP:")
	for n, v := range c.context.vars {
		fmt.Println(n, "=", v)
	}
	fmt.Println("--------------------")
	c.currentLine++
	return nil
}

func (c *Program) fnEnd(skipElse bool) error {
	if len(c.context.stackIfWhile) == 0 {
		panic("wrong block")
	}
	el := c.context.stackIfWhile[len(c.context.stackIfWhile)-1]
	if el.tp == "while" {
		c.currentLine = el.beginIndex
		return nil
	}
	if el.tp == "fn" {
		c.context.stackIfWhile = c.context.stackIfWhile[:len(c.context.stackIfWhile)-1]
		c.exitFromFunction(nil)
		return nil
	}
	if el.tp == "if" {
		removeIfStatement := true
		ls := c.lines[c.currentLine].Lexems

		if len(ls) == 3 && ls[0] == "}" && ls[1] == "else" && ls[2] == "{" {
			if skipElse {
				c.skipBlock()
			} else {
				removeIfStatement = false
				c.currentLine++ // in else
			}
		} else {
			c.currentLine++
		}

		if removeIfStatement {
			c.context.stackIfWhile = c.context.stackIfWhile[:len(c.context.stackIfWhile)-1]
		}
	}
	return nil
}

func (c *Program) isFunction(name string) bool {
	_, internalExists := c.functions[name]
	_, externalExists := c.parentFunctions[name]
	return internalExists || externalExists
}

func (c *Program) fnSet() (err error) {
	ls := c.lines[c.currentLine].Lexems
	leftPart := make([]string, 0)
	for i := 0; i < len(ls); i++ {
		if ls[i] == "=" {
			break
		}
		leftPart = append(leftPart, ls[i])
	}
	var rightPart []string
	if len(leftPart) == len(ls) {
		leftPart = nil
		rightPart = ls
	} else {
		rightPart = ls[len(leftPart)+1:]
	}
	if len(rightPart) == 0 {
		err = errors.New("no right part on operation")
		return
	}

	if len(leftPart) == 1 {
		if !c.isFunction(rightPart[0]) {
			if len(rightPart) == 3 {
				parameters := make([]string, 3)
				parameters[1] = rightPart[0]
				parameters[2] = rightPart[2]
				switch rightPart[1] {
				case "+":
					parameters[0] = "run.add"
				case "-":
					parameters[0] = "run.sub"
				case "*":
					parameters[0] = "run.mul"
				case "/":
					parameters[0] = "run.div"
				}
				if len(parameters[0]) > 0 {
					c.fnCall(leftPart, parameters)
					return
				} else {
					err = errors.New("wrong operation")
					return
				}
			}

			if len(rightPart) == 1 {
				v, err := c.get(rightPart[0])
				if err != nil {
					return err
				}
				c.set(leftPart[0], v)
				c.currentLine++
				return nil
			}
		}
	}

	err = c.fnCall(leftPart, rightPart)
	return
}

func (c *Program) set(name string, value interface{}) {
	c.context.set(name, value)
}

func (c *Program) get(name string) (interface{}, error) {
	return c.context.get(name)
}
