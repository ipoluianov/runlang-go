package runlang

import "strconv"

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
