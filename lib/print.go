package lib

import "fmt"

func Print(args ...interface{}) (result []interface{}, err error) {
	fmt.Println(args...)
	return
}
