package simplelog

import (
	"bufio"
	"os"
)

// misc defines
const (
	dateTimeTag = "<DT>"
)

// log destinations
const (
	STDOUT = 1 << iota     // write the log record to stdout
	FILE                   // write the log record to the log file
	MULTI  = STDOUT | FILE // write the log record to stdout and to the log file
)

// log service states bitmask
const (
	stopped = 1 << iota // the service is stopped and cannot process log requests
	running             // the service is running
)

// log service tasks
const (
	start = iota
	stop
	initlog
	switchlog
	setprefix
)

// log service attributes
const (
	logbuffer       = iota // defines the buffer size of the logMessage channel
	logfilename            // defines the log file name to be used
	logarchive             // a flag which defines whether the log should be archived
	appendlog              // a flag which defines whether the messages are appended to the existing log
	filelogprefix          // defines the prefix that is placed in front of each logging line in the log file
	stdoutlogprefix        // defines the prefix that is placed in front of each logging line in stdout
)

// signal to confirm actions across channels
type signal struct{}

// a logMessage represents the log message which will be sent to the log service.
type logMessage struct {
	destination int // the log destination bits, e.g. stdout, file, and so on.
	data        any // the payload of the log message, which will be sent to the log destination
}

// a configMessage represents the object which will be sent to the log service for configuration purposes.
type configMessage struct {
	task int            // used to trigger certain config tasks by the log service
	data map[int]string // config data used by the config task
}

// stdoutLogger is a data collection to support logging to stdout.
type stdoutLogger struct {
	self   *logger
	prefix string
}

// fileLogger is a data collection to support logging to files.
type fileLogger struct {
	writer *bufio.Writer
	desc   *os.File
	self   *logger
	prefix string
}

// logWriter interface includes definitions of the following method signatures:
//   - instance
type logWriter interface {
	instance() *logger // create and return a *logger instance
}
