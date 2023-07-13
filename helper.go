package simplelog

import (
	"fmt"
	// "strings"
)

// convertToString converts an parameter of type any into a parameter of type string.
func convertToString(value any) string {
	var str string
	var ok bool

	if str, ok = value.(string); !ok {
		// convert parameter into a string
		str = fmt.Sprint(value)
	}

	return str
}

// parseValues parses the variadic function parameters, builds a message from them and returns it.
// func parseValues(values []any) string {
// 	valueList := make([]string, len(values))
// 	for i, v := range values {
// 		valueList[i] = convertToString(v)
// 	}
// 	return strings.Join(valueList, " ") // + "\n"
// }
