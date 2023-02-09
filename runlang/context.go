package runlang

import (
	"errors"
)

type Context struct {
	returnToLine   int
	functionName   string
	vars           map[string]interface{}
	stackIfWhile   []*Block
	resultPlaces   []string
	lastCallResult []interface{}
}

type ConditionType int

const ConditionTypeLess = ConditionType(0)
const ConditionTypeLessEq = ConditionType(1)
const ConditionTypeEq = ConditionType(2)
const ConditionTypeMoreEq = ConditionType(3)
const ConditionTypeMore = ConditionType(4)

func NewContext(returnToLine int) *Context {
	var c Context
	c.returnToLine = returnToLine
	c.vars = make(map[string]interface{})
	return &c
}

func (c *Context) get(name string, constants map[string]interface{}) (interface{}, error) {
	if cVal, ok := constants[name]; ok {
		return cVal, nil
	}
	return c.vars[name], nil
}

func (c *Context) set(name string, value interface{}) {
	c.vars[name] = value
}

func (c *Context) calcCondition(v1 string, v2 string, op ConditionType, constants map[string]interface{}) (result bool, err error) {
	pv1, err := c.get(v1, constants)
	if err != nil {
		return false, err
	}
	pv2, err := c.get(v2, constants)
	if err != nil {
		return false, err
	}

	// int64
	pv1int, pv1int_ok := pv1.(int64)
	pv2int, pv2int_ok := pv2.(int64)
	if pv1int_ok && pv2int_ok {
		switch op {
		case ConditionTypeLess:
			result = pv1int < pv2int
			return
		case ConditionTypeLessEq:
			result = pv1int <= pv2int
			return
		case ConditionTypeEq:
			result = pv1int == pv2int
			return
		case ConditionTypeMoreEq:
			result = pv1int >= pv2int
			return
		case ConditionTypeMore:
			result = pv1int > pv2int
			return
		}
	}

	// double
	pv1double, pv1double_ok := pv1.(float64)
	pv2double, pv2double_ok := pv2.(float64)
	if pv1double_ok && pv2double_ok {
		switch op {
		case ConditionTypeLess:
			result = pv1double < pv2double
			return
		case ConditionTypeLessEq:
			result = pv1double <= pv2double
			return
		case ConditionTypeEq:
			result = pv1double == pv2double
			return
		case ConditionTypeMoreEq:
			result = pv1double >= pv2double
			return
		case ConditionTypeMore:
			result = pv1double > pv2double
			return
		}
	}

	err = errors.New("wrong contition")

	return
}
