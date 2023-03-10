package runlang

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ipoluianov/runlang-go/lib"
)

type UniFunc func(args ...interface{}) ([]interface{}, error)

type Program struct {
	currentLine     int
	lines           []*Line
	context         *Context
	functions       map[string]int
	stack           []*Context
	parentFunctions map[string]UniFunc
	debugMode       bool
	constants       map[string]interface{}
}

func NewProgram() *Program {
	var c Program
	c.debugMode = false
	c.context = NewContext(-1)
	c.functions = make(map[string]int)
	c.constants = make(map[string]interface{})
	c.parentFunctions = make(map[string]UniFunc)
	c.parentFunctions["run.print"] = lib.Print

	c.parentFunctions["run.string"] = lib.TypeString
	c.parentFunctions["run.int64"] = lib.TypeInt64
	c.parentFunctions["run.double"] = lib.TypeDouble
	return &c
}

func (c *Program) AddFunction(name string, f UniFunc) {
	c.parentFunctions[name] = f
}

func (c *Program) debug(args ...interface{}) {
	if !c.debugMode {
		return
	}
	fmt.Println(args...)
}

func (c *Program) Compile(code string) error {
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

		// Functions definitions
		if len(l.Lexems) > 0 && l.Lexems[0] == "fn" {
			c.functions[l.Lexems[1]] = lineIndex
		}
		if len(l.Lexems) > 0 {
			l0 := l.Lexems[0]
			switch l0 {
			case "return":
				l.Instruction = InstructionReturn
			case "fn":
				l.Instruction = InstructionFn
			case "if":
				l.Instruction = InstructionIf
				if len(l.Lexems) != 5 {
					return errors.New("wrong if")
				}
				l.ConditionVal1 = l.Lexems[1]
				switch l.Lexems[2] {
				case "<":
					l.ConditionOperation = ConditionTypeLess
				case "<=":
					l.ConditionOperation = ConditionTypeLessEq
				case "==":
					l.ConditionOperation = ConditionTypeEq
				case ">=":
					l.ConditionOperation = ConditionTypeMoreEq
				case ">":
					l.ConditionOperation = ConditionTypeMore
				default:
					return errors.New("wrong condition")
				}
				l.ConditionVal2 = l.Lexems[3]
			case "while":
				l.Instruction = InstructionWhile
				if len(l.Lexems) != 5 {
					return errors.New("wrong while")
				}
				l.ConditionVal1 = l.Lexems[1]
				switch l.Lexems[2] {
				case "<":
					l.ConditionOperation = ConditionTypeLess
				case "<=":
					l.ConditionOperation = ConditionTypeLessEq
				case "==":
					l.ConditionOperation = ConditionTypeEq
				case ">=":
					l.ConditionOperation = ConditionTypeMoreEq
				case ">":
					l.ConditionOperation = ConditionTypeMore
				default:
					return errors.New("wrong condition")
				}
				l.ConditionVal2 = l.Lexems[3]
			case "break":
				l.Instruction = InstructionBreak
			case "}":
				l.Instruction = InstructionEnd
			case "dump":
				l.Instruction = InstructionDump
			default:
				l.Instruction = InstructionSet
				ls := l.Lexems
				for i := 0; i < len(ls); i++ {
					if ls[i] == "=" {
						break
					}
					l.LeftPart = append(l.LeftPart, ls[i])
				}
				if len(l.LeftPart) == len(ls) {
					l.LeftPart = nil
					l.RightPart = ls
				} else {
					l.RightPart = ls[len(l.LeftPart)+1:]
				}
				if len(l.RightPart) == 0 {
					return errors.New("no right part on operation")
				}
			}
		}
	}

	for i := range c.lines {
		if len(c.lines[i].RightPart) > 0 {
			c.lines[i].RightPartIsFunction = c.isFunction(c.lines[i].RightPart[0])
		}
	}
	for _, line := range c.lines {
		for _, lex := range line.Lexems {
			v, err := parseConstrant1(lex)
			if err != nil {
				return err
			}
			if v != nil {
				c.constants[lex] = v
			}
		}

		if len(line.LeftPart) == 1 {
			if !line.RightPartIsFunction {
				if len(line.RightPart) == 3 {
					line.SetIsMath = true
				}
			}
		}

		if len(line.LeftPart) == 1 {
			if !line.RightPartIsFunction {
				if len(line.RightPart) == 1 {
					line.SetIsOne = true
				}
			}
		}

	}
	return nil
}

