// Package simpleLog is a logging package with the focus on simplicity and
// ease of use. It utilizes the log package from the standard library with
// some advanced features.
// Once started, the simple logger runs as a service and listens for logging
// requests through the functions WriteTo[Stdout|File|Multiple].
// As the name of the WriteTo functions suggests, the simple logger writes
// to either standard out, a log file, or multiple targets.
package simplelog

import (
	"log"
)

// message catalog
const (
	m001 = "log file not initialized"
	m002 = "log service was already started"
	m003 = "log service is not running"
	m004 = "log service has not been started"
)

// Startup starts the log service.
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
// The log service runs in its own goroutine.
func Startup(bufferSize int) {
	s.startup(bufferSize)
}

func (s *service) startup(bufferSize int) {
	if !s.isServiceRunning() {
		// setup log handle map
		s.sLog.logHandle = make(map[int]*log.Logger)

		// setup channels
		s.data = make(chan logMessage, bufferSize)
		s.config = make(chan configMessage)
		s.serviceStop = make(chan signal)
		s.serviceDone = make(chan signal)

		// start the log service
		go s.service()
		for {
			// wait until the service is up
			if s.isServiceRunning() {
				return
			}
		}
	} else {
		panic(m002)
	}
}

// Shutdown stops the log service and does some cleanup.
// Before the log service is stopped, all pending log messages are flushed and resources are released.
func Shutdown() {
	s.shutdown()
}

func (s *service) shutdown() {
	if s.isServiceRunning() {
		// stop the log service
		s.serviceStop <- signal{}
		<-s.serviceDone

		// cleanup
		s.sLog.fileHandle.Close()
	} else {
		panic(m003)
	}
}

// InitLogFile initializes the log file.
func InitLogFile(logName string) {
	s.initLogFile(logName)
}
func (s *service) initLogFile(logName string) {
	if s.isServiceRunning() {
		// initialize the log file
		s.config <- configMessage{initlog, logName}
		<-s.serviceDone
	} else {
		panic(m004)
	}
}

// ChangeLogName changes the log file name.
// As part of this task, the current log file is closed (not deleted) and a log file with the new name is created.
// The log service doesn't need to be stopped for this task.
func ChangeLogName(newLogName string) {
	s.changeLogName(newLogName)
}

func (s *service) changeLogName(newLogName string) {
	if s.isServiceRunning() {
		// change the log name
		s.config <- configMessage{changelog, newLogName}
		<-s.serviceDone
	} else {
		panic(m004)
	}
}

// WriteToStdout writes a log message to stdout.
func WriteToStdout(values ...any) {
	s.writeToStdout(values)
}

func (s *service) writeToStdout(values ...any) {
	if s.isServiceRunning() {
		msg := parseValues(values)
		s.data <- logMessage{stdout, msg}
	} else {
		panic(m004)
	}
}

// WriteToFile writes a log message to a log file.
func WriteToFile(values ...any) {
	s.writeToFile(values)
}

func (s *service) writeToFile(values ...any) {
	if s.isServiceRunning() {
		if s.sLog.fileHandle == nil {
			panic(m001)
		}
		msg := parseValues(values)
		s.data <- logMessage{file, msg}
	} else {
		panic(m004)
	}
}

// WriteToMulti writes a log message to multiple targets.
// Currently supported targets are stdout and file.
func WriteToMulti(values ...any) {
	s.writeToMulti(values)
}

func (s *service) writeToMulti(values ...any) {
	if s.isServiceRunning() {
		if s.sLog.fileHandle == nil {
			panic(m001)
		}
		msg := parseValues(values)
		s.data <- logMessage{multi, msg}
	} else {
		panic(m004)
	}
}
