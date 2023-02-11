package runlang

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

/*func (c *Context) get(name string, constants map[string]interface{}) (interface{}, error) {
	if cVal, ok := constants[name]; ok {
		return cVal, nil
	}
	return c.vars[name], nil
}*/

/*func (c *Context) set(name string, value interface{}) {
	c.vars[name] = value
}*/
