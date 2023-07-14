package simplelog

import (
	"fmt"
)

// convertToString converts an parameter of type any into a parameter of type string.
func convertToString(value any) string {
	var str string
	var ok bool

	if str, ok = value.(string); !ok {
		// it's not already a string - convert it
		str = fmt.Sprint(value)
	}

	return str
}
