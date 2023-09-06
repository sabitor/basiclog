package simplelog

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// logger represents an object that generates lines of output to an io.Writer.
type logger struct {
	destination io.Writer // log destination, e.g. stdout or bufio.Writer
	lineBuf     []byte    // buffer for one line of log data
}

// newLogger instantiates a new logger.
// The destination parameter sets the destination to which log data will be written.
func newLogger(destination io.Writer) *logger {
	return &logger{destination: destination}
}

// write writes the output for a logging event.
// Thereby one logging event corresponds to one line of output at the used log destination.
func (l *logger) write(logMsg *logMessage) error {
	var prefix []string
	l.lineBuf = l.lineBuf[:0] // reset log record

	switch logMsg.destination {
	case STDOUT:
		prefix = s.stdoutLogger.prefix
	case FILE:
		prefix = s.fileLogger.prefix
	}

	if len(prefix) > 0 {
		// build log prefix
		for _, v := range prefix {
			if strings.HasPrefix(v, dateTimeTag) && strings.HasSuffix(v, dateTimeTag) {
				// date/time placeholders found - replace with real date/time values
				t := time.Now()
				l.lineBuf = append(l.lineBuf, t.Format(strings.Trim(v, dateTimeTag))...)
			} else {
				// no date/time placeholders found
				l.lineBuf = append(l.lineBuf, v...)
			}
			l.lineBuf = append(l.lineBuf, ' ')
		}
	}

	// append payload to the log record
	l.lineBuf = append(l.lineBuf, fmt.Sprintln(logMsg.data...)...)
	// write log record to the log destination
	_, err := l.destination.Write(l.lineBuf)
	if err != nil {
		panic(err)
	}

	return err
}
