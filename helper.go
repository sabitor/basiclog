package simplelog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

// preprocessPrefix processes a specific logging prefix in a way that all time symbol
// placeholders are replaced with the corresponding reference time placeholders.
// If the input prefix does not contain a time symbol placeholder, the input prefix is
// returned unmodified to the caller.
func preprocessPrefix(prefix string) string {
	var newPrefix string
	timeSymbolToReferenceTime := map[string]string{
		"dd":     "02",
		"mm":     "01",
		"yyyy":   "2006",
		"HH":     "15",
		"MI":     "04",
		"SS":     "05",
		"FFFFFF": "000000",
	}

	if strings.Contains(prefix, dateTimeTag) {
		// dateTimeTag found
		if strings.Count(prefix, dateTimeTag)%2 != 0 {
			// no closing dateTimeTag found - input prefix is not going to be processed
			return prefix
		}

		// regexp to filter the input prefix by the dateTimeTag
		pattern := fmt.Sprintf("%s.*?%s", dateTimeTag, dateTimeTag)
		dateTimeFilter := regexp.MustCompile(pattern)
		// regexp to filter the partitions tagged with dateTime to strings consisting of time symbol strings and others
		pattern = "d{2}|m{2}|y{4}|H{2}|MI|S{2}|F{6}|."
		symbolFilter := regexp.MustCompile(pattern)
		startIdxNonDateTimeParts, stopIdxNonDateTimeParts := 0, 0
		partitions := dateTimeFilter.FindAllString(prefix, -1)

		for _, element := range partitions {
			// to support such cases: some_data<DT>time_symbol(s)<DT>some_more_date<DT>time_symbol(s)<DT>
			stopIdxNonDateTimeParts = strings.Index(prefix, element)
			newPrefix += prefix[startIdxNonDateTimeParts:stopIdxNonDateTimeParts]
			startIdxNonDateTimeParts = stopIdxNonDateTimeParts + len(element)
			// remove dateTimeTags from time symbol
			element = strings.Trim(element, dateTimeTag)

			timeSymbol := symbolFilter.FindAllString(element, -1)
			for _, s := range timeSymbol {
				// translate all found time symbols into their corresponding reference time attribute
				if v, ok := timeSymbolToReferenceTime[s]; ok {
					newPrefix += v
				} else {
					newPrefix += s
				}
			}
		}

		// to support such cases: <DT>time_symbol(s)<DT>more_data
		if len(newPrefix) > 0 {
			lastDateTimeTagIdx := strings.LastIndex(prefix, dateTimeTag)
			newPrefix += prefix[lastDateTimeTagIdx+len(dateTimeTag):]
		} else {
			// to support such cases: <DT><DT>some_data
			newPrefix = prefix
		}
	} else {
		// no dateTimeTag found - input prefix is not going to be processed
		newPrefix = prefix
	}

	return newPrefix
}
