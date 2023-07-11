// Package simplelog is a logging package with the focus on simplicity and
// ease of use. It utilizes the log package from the standard library with
// some advanced features.
// Once started, the simple logger runs as a service and listens for logging
// requests through the functions WriteTo[Stdout|File|Multiple].
// As the name of the WriteTo functions suggests, the simple logger writes
// to either standard out, a log file, or multiple targets.
package simplelog

// message catalog
const (
	m000 = "control is not running"
	m001 = "log file not initialized"
	m002 = "log service was already started"
	m003 = "log service is not running"
	m004 = "log service has not been started"
	m005 = "log file already initialized"
	m006 = "log file already exists"
)

// Startup starts the log service.
// The log service runs in its own goroutine.
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
func Startup(bufferSize int) {
	if !c.checkState(running) {
		s.setAttribut(logbuffer, bufferSize)
		c.service(start)
	} else {
		panic(m002)
	}
}

// Shutdown stops the log service and does some cleanup.
// Before the log service is stopped, all pending log messages are flushed and resources are released.
// The archive flag indicates whether the log file will be archived (true) or not (false).
// Archiving a log file means that it will be renamed and no new messages will be appended on a new run.
// The archived log file is of the following format: <orig file name>_yymmddHHMMSS.
func Shutdown(archive bool) {
	if c.checkState(running) {
		s.setAttribut(logarchive, archive)
		c.service(stop)
	} else {
		panic(m003)
	}
}

// InitLog initializes the log file.
// The logName specifies the name of the log file.
// The append flag indicates whether messages are appended to the existing log file (true),
// or on a new run whether the old log is removed and a new log is created in its place (false).
func InitLog(logName string, append bool) {
	if s.fileDesc != nil {
		panic(m005)
	}
	if c.checkState(running) {
		s.setAttribut(appendlog, append)
		s.setAttribut(logfilename, logName)
		c.service(initlog)
	} else {
		panic(m004)
	}
}

// SwitchLog closes the current log file and a new log file with the specified name is created and used.
// Thereby, the current log file is not deleted, the new log file must not exist and the log service
// doesn't need to be stopped for this task.
// The newLogName specifies the name of the new log to switch to.
func SwitchLog(newLogName string) {
	if c.checkState(running) {
		s.setAttribut(logfilename, newLogName)
		c.service(switchlog)
	} else {
		panic(m004)
	}
}

// WriteToStdout writes a log message to stdout.
// The logValues parameter consists of a number of different parameters that are logged to stdout.
func WriteToStdout(logValues ...any) {
	if c.checkState(running) {
		msg := parseValues(logValues)
		s.logData <- logMessage{stdout, msg}
	} else {
		panic(m004)
	}
}

// WriteToFile writes a log message to a log file.
// The logValues parameter consists of a number of different parameters that are logged to a log file.
func WriteToFile(logValues ...any) {
	if c.checkState(running) {
		msg := parseValues(logValues)
		s.logData <- logMessage{file, msg}
	} else {
		panic(m004)
	}
}

// WriteToMulti writes a log message to multiple targets.
// The logValues parameter consists of a number of different parameters that are logged to multiple targets,
// here it is stdout and a log file.
func WriteToMulti(logValues ...any) {
	if c.checkState(running) {
		msg := parseValues(logValues)
		s.logData <- logMessage{multi, msg}
	} else {
		panic(m004)
	}
}
