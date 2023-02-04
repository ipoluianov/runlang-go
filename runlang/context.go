package runlang

import (
	"errors"
	"strconv"
)

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

func (c *Context) get(name string) interface{} {
	iVal, err := strconv.ParseInt(name, 10, 64)
	if err == nil {
		return iVal
	}
	return c.vars[name]
}

func (c *Context) set(name string, value interface{}) {
	c.vars[name] = value
}

func (c *Context) calcCondition(cond []string) (result bool, err error) {
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
