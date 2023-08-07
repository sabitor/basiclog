// Package simplelog is a logging package with the focus on simplicity,
// ease of use and performance.
// Once started, the simple logger runs as a service and listens for logging
// requests. Logging requests can be send to different logging destinations,
// such as standard out, a log file, or both.
// The simple logger can be used simultaneously from multiple goroutines.
package simplelog

import (
	"os"
)

// message catalog
const (
	m000 = "log service is not running"
	m001 = "log service was already started"
	m002 = "log service has not been started"
	m003 = "unknown log destination specified"
)

// SetPrefix sets the prefix for logging lines.
// If the prefix should also contain actual date and time data, the following placeholders
// can be applied for given data:
//
//	year: yyyy
//	month: mm
//	day: dd
//	hour: HH
//	minute: MI
//	second: SS
//	millisecond: FFFFFF
//
// In addition, to distinguish and parse date and time information, placeholders have to be delimited by
// <DT>...<DT> tags and can be used for example as follows: <DT>yyyy-mm-dd HH:MI:SS.FFFFFF<DT>.
// Note that not all placeholders have to be used, they can be used in any order and even
// non-datetime characters or strings can be integrated.
//
// The destination specifies the name of the log destination where the prefix should be used, e.g. STDOUT or FILE.
// The prefix specifies the prefix for each logging line for a given log destination.
func SetPrefix(destination int, prefix string) {
	if s.isActive() {
		preparedPrefix := preprocessPrefix(prefix)
		switch destination {
		case STDOUT:
			s.configService <- configMessage{setprefix, map[int]any{stdoutlogprefix: preparedPrefix}}
		case FILE:
			s.configService <- configMessage{setprefix, map[int]any{filelogprefix: preparedPrefix}}
		default:
			panic(m003)
		}
		<-s.configServiceResponse
	} else {
		panic(m002)
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
		panic(m000)
	}
}

// Startup starts the log service.
// The log service runs in its own goroutine.
// The logName parameter specifies the name of the log file.
// With appendLog it is possible to specify, if a new run of the application first truncates the
// old log before new log entries are written (false) or if new messages are appended to the already
// existing log (true).
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
func Startup(logName string, appendlog bool, bufferSize int) {
	if !s.isActive() {
		s.dataQueue = make(chan logMessage, bufferSize)
		s.configService = make(chan configMessage)
		s.configServiceResponse = make(chan error)
		s.stopService = make(chan bool)
		s.stopServiceResponse = make(chan struct{})
		serviceRunning := make(chan bool)

		go s.run(serviceRunning)
		if !<-serviceRunning {
			panic(m000)
		} else {
			s.setActive(true)
		}

		// initialize log file
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
		panic(m001)
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
		panic(m002)
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
			panic(m003)
		}
	} else {
		panic(m002)
	}
}
