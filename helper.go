package simplelog

import (
	"fmt"
	"strconv"
)

// convertToString converts an parameter of type any into a string.
func convertToString(value any) string {
	var s string
	var ok bool

	if s, ok = value.(string); !ok {
		// it's not already a string - convert it
		s = fmt.Sprint(value)
	}

	return s
}

// convertToInt converts an parameter of type any into an integer.
func convertToInt(value any) int {
	var i int
	var ok bool

	if i, ok = value.(int); !ok {
		// it's not already an integer - convert it
		i, _ = strconv.Atoi(fmt.Sprint(value))
	}

	return i
}
