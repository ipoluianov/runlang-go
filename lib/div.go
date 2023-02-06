package lib

import (
	"errors"
)

func Div(args ...interface{}) (result []interface{}, err error) {
	if len(args) != 2 {
		err = errors.New("wrong arguments")
		return
	}
	_, isInt64 := args[0].(int64)
	if isInt64 {
		x, xOk := args[0].(int64)
		if !xOk {
			err = errors.New("wrong parameter")
			return
		}
		y, yOk := args[1].(int64)
		if !yOk {
			err = errors.New("wrong parameter")
			return
		}
		result = append(result, x/y)
		return
	}
	_, isDouble := args[0].(float64)
	if isDouble {
		x, xOk := args[0].(float64)
		if !xOk {
			err = errors.New("wrong parameter")
			return
		}
		y, yOk := args[1].(float64)
		if !yOk {
			err = errors.New("wrong parameter")
			return
		}
		result = append(result, x/y)
		return
	}
	return nil, errors.New("wrong type")
}
