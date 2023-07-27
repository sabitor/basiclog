package simplelog

import (
	"fmt"
	"io"
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
	t := time.Now()
	l.lineBuf = l.lineBuf[:0] // reset logging line

	// append the prefix to the logging line
	switch logMsg.destination {
	case STDOUT:
		if len(s.stdoutLogger.prefix) > 0 {
			l.lineBuf = append(l.lineBuf, t.Format(s.stdoutLogger.prefix)...)
			l.lineBuf = append(l.lineBuf, ' ')
		}
	case FILE:
		if len(s.fileLogger.prefix) > 0 {
			l.lineBuf = append(l.lineBuf, t.Format(s.fileLogger.prefix)...)
			l.lineBuf = append(l.lineBuf, ' ')
		}
	}

	// append payload to the logging line
	l.lineBuf = append(l.lineBuf, fmt.Sprintln(logMsg.data...)...)
	// write logging line to the log destination
	_, err := l.destination.Write(l.lineBuf)
	if err != nil {
		panic(err)
	}

	return err
}
