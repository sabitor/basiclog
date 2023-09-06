package simplelog

import (
	"bufio"
	"os"
)

// general
const (
	dateTimeTag = "#"
)

// log destinations
const (
	STDOUT = 1 << iota     // write the log record to stdout
	FILE                   // write the log record to the log file
	MULTI  = STDOUT | FILE // write the log record to stdout and to the log file
)

// log service tasks
const (
	initlog = iota
	switchlog
	setprefix
)

// log service attributes
const (
	logbuffer       = iota // defines the buffer size of the logMessage channel
	logfilename            // defines the log file name to be used
	logflag                // a flag or a combination of flags which specifies how to open the log file
	filelogprefix          // defines the prefix that is placed in front of each log line in the log file
	stdoutlogprefix        // defines the prefix that is placed in front of each log line in stdout
)

// a logMessage represents the log message which will be sent to the log service.
type logMessage struct {
	destination int   // the log destination bits, e.g. stdout, file, and so on.
	data        []any // the payload of the log message
}

// a configMessage represents the object which will be sent to the log service for configuration purposes.
type configMessage struct {
	task int         // refers to log service tasks used to trigger certain config tasks
	data map[int]any // config data used by the config task
}

// stdoutLogger is a data collection to support logging to stdout.
type stdoutLogger struct {
	self   *logger
	prefix []string // prefix for each stdout log record
}

// fileLogger is a data collection to support logging to files.
type fileLogger struct {
	writer *bufio.Writer
	desc   *os.File
	self   *logger
	prefix []string // prefix for each file log record
}

// logWriter interface includes definitions of the following method signatures:
//   - instance
type logWriter interface {
	instance() *logger // create and return a *logger instance
}
