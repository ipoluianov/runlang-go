package lib

import (
	"errors"
	"fmt"
	"strconv"
)

func TypeInt64(args ...interface{}) (result []interface{}, err error) {
	if len(args) != 1 {
		err = errors.New("not enough arguments")
		return
	}
	str := fmt.Sprint(args[0])
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		err = errors.New("cannot convert")
		return
	}
	result = append(result, v)
	return
}