func (c *Program) RunFn(functionName string, args ...interface{}) (result []interface{}, err error) {
	if _, ok := c.functions[functionName]; ok {
		c.currentLine = len(c.lines)
		argsValues := make([]string, 0)
		for i := range args {
			if strValue, ok := args[i].(string); ok {
				argsValues = append(argsValues, fmt.Sprint("\"", strValue, "\""))
			} else {
				argsValues = append(argsValues, fmt.Sprint(args[i]))
			}
		}
		callBody := make([]string, 0)
		callBody = append(callBody, functionName)
		callBody = append(callBody, argsValues...)
		c.fnCall(nil, callBody)
		err = c.run()
		return c.context.lastCallResult, err
	}
	return nil, errors.New("wrong function")
}

func (c *Program) Run() (err error) {
	c.currentLine = 0
	return c.run()
}

func (c *Program) run() (err error) {
	for c.currentLine >= 0 && c.currentLine < len(c.lines) {
		err = c.ExecLine()
		if err != nil {
			err = errors.New("line " + fmt.Sprint(c.currentLine) + ":" + fmt.Sprint(c.lines[c.currentLine].Lexems) + " = " + err.Error())
			return
		}
	}
	return
}

func (c *Program) ExecLine() (err error) {
	switch c.lines[c.currentLine].Instruction {
	case InstructionSet:
		err = c.fnSet()
	case InstructionReturn:
		err = c.fnReturn()
	case InstructionBreak:
		err = c.fnBreak()
	case InstructionDump:
		err = c.fnDump()
	case InstructionEnd:
		err = c.fnEnd(true)
	case InstructionFn:
		err = c.fnFn()
	case InstructionIf:
		err = c.fnIf()
	case InstructionWhile:
		err = c.fnWhile()
	}
	return
}

