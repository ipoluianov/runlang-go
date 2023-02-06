package lib

import (
	"errors"
	"fmt"
)

func TypeString(args ...interface{}) (result []interface{}, err error) {
	if len(args) != 1 {
		err = errors.New("not enough arguments")
		return
	}
	str := fmt.Sprint(args[0])
	result = append(result, str)
	return
}
