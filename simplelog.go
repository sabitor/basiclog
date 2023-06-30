// Package simplelog is a logging package with the focus on simplicity and
// ease of use. It utilizes the log package from the standard library with
// some advanced features.
// Once started, the simple logger runs as a service and listens for logging
// requests through the functions WriteTo[Stdout|File|Multiple].
// As the name of the WriteTo functions suggests, the simple logger writes
// to either standard out, a log file, or multiple targets.
package simplelog

import (
	"os"
)

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
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
// The log service runs in its own goroutine.
func Startup(bufferSize int) {
	if !c.checkState(running) {
		// start the log service
		s.setAttribut(logbuffer, bufferSize)
		c.service(start)
	} else {
		panic(m002)
	}
}

// Shutdown stops the log service and does some cleanup.
// Before the log service is stopped, all pending log messages are flushed and resources are released.
func Shutdown() {
	if c.checkState(running) {
		// stop the log service
		c.service(stop)
	} else {
		panic(m003)
	}
}

// InitLogFile initializes the log file.
func InitLogFile(logName string, removeLog bool) {
	if s.fileDesc != nil {
		panic(m005)
	}
	if c.checkState(running) {
		if removeLog {
			// remove log from previous run
			var err error
			if _, err = os.Stat(logName); err == nil {
				if err = os.Remove(logName); err != nil {
					panic(err)
				}
			}
		}
		// initialize the log file
		s.setAttribut(logfilename, logName)
		c.service(initlog)
	} else {
		panic(m004)
	}
}

// NewLogName closes the current log file and a new log file with the specified name is created.
// The current log file is not deleted.
// The new log file must not exist.
// The log service doesn't need to be stopped for this task.
func NewLogName(newLogName string) {
	if c.checkState(running) {
		if _, err := os.Stat(newLogName); err == nil {
			panic(m006)
		}
		// setup a new log file
		s.setAttribut(logfilename, newLogName)
		c.service(newlog)
	} else {
		panic(m004)
	}
}

// WriteToStdout writes a log message to stdout.
func WriteToStdout(values ...any) {
	if c.checkState(running) {
		msg := parseValues(values)
		s.logData <- logMessage{stdout, msg}
	} else {
		panic(m004)
	}
}

// WriteToFile writes a log message to a log file.
func WriteToFile(values ...any) {
	if c.checkState(running) {
		msg := parseValues(values)
		s.logData <- logMessage{file, msg}
	} else {
		panic(m004)
	}
}

// WriteToMulti writes a log message to multiple targets.
func WriteToMulti(values ...any) {
	if c.checkState(running) {
		msg := parseValues(values)
		s.logData <- logMessage{multi, msg}
	} else {
		panic(m004)
	}
}
