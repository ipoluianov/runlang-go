package runlang

import (
	"strconv"
	"strings"
)

func parseConstrant(value string) interface{} {
	if value[0] == '-' || (value[0] >= '0' && value[0] <= '9') {
		if strings.Contains(value, ".") {
			dVal, err := strconv.ParseFloat(value, 64)
			if err == nil {
				return dVal
			}
		} else {
			if value[0] == '0' && value[1] == 'x' {
				value = value[2:]
				iVal, err := strconv.ParseInt(value, 16, 64)
				if err == nil {
					return iVal
				} else {
					return nil
				}
			} else {
				iVal, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					return iVal
				}
			}
		}
	}

	if value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}

	return nil
}
