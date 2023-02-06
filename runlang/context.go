package runlang

import (
	"errors"
)

type Context struct {
	returnToLine int
	vars         map[string]interface{}
	stackIfWhile []*Block
	resultPlaces []string
}

func NewContext(returnToLine int) *Context {
	var c Context
	c.returnToLine = returnToLine
	c.vars = make(map[string]interface{})
	return &c
}

func (c *Context) get(name string) (interface{}, error) {
	if len(name) == 0 {
		return nil, errors.New("empty lexem")
	}
	constValue, err := parseConstrant(name)
	if err != nil {
		return nil, err
	}
	if constValue != nil {
		return constValue, nil
	}
	return c.vars[name], nil
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
	pv1, err := c.get(p1)
	if err != nil {
		return false, err
	}
	pv2, err := c.get(p2)
	if err != nil {
		return false, err
	}

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
			return
		case "<=":
			result = pv1double <= pv2double
			return
		case "==":
			result = pv1double == pv2double
			return
		case ">=":
			result = pv1double >= pv2double
			return
		case ">":
			result = pv1double > pv2double
			return
		}
	}

	err = errors.New("wrong contition")

	return
}