func (c *Program) fnCall(resultPlaces []string, funcCallBody []string) (err error) {
	functionName := funcCallBody[0]

	if functionLineIndex, internalExists := c.functions[functionName]; internalExists {
		//functionLineIndex := c.functions[functionName]
		ls := c.lines[functionLineIndex].Lexems
		functionLineParameters := ls[2 : len(ls)-1]
		parameters, err := c.parseParameters(funcCallBody[1:])
		if err != nil {
			return err
		}
		ctx := NewContext(c.currentLine + 1)
		ctx.functionName = functionName
		for i := range functionLineParameters {
			ctx.vars[functionLineParameters[i]] = nil
			if i < len(parameters) {
				ctx.vars[functionLineParameters[i]] = parameters[i]
			}
		}
		c.stack = append(c.stack, c.context)
		c.context = ctx
		c.currentLine = functionLineIndex + 1
		block := NewBlock(BlockTypeFn, -1)
		ctx.resultPlaces = resultPlaces
		c.context.stackIfWhile = append(c.context.stackIfWhile, block)
		return nil
	}

	if extFunction, externalExists := c.parentFunctions[functionName]; externalExists {
		parameters, err := c.parseParameters(funcCallBody[1:])
		if err != nil {
			return err
		}
		var resultValues []interface{}
		resultValues, err = extFunction(parameters...)
		if err != nil {
			return err
		}
		// Put results into local variables
		for i := range resultPlaces {
			if i < len(resultValues) {
				c.set(resultPlaces[i], resultValues[i])
			}
		}
		c.currentLine++
		return nil
	}
	err = errors.New("unknown function " + functionName)
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
	c.context.lastCallResult = results
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
	line := c.lines[c.currentLine].Lexems[1:]
	c.context.stackIfWhile = append(c.context.stackIfWhile, NewBlock(BlockTypeIf, c.currentLine+1))
	// if a > b {
	if line[len(line)-1] == "{" {
		line = line[:len(line)-1]
	}
	l := c.lines[c.currentLine]
	cond, err := c.calcCondition(l.ConditionVal1, l.ConditionVal2, l.ConditionOperation)
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
		c.context.stackIfWhile = append(c.context.stackIfWhile, NewBlock(BlockTypeWhile, c.currentLine))
	}

	l := c.lines[c.currentLine]
	cond, err := c.calcCondition(l.ConditionVal1, l.ConditionVal2, l.ConditionOperation)
	if err != nil {
		return err
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
		if last.tp == BlockTypeWhile {
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
		return errors.New("no more instructions")
	}
	el := c.context.stackIfWhile[len(c.context.stackIfWhile)-1]
	if el.tp == BlockTypeWhile {
		c.currentLine = el.beginIndex
		return nil
	}
	if el.tp == BlockTypeFn {
		c.context.stackIfWhile = c.context.stackIfWhile[:len(c.context.stackIfWhile)-1]
		c.exitFromFunction(nil)
		return nil
	}
	if el.tp == BlockTypeIf {
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
	line := c.lines[c.currentLine]
	if line.SetIsMath {
		var result interface{}
		var val1 interface{}
		var val2 interface{}
		val1, err = c.get(line.RightPart[0])
		if err != nil {
			return err
		}
		val2, err = c.get(line.RightPart[2])
		if err != nil {
			return err
		}
		if val1i64, val1i64ok := val1.(int64); val1i64ok {
			if val2i64, val2i64ok := val2.(int64); val2i64ok {
				switch line.RightPart[1] {
				case "+":
					result = val1i64 + val2i64
				case "-":
					result = val1i64 - val2i64
				case "*":
					result = val1i64 * val2i64
				case "/":
					result = val1i64 / val2i64
				default:
					return errors.New("wrong operation")
				}
			}
		}
		if val1f64, val1f64ok := val1.(float64); val1f64ok {
			if val2f64, val2f64ok := val2.(float64); val2f64ok {
				switch line.RightPart[1] {
				case "+":
					result = val1f64 + val2f64
				case "-":
					result = val1f64 - val2f64
				case "*":
					result = val1f64 * val2f64
				case "/":
					result = val1f64 / val2f64
				default:
					return errors.New("wrong operation")
				}
			}
		}
		c.set(line.LeftPart[0], result)
		c.currentLine++
		return
	}

	if line.SetIsOne {
		v, err := c.get(line.RightPart[0])
		if err != nil {
			return err
		}
		c.set(line.LeftPart[0], v)
		c.currentLine++
		return nil
	}

	err = c.fnCall(line.LeftPart, line.RightPart)
	return
}

func (c *Program) set(name string, value interface{}) {
	c.context.vars[name] = value
}

func (c *Program) get(name string) (interface{}, error) {
	if cVal, ok := c.constants[name]; ok {
		return cVal, nil
	}
	v := c.context.vars[name]
	return v, nil
}

func (c *Program) calcCondition(v1 string, v2 string, op ConditionType) (result bool, err error) {
	var val1 interface{}
	var val2 interface{}

	val1, err = c.get(v1)
	if err != nil {
		return false, err
	}
	val2, err = c.get(v2)
	if err != nil {
		return false, err
	}
	if val1i64, val1i64ok := val1.(int64); val1i64ok {
		if val2i64, val2i64ok := val2.(int64); val2i64ok {
			switch op {
			case ConditionTypeLess:
				result = val1i64 < val2i64
				return
			case ConditionTypeLessEq:
				result = val1i64 <= val2i64
				return
			case ConditionTypeEq:
				result = val1i64 == val2i64
				return
			case ConditionTypeMoreEq:
				result = val1i64 >= val2i64
				return
			case ConditionTypeMore:
				result = val1i64 > val2i64
				return
			}
		}
	}
	if val1f64, val1f64ok := val1.(float64); val1f64ok {
		if val2f64, val2f64ok := val2.(float64); val2f64ok {
			switch op {
			case ConditionTypeLess:
				result = val1f64 < val2f64
				return
			case ConditionTypeLessEq:
				result = val1f64 <= val2f64
				return
			case ConditionTypeEq:
				result = val1f64 == val2f64
				return
			case ConditionTypeMoreEq:
				result = val1f64 >= val2f64
				return
			case ConditionTypeMore:
				result = val1f64 > val2f64
				return
			}
		}
	}

	return false, errors.New("wrong condition")
}
