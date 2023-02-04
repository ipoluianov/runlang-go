package lib

import (
	"errors"
	"math"
)

func Pow(args ...interface{}) (result []interface{}, err error) {
	if len(args) != 2 {
		err = errors.New("wrong arguments")
		return
	}
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
	result = append(result, math.Pow(x, y))
	return
}
