package simplelog

import (
	"fmt"
	"strings"
)

// parseValues parses the variadic function parameters, builds a message from them and returns it.
func parseValues(values []any) string {
	valueList := make([]string, len(values))
	for i, v := range values {
		if s, ok := v.(string); ok {
			// the parameter is already a string; no conversion is required
			valueList[i] = s
		} else {
			// convert parameter into a string
			valueList[i] = fmt.Sprint(v)
		}
	}
	return strings.Join(valueList, " ")
}
