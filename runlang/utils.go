package runlang

import (
	"errors"
	"strconv"
	"strings"
)

func parseConstrant(value string) (interface{}, error) {
	if value[0] == '-' || (value[0] >= '0' && value[0] <= '9') {
		if strings.Contains(value, ".") {
			dVal, err := strconv.ParseFloat(value, 64)
			if err == nil {
				return dVal, nil
			} else {
				return 0, err
			}
		} else {
			if len(value) > 1 && value[0] == '0' && value[1] == 'x' {
				value = value[2:]
				iVal, err := strconv.ParseInt(value, 16, 64)
				if err == nil {
					return iVal, nil
				} else {
					return nil, errors.New("wrong hex value")
				}
			} else {
				iVal, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					return iVal, nil
				}
				return 0, err
			}
		}
	}

	if value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1], nil
	}

	return nil, nil
}
