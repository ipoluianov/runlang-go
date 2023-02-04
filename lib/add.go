package lib

import (
	"errors"
)

func Add(args ...interface{}) (result []interface{}, err error) {
	if len(args) != 2 {
		err = errors.New("wrong arguments")
	}
	x, xOk := args[0].(float64)
	if !xOk {
		err = errors.New("wrong parameter")
	}
	y, yOk := args[1].(float64)
	if !yOk {
		err = errors.New("wrong parameter")
	}
	result = append(result, x+y)
	return
}
