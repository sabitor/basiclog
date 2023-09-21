// Package simplelog is a logging package with the focus on simplicity,
// ease of use and performance.
// Once started, the simple logger runs as a service and listens for logging
// requests. Logging requests can be send to different log destinations,
// such as standard out, a log file, or both.
// The simple logger can be used simultaneously from multiple goroutines.
package simplelog

import (
	"os"
)

// message catalog
const (
	sg000 = "log service is not running"
	sg001 = "log service was already started"
	sg002 = "log service has not been started"
	sg003 = "unknown log destination specified"
	sg004 = "log file not setup"
)

// SetPrefix sets the prefix for log records.
// If the prefix should also contain actual time data, the Golang reference time placeholders can be used accordingly:
//
//	year: 2006
//	month: 01
//	day: 02
//	hour: 15
//	minute: 04
//	second: 05
//	millisecond: 000000
//
// In addition, to distinguish and parse date and time information, the reference time string has to be
// delimited by # tags and can be used for example as follows: #2006-01-02 15:04:05.000000#.
// Note that not all placeholders have to be used and they can be used in any order.
//
// The destination specifies the name of the log destination where the prefix should be used, e.g. STDOUT or FILE.
// The prefix specifies the prefix for each log record for a given log destination.
func SetPrefix(destination int, prefix ...string) {
	if s.isActive() {
		switch destination {
		case STDOUT:
			s.configService <- configMessage{setprefix, map[int]any{stdoutlogprefix: prefix}}
		case FILE:
			s.configService <- configMessage{setprefix, map[int]any{filelogprefix: prefix}}
		default:
			panic(sg003)
		}
		<-s.configServiceResponse
	} else {
		panic(sg002)
	}
}

// Shutdown stops the log service including post-processing and cleanup.
// Before the log service is stopped, all pending log messages are flushed and resources are released.
// Archiving a log file means that it will be renamed and no new messages will be appended on a new run.
// The archived log file is of the following format: <log file name>_yyyymmddHHMMSS.
// The archivelog flag indicates whether the log file will be archived (true) or not (false).
func Shutdown(archivelog bool) {
	if s.isActive() {
		s.stop(archivelog)
		s.setActive(false)
	} else {
		panic(sg000)
	}
}

// Startup starts the log service.
// The log service runs in its own goroutine.
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
func Startup(bufferSize int) {
	if !s.isActive() {
		s.dataQueue = make(chan logMessage, bufferSize)
		s.configService = make(chan configMessage)
		s.configServiceResponse = make(chan error)
		s.stopService = make(chan bool)
		s.stopServiceResponse = make(chan struct{})
		serviceRunning := make(chan bool)

		go s.run(serviceRunning)
		if !<-serviceRunning {
			panic(sg000)
		} else {
			s.setActive(true)
		}
	} else {
		panic(sg001)
	}
}

// SetupLog opens and initially creates a log file.
// The logName parameter specifies the name of the log file.
// With appendLog it is possible to specify, if a new run of the application first truncates the
// old log before new log entries are written (false) or if new messages are appended to the already
// existing log (true).
func SetupLog(logName string, appendlog bool) {
	if s.isActive() {
		var flag int
		if appendlog {
			flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
		} else {
			flag = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
		}
		s.configService <- configMessage{initlog, map[int]any{logflag: flag, logfilename: logName}}
		if err := <-s.configServiceResponse; err != nil {
			panic(err)
		}
	} else {
		panic(sg002)
	}
}

// SwitchLog closes the current log file and a new log file with the specified name is created and used.
// Thereby, the current log file is not deleted, the new log file must not exist and the log service
// doesn't need to be stopped for this task. The new log file must not exist.
// The newLogName specifies the name of the new log to switch to.
func SwitchLog(newLogName string) {
	if s.isActive() {
		var err error
		flag := os.O_EXCL | os.O_CREATE | os.O_WRONLY
		s.configService <- configMessage{switchlog, map[int]any{logflag: flag, logfilename: newLogName}}
		if err = <-s.configServiceResponse; err != nil {
			panic(err)
		}
	} else {
		panic(sg002)
	}
}

// Write writes a log message to a specified destination.
// The destination parameter specifies the log destination, where the data will be written to.
// The logValues parameter consists of one or multiple values that are logged.
func Write(destination int, values ...any) {
	if s.isActive() {
		switch destination {
		case STDOUT:
			s.dataQueue <- logMessage{STDOUT, values}
		case FILE:
			s.dataQueue <- logMessage{FILE, values}
		case MULTI:
			s.dataQueue <- logMessage{MULTI, values}
		default:
			panic(sg003)
		}
	} else {
		panic(sg002)
	}
}

// ConditionalWrite writes or doesn't write a log message to a specified destination based on a condition.
// The condition parameter enables (true) or disables (false) whether or not a message is written.
// The destination parameter specifies the log destination, where the data will be written to.
// The logValues parameter consists of one or multiple values that are logged.
func ConditionalWrite(condition bool, destination int, values ...any) {
	if s.isActive() {
		if condition {
			switch destination {
			case STDOUT:
				s.dataQueue <- logMessage{STDOUT, values}
			case FILE:
				s.dataQueue <- logMessage{FILE, values}
			case MULTI:
				s.dataQueue <- logMessage{MULTI, values}
			default:
				panic(sg003)
			}
		}
	} else {
		panic(sg002)
	}
}
