package simplelog

import (
	"fmt"
	"io"
	// "time"
)

// TBD
type logger struct {
	destination io.Writer // log destination, e.g. stdout or bufio.Writer
	lineBuf     []byte    // buffer for one line of log data
}

// newLogger instantiates a new Logger.
// The destination parameter sets the destination to which log data will be written.
func newLogger(destination io.Writer) *logger {
	return &logger{destination: destination}
}

// write writes the output for a logging event.
// Thereby one log event corresponds to one line of output at the used log destination.
// The logValues parameter consists of one or multiple values that are logged.
func (l *logger) write(logValues []any) error {
	l.lineBuf = l.lineBuf[:0]
	l.lineBuf = append(l.lineBuf, fmt.Sprintln(logValues...)...)
	_, err := l.destination.Write(l.lineBuf)
	return err
}
